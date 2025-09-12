package handlers

import (
	"bytes"
	"fmt"
	"log/slog"
	"net/http"
	"reflect"
	"strings"
	"time"

	goAMF3 "github.com/breign/goAMF3"
	"github.com/mk6i/retro-aim-server/state"
	"github.com/speps/go-amf"
)

// AMFVersion represents the AMF encoding version
type AMFVersion int

const (
	AMF0 AMFVersion = 0
	AMF3 AMFVersion = 3
)

// AMFEncoder handles AMF encoding operations for WebAPI responses
type AMFEncoder struct {
	logger *slog.Logger
}

// NewAMFEncoder creates a new AMF encoder instance
func NewAMFEncoder(logger *slog.Logger) *AMFEncoder {
	return &AMFEncoder{logger: logger}
}

// EncodeAMF encodes data to AMF format based on the specified version
func (e *AMFEncoder) EncodeAMF(data interface{}, version AMFVersion) ([]byte, error) {
	switch version {
	case AMF0:
		// Convert to AMF0-compatible format and encode
		amfData := e.toAMFCompatibleWithVersion(data, version)
		var buf bytes.Buffer
		_, err := amf.EncodeAMF0(&buf, amfData)
		if err != nil {
			return nil, fmt.Errorf("AMF0 encoding failed: %w", err)
		}
		return buf.Bytes(), nil
	case AMF3:
		// For AMF3, use goAMF3 which properly supports it
		// Convert to a regular map structure (no ECMAArray needed)
		amfData := e.toAMF3Compatible(data)
		// goAMF3 panics on nil values, ensure we sanitize
		sanitized := e.sanitizeForAMF3(amfData)
		encoded := goAMF3.EncodeAMF3(sanitized)
		return encoded, nil
	default:
		return nil, fmt.Errorf("unsupported AMF version: %d", version)
	}
}

// toAMF3Compatible converts Go types to AMF3-compatible format for goAMF3
func (e *AMFEncoder) toAMF3Compatible(data interface{}) interface{} {
	if data == nil {
		return map[string]interface{}{}
	}

	// goAMF3 handles regular Go types well, just need to ensure maps are used
	// Don't use ECMAArray for AMF3 - just regular maps
	switch d := data.(type) {
	case BaseResponse:
		return e.baseResponseToMapWithVersion(d, AMF3)
	case ResponseBody:
		return e.responseBodyToMapWithVersion(d, AMF3)
	case ErrorResponse:
		return e.errorResponseToMapWithVersion(d, AMF3)
	case StartSessionResponse:
		// Special handling for StartSessionResponse
		return map[string]interface{}{
			"response": map[string]interface{}{
				"statusCode": d.Response.StatusCode,
				"statusText": d.Response.StatusText,
				"data": map[string]interface{}{
					"aimsid":          d.Response.Data.AimSID,
					"fetchTimeout":    d.Response.Data.FetchTimeout,
					"timeToNextFetch": d.Response.Data.TimeToNextFetch,
					"fetchBaseURL":    d.Response.Data.FetchBaseURL, // Required for Gromit
					"events":          d.Response.Data.Events,
					"wellKnownUrls":   d.Response.Data.WellKnownUrls,
				},
			},
		}
	case FetchEventsResponse:
		// Special handling for FetchEventsResponse
		// goAMF3 can't handle uint64, must convert to int
		return map[string]interface{}{
			"response": map[string]interface{}{
				"statusCode": d.Response.StatusCode,
				"statusText": d.Response.StatusText,
				"data": map[string]interface{}{
					"events":          d.Response.Data.Events,
					"lastSeqNum":      int(d.Response.Data.LastSeqNum), // Convert uint64 to int
					"timeToNextFetch": d.Response.Data.TimeToNextFetch,
					"fetchBaseURL":    d.Response.Data.FetchBaseURL,
				},
			},
		}
	case EndSessionResponse:
		// Special handling for EndSessionResponse - Gromit expects flat structure
		// Based on Gromit's MockServer, it expects:
		// { "data": {}, "statusCode": 200, "statusText": "OK" }
		return map[string]interface{}{
			"data":       map[string]interface{}{}, // Empty data object
			"statusCode": d.Response.StatusCode,
			"statusText": d.Response.StatusText,
		}
	default:
		// For other types, convert structs to maps
		return e.convertToMap(data)
	}
}

