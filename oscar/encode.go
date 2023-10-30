package oscar

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"reflect"
)

func Marshal(v any, w io.Writer) error {
	return marshal(reflect.TypeOf(v), reflect.ValueOf(v), "", w)
}

func marshal(t reflect.Type, v reflect.Value, tag reflect.StructTag, w io.Writer) error {
	switch t.Kind() {
	case reflect.Struct:
		for i := 0; i < t.NumField(); i++ {
			if err := marshal(t.Field(i).Type, v.Field(i), t.Field(i).Tag, w); err != nil {
				return err
			}
		}
	case reflect.String:
		if l, ok := tag.Lookup("len_prefix"); ok {
			switch l {
			case "uint8":
				if err := binary.Write(w, binary.BigEndian, uint8(len(v.String()))); err != nil {
					return err
				}
			case "uint16":
				if err := binary.Write(w, binary.BigEndian, uint16(len(v.String()))); err != nil {
					return err
				}
			default:
				panic("length not set")
			}
		}
		if err := binary.Write(w, binary.BigEndian, []byte(v.String())); err != nil {
			return err
		}
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
				return err
			}
		}
		//todo what if both len_prefix and count_prefix are set?
		if l, ok := tag.Lookup("len_prefix"); ok {
			switch l {
			case "uint8":
				if err := binary.Write(w, binary.BigEndian, uint8(buf.Len())); err != nil {
					return err
				}
			case "uint16":
				if err := binary.Write(w, binary.BigEndian, uint16(buf.Len())); err != nil {
					return err
				}
			}
		}
		if l, ok := tag.Lookup("count_prefix"); ok {
			switch l {
			case "uint8":
				if err := binary.Write(w, binary.BigEndian, uint8(v.Len())); err != nil {
					return err
				}
			case "uint16":
				if err := binary.Write(w, binary.BigEndian, uint16(v.Len())); err != nil {
					return err
				}
			}
		}
		if _, err := w.Write(buf.Bytes()); err != nil {
			return err
		}
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Array:
		if err := binary.Write(w, binary.BigEndian, v.Interface()); err != nil {
			return err
		}
	default:
		return errors.New("unsupported type for marshalling")
	}

	return nil
}
