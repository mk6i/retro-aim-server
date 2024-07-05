package wire

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"reflect"
)

var ErrMarshalFailure = errors.New("failed to marshal")
var ErrMarshalFailureNilSNAC = errors.New("attempting to marshal a nil SNAC")

func Marshal(v any, w io.Writer) error {
	return marshal(reflect.TypeOf(v), reflect.ValueOf(v), "", w)
}

func marshal(t reflect.Type, v reflect.Value, tag reflect.StructTag, w io.Writer) error {
	if t == nil {
		return ErrMarshalFailureNilSNAC
	}
	switch t.Kind() {
	case reflect.Struct:
		marshalEachField := func(w io.Writer) error {
			for i := 0; i < t.NumField(); i++ {
				if err := marshal(t.Field(i).Type, v.Field(i), t.Field(i).Tag, w); err != nil {
					return err
				}
			}
			return nil
		}
		if lenTag, ok := tag.Lookup("len_prefix"); ok {
			buf := &bytes.Buffer{}
			if err := marshalEachField(buf); err != nil {
				return err
			}
			// write struct length
			if err := writeUnsignedInt(lenTag, buf.Len(), w); err != nil {
				return err
			}
			// write struct bytes
			if buf.Len() > 0 {
				_, err := w.Write(buf.Bytes())
				return err
			}
			return nil
		}
		return marshalEachField(w)
	case reflect.String:
		if lenTag, ok := tag.Lookup("len_prefix"); ok {
			if err := writeUnsignedInt(lenTag, len(v.String()), w); err != nil {
				return err
			}
		}
		return binary.Write(w, binary.BigEndian, []byte(v.String()))
	case reflect.Slice:
		// todo: only write to temporary buffer if len_prefix is set
		buf := &bytes.Buffer{}
		if t.Elem().Kind() == reflect.Struct {
			for j := 0; j < v.Len(); j++ {
				element := v.Index(j)
				if err := Marshal(element.Interface(), buf); err != nil {
					return err
				}
			}
		} else {
			if err := binary.Write(buf, binary.BigEndian, v.Interface()); err != nil {
				return fmt.Errorf("%w: error marshalling %s", ErrMarshalFailure, t.Elem().Kind())
			}
		}

		var hasLenPrefix bool
		if l, ok := tag.Lookup("len_prefix"); ok {
			hasLenPrefix = true
			if err := writeUnsignedInt(l, buf.Len(), w); err != nil {
				return err
			}
		}
		if l, ok := tag.Lookup("count_prefix"); ok {
			if hasLenPrefix {
				return fmt.Errorf("%w: struct elem has both len_prefix and count_prefix: ", ErrMarshalFailure)
			}
			if err := writeUnsignedInt(l, v.Len(), w); err != nil {
				return err
			}
		}
		if buf.Len() > 0 {
			_, err := w.Write(buf.Bytes())
			return err
		}
		return nil
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return binary.Write(w, binary.BigEndian, v.Interface())
	default:
		return fmt.Errorf("%w: unsupported type %v", ErrMarshalFailure, t.Kind())
	}
}

func writeUnsignedInt(intType string, intVal int, w io.Writer) error {
	switch intType {
	case "uint8":
		if err := binary.Write(w, binary.BigEndian, uint8(intVal)); err != nil {
			return err
		}
	case "uint16":
		if err := binary.Write(w, binary.BigEndian, uint16(intVal)); err != nil {
			return err
		}
	default:
		return fmt.Errorf("%w: unsupported type %s. allowed types: uint8, uint16", ErrMarshalFailure, intType)
	}
	return nil
}