// sanitizeForAMF3 recursively removes nil values from the data structure
// because goAMF3 panics when encountering nil values in maps
func (e *AMFEncoder) sanitizeForAMF3(data interface{}) interface{} {
	if data == nil {
		return map[string]interface{}{}
	}

	switch v := data.(type) {
	case uint64:
		// goAMF3 can't handle uint64, convert to int
		return int(v)
	case uint32:
		// Convert all unsigned to signed for safety
		return int(v)
	case uint16:
		return int(v)
	case uint8:
		return int(v)
	case uint:
		return int(v)
	case map[string]interface{}:
		result := make(map[string]interface{})
		for key, val := range v {
			if val == nil {
				// For fields like 'data', replace with empty map
				// For other fields, skip them
				if key == "data" {
					result[key] = map[string]interface{}{}
				}
				continue
			}
			result[key] = e.sanitizeForAMF3(val)
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = e.sanitizeForAMF3(item)
		}
		return result
	case []state.WebAPIEvent:
		// Handle WebAPIEvent arrays specially
		result := make([]interface{}, len(v))
		for i, event := range v {
			// AMF3 has a 29-bit limit for integers
			// Keep seqNum small by using modulo
			seqNum := int(event.SeqNum % (1 << 29))
			// Convert timestamp to seconds ago to keep it small
			timestampSec := int(time.Now().Unix() - event.Timestamp)
			if timestampSec < 0 {
				timestampSec = 0
			}

			result[i] = map[string]interface{}{
				"type":      event.Type,
				"seqNum":    seqNum,
				"timestamp": timestampSec,
				"data":      e.sanitizeForAMF3(event.Data),
			}
		}
		return result
	case state.WebAPIEvent:
		// Handle single WebAPIEvent
		// AMF3 has a 29-bit limit for integers
		seqNum := int(v.SeqNum % (1 << 29))
		// Convert timestamp to seconds ago to keep it small
		timestampSec := int(time.Now().Unix() - v.Timestamp)
		if timestampSec < 0 {
			timestampSec = 0
		}

		return map[string]interface{}{
			"type":      v.Type,
			"seqNum":    seqNum,
			"timestamp": timestampSec,
			"data":      e.sanitizeForAMF3(v.Data),
		}
	default:
		// For other types, use reflection to check if it's a struct
		// and convert to map
		rv := reflect.ValueOf(data)
		if rv.Kind() == reflect.Struct {
			return e.structToMap(rv)
		}
		return data
	}
}

// toAMFCompatibleWithVersion converts Go types to AMF-compatible types with version awareness
func (e *AMFEncoder) toAMFCompatibleWithVersion(data interface{}, version AMFVersion) interface{} {
	if data == nil {
		return nil
	}

	v := reflect.ValueOf(data)

	// Handle pointers
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
		data = v.Interface()
	}

	// Handle known types first for performance
	switch d := data.(type) {
	case BaseResponse:
		result := e.baseResponseToMapWithVersion(d, version)
		// For AMF0, convert to ECMAArray; for AMF3, return plain map
		if version == AMF0 {
			return amf.ECMAArray(result)
		}
		return result
	case ResponseBody:
		result := e.responseBodyToMapWithVersion(d, version)
		// For AMF0, convert to ECMAArray; for AMF3, return plain map
		if version == AMF0 {
			return amf.ECMAArray(result)
		}
		return result
	case ErrorResponse:
		result := e.errorResponseToMapWithVersion(d, version)
		// For AMF0, convert to ECMAArray; for AMF3, return plain map
		if version == AMF0 {
			return amf.ECMAArray(result)
		}
		return result
	case map[string]interface{}:
		// Handle nested maps - recursively convert values
		result := make(map[string]interface{})
		for k, v := range d {
			result[k] = e.toAMFCompatibleWithVersion(v, version)
		}
		// For AMF0, convert to ECMAArray; for AMF3, return plain map
		if version == AMF0 {
			return amf.ECMAArray(result)
		}
		return result
	case []interface{}:
		// Handle arrays - recursively convert elements
		result := make([]interface{}, len(d))
		for i, v := range d {
			result[i] = e.toAMFCompatibleWithVersion(v, version)
		}
		return result
	case time.Time:
		return d
	case []byte:
		// Convert byte arrays to base64 strings for AMF
		return string(d)
	}

	// Handle reflection-based conversion
	switch v.Kind() {
	case reflect.Struct:
		result := e.structToMapWithVersion(v, version)
		if version == AMF3 {
			return amf.ECMAArray(result)
		}
		return result
	case reflect.Slice, reflect.Array:
		// For AMF3, we need to ensure it's []interface{}
		result := e.sliceToArrayWithVersion(v, version)
		if version == AMF3 {
			// AMF3 requires []interface{}, ensure conversion
			return result
		}
		return result
	case reflect.Map:
		result := e.mapToAMFMapWithVersion(v, version)
		// For AMF0, convert to ECMAArray; for AMF3, return plain map
		if version == AMF0 {
			return amf.ECMAArray(result)
		}
		return result
	case reflect.Bool, reflect.String:
		return data
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		// Convert to int for AMF compatibility
		return int(v.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		// Convert to int for AMF compatibility
		return int(v.Uint())
	case reflect.Float32, reflect.Float64:
		return v.Float()
	default:
		// Default: convert to string
		return fmt.Sprintf("%v", data)
	}
}

