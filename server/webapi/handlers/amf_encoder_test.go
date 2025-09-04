package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/speps/go-amf"
)

func TestAMFEncoderBasicTypes(t *testing.T) {
	encoder := NewAMFEncoder(nil)

	tests := []struct {
		name    string
		input   interface{}
		version AMFVersion
		wantErr bool
	}{
		{"String AMF0", "hello world", AMF0, false},
		{"Number AMF0", 42, AMF0, false},
		{"Float AMF0", 3.14159, AMF0, false},
		{"Boolean AMF0", true, AMF0, false},
		{"Null AMF0", nil, AMF0, false},
		{"String AMF3", "hello world", AMF3, false},
		{"Number AMF3", 42, AMF3, false},
		{"Float AMF3", 3.14159, AMF3, false},
		{"Boolean AMF3", false, AMF3, false},
		{"Null AMF3", nil, AMF3, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := encoder.EncodeAMF(tt.input, tt.version)
			if (err != nil) != tt.wantErr {
				t.Fatalf("EncodeAMF() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && len(data) == 0 {
				t.Fatal("EncodeAMF() returned empty data")
			}

			// Try to decode the data to verify it's valid AMF
			if !tt.wantErr {
				var decodeErr error

				switch tt.version {
				case AMF0:
					_, decodeErr = amf.DecodeAMF0(data)
				case AMF3:
					_, decodeErr = amf.DecodeAMF3(data)
				}

				if decodeErr != nil {
					t.Fatalf("Failed to decode AMF data: %v", decodeErr)
				}
			}
		})
	}
}

func TestAMFEncoderComplexTypes(t *testing.T) {
	encoder := NewAMFEncoder(nil)

	tests := []struct {
		name    string
		input   interface{}
		version AMFVersion
	}{
		{
			name: "Map",
			input: map[string]interface{}{
				"name":   "John Doe",
				"age":    30,
				"active": true,
			},
			version: AMF0,
		},
		{
			name: "Array",
			input: []interface{}{
				"item1",
				42,
				true,
				nil,
			},
			version: AMF0,
		},
		{
			name: "BaseResponse",
			input: BaseResponse{
				Response: ResponseBody{
					StatusCode: 200,
					StatusText: "OK",
					Data: map[string]interface{}{
						"user":   "testuser",
						"online": true,
						"buddies": []interface{}{
							"friend1",
							"friend2",
						},
					},
				},
			},
			version: AMF0,
		},
		{
			name: "ErrorResponse",
			input: ErrorResponse{
				Response: struct {
					StatusCode int    `json:"statusCode" xml:"statusCode"`
					StatusText string `json:"statusText" xml:"statusText"`
				}{
					StatusCode: 404,
					StatusText: "Not Found",
				},
			},
			version: AMF0,
		},
		{
			name: "Time",
			input: map[string]interface{}{
				"timestamp": time.Now(),
				"name":      "Event",
			},
			version: AMF0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := encoder.EncodeAMF(tt.input, tt.version)
			if err != nil {
				t.Fatalf("EncodeAMF() error = %v", err)
			}

			if len(data) == 0 {
				t.Fatal("EncodeAMF() returned empty data")
			}

			// Verify the data is valid AMF
			var decodeErr error

			switch tt.version {
			case AMF0:
				_, decodeErr = amf.DecodeAMF0(data)
			case AMF3:
				_, decodeErr = amf.DecodeAMF3(data)
			}

			if decodeErr != nil {
				t.Fatalf("Failed to decode AMF data: %v", decodeErr)
			}

			// Log the size for performance comparison
			t.Logf("%s: %d bytes", tt.name, len(data))
		})
	}
}

func TestDetectAMFVersion(t *testing.T) {
	tests := []struct {
		name     string
		request  *http.Request
		expected AMFVersion
	}{
		{
			name:     "Query parameter amf0",
			request:  httptest.NewRequest("GET", "/?f=amf0", nil),
			expected: AMF0,
		},
		{
			name:     "Query parameter amf3",
			request:  httptest.NewRequest("GET", "/?f=amf3", nil),
			expected: AMF3,
		},
		{
			name:     "Query parameter amf",
			request:  httptest.NewRequest("GET", "/?f=amf", nil),
			expected: AMF0,
		},
		{
			name: "Accept header AMF3",
			request: func() *http.Request {
				req := httptest.NewRequest("GET", "/", nil)
				req.Header.Set("Accept", "application/x-amf3")
				return req
			}(),
			expected: AMF3,
		},
		{
			name: "Accept header AMF",
			request: func() *http.Request {
				req := httptest.NewRequest("GET", "/", nil)
				req.Header.Set("Accept", "application/x-amf")
				return req
			}(),
			expected: AMF0,
		},
		{
			name:     "No AMF indication",
			request:  httptest.NewRequest("GET", "/", nil),
			expected: AMF0,
		},
		{
			name:     "Nil request",
			request:  nil,
			expected: AMF0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version := DetectAMFVersion(tt.request)
			if version != tt.expected {
				t.Errorf("DetectAMFVersion() = %v, want %v", version, tt.expected)
			}
		})
	}
}

