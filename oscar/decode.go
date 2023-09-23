package oscar

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"reflect"
)

func Unmarshal(v any, r io.Reader) error {
	return unmarshal(reflect.TypeOf(v).Elem(), reflect.ValueOf(v).Elem(), "", r)
}

func unmarshal(t reflect.Type, v reflect.Value, tag reflect.StructTag, r io.Reader) error {
	switch v.Kind() {
	case reflect.Struct:
		switch v.Interface().(type) {
		case TLVRestBlock:
			val := TLVRestBlock{}
			if err := val.read(r); err != nil {
				return err
			}
			v.Set(reflect.ValueOf(val))
		case TLVLBlock:
			val := TLVLBlock{}
			if err := val.read(r); err != nil {
				return err
			}
			v.Set(reflect.ValueOf(val))
		case TLVBlock:
			val := TLVBlock{}
			if err := val.read(r); err != nil {
				return err
			}
			v.Set(reflect.ValueOf(val))
		default:
			for i := 0; i < v.NumField(); i++ {
				if err := unmarshal(t.Field(i).Type, v.Field(i), t.Field(i).Tag, r); err != nil {
					return err
				}
			}
		}
	case reflect.String:
		var bufLen int
		if lenTag, ok := tag.Lookup("len_prefix"); ok {
			switch lenTag {
			case "uint8":
				var l uint8
				if err := binary.Read(r, binary.BigEndian, &l); err != nil {
					return err
				}
				bufLen = int(l)
			case "uint16":
				var l uint16
				if err := binary.Read(r, binary.BigEndian, &l); err != nil {
					return err
				}
				bufLen = int(l)
			default:
				panic("invalid len_prefix")
			}
		} else {
			panic("string length not set")
		}
		buf := make([]byte, bufLen)
		if _, err := r.Read(buf); err != nil {
			return err
		}
		// todo is there a more efficient way?
		v.SetString(string(buf))
	case reflect.Uint8:
		var l uint8
		if err := binary.Read(r, binary.BigEndian, &l); err != nil {
			return err
		}
		v.Set(reflect.ValueOf(l))
	case reflect.Uint16:
		var l uint16
		if err := binary.Read(r, binary.BigEndian, &l); err != nil {
			return err
		}
		v.Set(reflect.ValueOf(l))
	case reflect.Uint32:
		var l uint32
		if err := binary.Read(r, binary.BigEndian, &l); err != nil {
			return err
		}
		v.Set(reflect.ValueOf(l))
	case reflect.Uint64:
		var l uint64
		if err := binary.Read(r, binary.BigEndian, &l); err != nil {
			return err
		}
		v.Set(reflect.ValueOf(l))
	case reflect.Slice:
		var bufLen int
		if lenTag, ok := tag.Lookup("len_prefix"); ok {
			switch lenTag {
			case "uint8":
				var l uint8
				if err := binary.Read(r, binary.BigEndian, &l); err != nil {
					return err
				}
				bufLen = int(l)
			case "uint16":
				var l uint16
				if err := binary.Read(r, binary.BigEndian, &l); err != nil {
					return err
				}
				bufLen = int(l)
			default:
				panic("length not set")
			}

			buf := make([]byte, bufLen)
			if _, err := r.Read(buf); err != nil {
				return err
			}
			b := bytes.NewBuffer(buf)
			slice := reflect.New(v.Type()).Elem()
			for b.Len() > 0 {
				v1 := reflect.New(v.Type().Elem()).Interface()
				if err := Unmarshal(v1, b); err != nil {
					return err
				}
				slice = reflect.Append(slice, reflect.ValueOf(v1).Elem())
			}
			v.Set(slice)
		} else {
			slice := reflect.New(v.Type()).Elem()
			for {
				v1 := reflect.New(v.Type().Elem()).Interface()
				if err := Unmarshal(v1, r); err != nil {
					if err == io.EOF {
						break
					}
					return err
				}
				slice = reflect.Append(slice, reflect.ValueOf(v1).Elem())
			}
			v.Set(slice)
		}
	case reflect.Array:
		buf := make([]byte, v.Len())
		if _, err := r.Read(buf); err != nil {
			return err
		}
		array := reflect.New(v.Type()).Elem()
		for j := 0; j < len(buf); j++ {
			array.Index(j).SetUint(uint64(buf[j]))
		}
		v.Set(array)
	default:
		return errors.New("unsupported type for unmarshalling")
	}

	return nil
}