// toAMFCompatible converts Go types to AMF-compatible types (backward compatibility)
func (e *AMFEncoder) toAMFCompatible(data interface{}) interface{} {
	return e.toAMFCompatibleWithVersion(data, AMF0)
}

// baseResponseToMapWithVersion converts BaseResponse to AMF-compatible map with version awareness
func (e *AMFEncoder) baseResponseToMapWithVersion(resp BaseResponse, version AMFVersion) map[string]interface{} {
	return map[string]interface{}{
		"response": e.responseBodyToMapWithVersion(resp.Response, version),
	}
}

// responseBodyToMapWithVersion converts ResponseBody to AMF-compatible map with version awareness
func (e *AMFEncoder) responseBodyToMapWithVersion(body ResponseBody, version AMFVersion) map[string]interface{} {
	m := map[string]interface{}{
		"statusCode": body.StatusCode,
		"statusText": body.StatusText,
	}
	if body.Data != nil {
		// toAMFCompatibleWithVersion now handles ECMAArray conversion for maps
		m["data"] = e.toAMFCompatibleWithVersion(body.Data, version)
	} else if version == AMF3 {
		// For AMF3, always include data field even if empty to prevent truncation
		m["data"] = map[string]interface{}{}
	}
	return m
}

// errorResponseToMapWithVersion converts ErrorResponse to AMF-compatible map with version awareness
func (e *AMFEncoder) errorResponseToMapWithVersion(err ErrorResponse, version AMFVersion) map[string]interface{} {
	return map[string]interface{}{
		"response": map[string]interface{}{
			"statusCode": err.Response.StatusCode,
			"statusText": err.Response.StatusText,
		},
	}
}

// baseResponseToMap converts BaseResponse to AMF-compatible map
func (e *AMFEncoder) baseResponseToMap(resp BaseResponse) map[string]interface{} {
	return e.baseResponseToMapWithVersion(resp, AMF0)
}

// responseBodyToMap converts ResponseBody to AMF-compatible map
func (e *AMFEncoder) responseBodyToMap(body ResponseBody) map[string]interface{} {
	return e.responseBodyToMapWithVersion(body, AMF0)
}

// errorResponseToMap converts ErrorResponse to AMF-compatible map
func (e *AMFEncoder) errorResponseToMap(err ErrorResponse) map[string]interface{} {
	return e.errorResponseToMapWithVersion(err, AMF0)
}

// structToMapWithVersion converts a struct to a map using JSON tags with version awareness
func (e *AMFEncoder) structToMapWithVersion(v reflect.Value, version AMFVersion) map[string]interface{} {
	result := make(map[string]interface{})
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		// Skip unexported fields
		if !fieldValue.CanInterface() {
			continue
		}

		// Get JSON tag
		jsonTag := field.Tag.Get("json")
		if jsonTag == "-" {
			continue
		}

		// Parse JSON tag
		tagParts := strings.Split(jsonTag, ",")
		fieldName := tagParts[0]
		if fieldName == "" {
			fieldName = field.Name
		}

		// Check for omitempty
		omitEmpty := false
		for _, part := range tagParts[1:] {
			if part == "omitempty" {
				omitEmpty = true
				break
			}
		}

		// Skip if omitempty and value is zero
		if omitEmpty && e.isZeroValue(fieldValue) {
			continue
		}

		// Get field value and convert recursively
		fieldData := fieldValue.Interface()
		result[fieldName] = e.toAMFCompatibleWithVersion(fieldData, version)
	}

	return result
}

// structToMap converts a struct to a map using JSON tags
func (e *AMFEncoder) structToMap(v reflect.Value) map[string]interface{} {
	return e.structToMapWithVersion(v, AMF0)
}