func TestIsAMFRequest(t *testing.T) {
	tests := []struct {
		name     string
		request  *http.Request
		expected bool
	}{
		{
			name:     "Query parameter amf",
			request:  httptest.NewRequest("GET", "/?f=amf", nil),
			expected: true,
		},
		{
			name:     "Query parameter amf0",
			request:  httptest.NewRequest("GET", "/?f=amf0", nil),
			expected: true,
		},
		{
			name:     "Query parameter amf3",
			request:  httptest.NewRequest("GET", "/?f=amf3", nil),
			expected: true,
		},
		{
			name: "Accept header",
			request: func() *http.Request {
				req := httptest.NewRequest("GET", "/", nil)
				req.Header.Set("Accept", "application/x-amf")
				return req
			}(),
			expected: true,
		},
		{
			name:     "JSON format",
			request:  httptest.NewRequest("GET", "/?f=json", nil),
			expected: false,
		},
		{
			name:     "No format",
			request:  httptest.NewRequest("GET", "/", nil),
			expected: false,
		},
		{
			name:     "Nil request",
			request:  nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsAMFRequest(tt.request)
			if result != tt.expected {
				t.Errorf("IsAMFRequest() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSendAMF(t *testing.T) {
	tests := []struct {
		name         string
		request      *http.Request
		data         interface{}
		expectStatus int
	}{
		{
			name:    "Simple response",
			request: httptest.NewRequest("GET", "/?f=amf", nil),
			data: BaseResponse{
				Response: ResponseBody{
					StatusCode: 200,
					StatusText: "OK",
					Data:       map[string]interface{}{"test": "value"},
				},
			},
			expectStatus: http.StatusOK,
		},
		{
			name:    "AMF0 response with array",
			request: httptest.NewRequest("GET", "/?f=amf0", nil),
			data: BaseResponse{
				Response: ResponseBody{
					StatusCode: 200,
					StatusText: "OK",
					Data:       []interface{}{"item1", "item2"},
				},
			},
			expectStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// First test if the encoder can handle the data
			encoder := NewAMFEncoder(nil)
			version := DetectAMFVersion(tt.request)
			_, encodeErr := encoder.EncodeAMF(tt.data, version)
			if encodeErr != nil {
				t.Fatalf("Encoding failed: %v", encodeErr)
			}

			w := httptest.NewRecorder()
			SendAMF(w, tt.request, tt.data, nil)

			resp := w.Result()
			if resp.StatusCode != tt.expectStatus {
				t.Errorf("Expected status %d, got %d", tt.expectStatus, resp.StatusCode)
				// Print response body for debugging
				body := w.Body.String()
				t.Logf("Response body: %s", body)
			}

			contentType := resp.Header.Get("Content-Type")
			if contentType != "application/x-amf" {
				t.Errorf("Expected Content-Type application/x-amf, got %s", contentType)
			}

			body := w.Body.Bytes()
			if len(body) == 0 {
				t.Error("Response body is empty")
			}
		})
	}
}

func TestStructToMap(t *testing.T) {
	encoder := NewAMFEncoder(nil)

	type TestStruct struct {
		Name     string `json:"name"`
		Age      int    `json:"age"`
		Active   bool   `json:"active"`
		Hidden   string `json:"-"`
		Optional string `json:"optional,omitempty"`
		NoTag    string
	}

	testStruct := TestStruct{
		Name:     "John",
		Age:      30,
		Active:   true,
		Hidden:   "should not appear",
		Optional: "", // should be omitted
		NoTag:    "should appear with field name",
	}

	result := encoder.toAMFCompatible(testStruct)
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Expected map[string]interface{}")
	}

	// Check expected fields
	if resultMap["name"] != "John" {
		t.Errorf("Expected name=John, got %v", resultMap["name"])
	}
	if resultMap["age"] != 30 {
		t.Errorf("Expected age=30, got %v", resultMap["age"])
	}
	if resultMap["active"] != true {
		t.Errorf("Expected active=true, got %v", resultMap["active"])
	}
	if resultMap["NoTag"] != "should appear with field name" {
		t.Errorf("Expected NoTag field, got %v", resultMap["NoTag"])
	}

	// Check omitted fields
	if _, exists := resultMap["Hidden"]; exists {
		t.Error("Hidden field should not appear")
	}
	if _, exists := resultMap["optional"]; exists {
		t.Error("Optional empty field should be omitted")
	}
}

func TestSliceToArray(t *testing.T) {
	encoder := NewAMFEncoder(nil)

	input := []interface{}{
		"string",
		42,
		true,
		nil,
		map[string]interface{}{"nested": "value"},
	}

	result := encoder.toAMFCompatible(input)
	resultArray, ok := result.([]interface{})
	if !ok {
		t.Fatal("Expected []interface{}")
	}

	if len(resultArray) != 5 {
		t.Errorf("Expected 5 elements, got %d", len(resultArray))
	}

	if resultArray[0] != "string" {
		t.Errorf("Expected first element to be 'string', got %v", resultArray[0])
	}
	if resultArray[1] != 42 {
		t.Errorf("Expected second element to be 42, got %v", resultArray[1])
	}
	if resultArray[2] != true {
		t.Errorf("Expected third element to be true, got %v", resultArray[2])
	}
	if resultArray[3] != nil {
		t.Errorf("Expected fourth element to be nil, got %v", resultArray[3])
	}

	nested, ok := resultArray[4].(map[string]interface{})
	if !ok {
		t.Error("Expected fifth element to be map")
	} else if nested["nested"] != "value" {
		t.Errorf("Expected nested value, got %v", nested["nested"])
	}
}

// Benchmark tests
func BenchmarkAMFEncoding(b *testing.B) {
	encoder := NewAMFEncoder(nil)
	data := BaseResponse{
		Response: ResponseBody{
			StatusCode: 200,
			StatusText: "OK",
			Data: map[string]interface{}{
				"users": []interface{}{
					map[string]interface{}{"name": "user1", "online": true},
					map[string]interface{}{"name": "user2", "online": false},
					map[string]interface{}{"name": "user3", "online": true},
				},
				"timestamp": time.Now().Unix(),
				"server":    "retro-aim-server",
			},
		},
	}

	b.Run("AMF0", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = encoder.EncodeAMF(data, AMF0)
		}
	})

	b.Run("AMF3", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = encoder.EncodeAMF(data, AMF3)
		}
	})
}