// sliceToArrayWithVersion converts a slice to an AMF-compatible array with version awareness
func (e *AMFEncoder) sliceToArrayWithVersion(v reflect.Value, version AMFVersion) []interface{} {
	length := v.Len()
	result := make([]interface{}, length)

	for i := 0; i < length; i++ {
		elem := v.Index(i)
		if elem.CanInterface() {
			result[i] = e.toAMFCompatibleWithVersion(elem.Interface(), version)
		} else {
			result[i] = nil
		}
	}

	return result
}

// sliceToArray converts a slice to an AMF-compatible array
func (e *AMFEncoder) sliceToArray(v reflect.Value) []interface{} {
	return e.sliceToArrayWithVersion(v, AMF0)
}

// mapToAMFMapWithVersion converts a Go map to an AMF-compatible map with version awareness
func (e *AMFEncoder) mapToAMFMapWithVersion(v reflect.Value, version AMFVersion) map[string]interface{} {
	result := make(map[string]interface{})

	for _, key := range v.MapKeys() {
		// Convert key to string (AMF only supports string keys)
		keyStr := fmt.Sprintf("%v", key.Interface())
		value := v.MapIndex(key)

		if value.CanInterface() {
			result[keyStr] = e.toAMFCompatibleWithVersion(value.Interface(), version)
		}
	}

	return result
}

// mapToAMFMap converts a Go map to an AMF-compatible map
func (e *AMFEncoder) mapToAMFMap(v reflect.Value) map[string]interface{} {
	return e.mapToAMFMapWithVersion(v, AMF0)
}

// convertToMap converts any data to a map structure for AMF3
func (e *AMFEncoder) convertToMap(data interface{}) interface{} {
	if data == nil {
		// For AMF3, return empty map instead of nil to avoid truncation
		return map[string]interface{}{}
	}

	// If already a map, return as-is (even if empty)
	if m, ok := data.(map[string]interface{}); ok {
		if m == nil {
			return map[string]interface{}{}
		}
		return m
	}

	v := reflect.ValueOf(data)

	// Handle pointers
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
		data = v.Interface()
	}

	// Handle different types
	switch v.Kind() {
	case reflect.Struct:
		return e.structToMap(v)
	case reflect.Map:
		return e.mapToAMFMap(v)
	case reflect.Slice, reflect.Array:
		result := make([]interface{}, v.Len())
		for i := 0; i < v.Len(); i++ {
			elem := v.Index(i)
			if elem.CanInterface() {
				result[i] = e.convertToMap(elem.Interface())
			}
		}
		return result
	default:
		// For basic types, return as-is
		return data
	}
}

// isZeroValue checks if a reflect.Value is a zero value
func (e *AMFEncoder) isZeroValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	case reflect.Struct:
		// For time.Time, check if it's zero
		if t, ok := v.Interface().(time.Time); ok {
			return t.IsZero()
		}
		// For other structs, we can't easily determine zero value
		return false
	}
	return false
}

// DetectAMFVersion determines which AMF version to use based on the request
func DetectAMFVersion(r *http.Request) AMFVersion {
	if r == nil {
		return AMF0
	}

	// Check query parameter first (highest priority)
	format := strings.ToLower(r.URL.Query().Get("f"))
	switch format {
	case "amf0":
		return AMF0
	case "amf3":
		return AMF3
	case "amf":
		// Default to AMF3 for modern clients (Gromit expects AMF3)
		return AMF3
	}

	// Check Accept header for version hint
	accept := r.Header.Get("Accept")
	if strings.Contains(accept, "amf3") || strings.Contains(accept, "AMF3") {
		return AMF3
	}
	if strings.Contains(accept, "amf") || strings.Contains(accept, "AMF") {
		return AMF0
	}

	// Check Content-Type header (for POST requests)
	contentType := r.Header.Get("Content-Type")
	if strings.Contains(contentType, "amf3") || strings.Contains(contentType, "AMF3") {
		return AMF3
	}
	if strings.Contains(contentType, "amf") || strings.Contains(contentType, "AMF") {
		return AMF0
	}

	// Default to AMF0
	return AMF0
}

// IsAMFRequest checks if the request is asking for AMF format
func IsAMFRequest(r *http.Request) bool {
	if r == nil {
		return false
	}

	// Check query parameter
	format := strings.ToLower(r.URL.Query().Get("f"))
	if format == "amf" || format == "amf0" || format == "amf3" {
		return true
	}

	// Check Accept header
	accept := strings.ToLower(r.Header.Get("Accept"))
	if strings.Contains(accept, "application/x-amf") ||
		strings.Contains(accept, "application/amf") {
		return true
	}

	return false
}