func TestZeroValueDetection(t *testing.T) {
	encoder := NewAMFEncoder(nil)

	type TestStruct struct {
		EmptyString string    `json:"emptyString,omitempty"`
		ZeroInt     int       `json:"zeroInt,omitempty"`
		FalseValue  bool      `json:"falseValue,omitempty"`
		ZeroTime    time.Time `json:"zeroTime,omitempty"`
		ValidString string    `json:"validString,omitempty"`
		ValidInt    int       `json:"validInt,omitempty"`
		TrueValue   bool      `json:"trueValue,omitempty"`
	}

	testStruct := TestStruct{
		EmptyString: "",
		ZeroInt:     0,
		FalseValue:  false,
		ZeroTime:    time.Time{},
		ValidString: "test",
		ValidInt:    42,
		TrueValue:   true,
	}

	result := encoder.toAMFCompatible(testStruct)
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Expected map[string]interface{}")
	}

	// Should be omitted (zero values)
	omittedFields := []string{"emptyString", "zeroInt", "falseValue", "zeroTime"}
	for _, field := range omittedFields {
		if _, exists := resultMap[field]; exists {
			t.Errorf("Field %s should be omitted (zero value)", field)
		}
	}

	// Should be present (non-zero values)
	presentFields := map[string]interface{}{
		"validString": "test",
		"validInt":    42,
		"trueValue":   true,
	}
	for field, expected := range presentFields {
		if actual, exists := resultMap[field]; !exists {
			t.Errorf("Field %s should be present", field)
		} else if actual != expected {
			t.Errorf("Field %s: expected %v, got %v", field, expected, actual)
		}
	}
}
