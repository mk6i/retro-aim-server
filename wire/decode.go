package wire

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"reflect"
)

var ErrUnmarshalFailure = errors.New("failed to unmarshal")

func Unmarshal(v any, r io.Reader) error {
	return unmarshal(reflect.TypeOf(v).Elem(), reflect.ValueOf(v).Elem(), "", r)
}

func unmarshal(t reflect.Type, v reflect.Value, tag reflect.StructTag, r io.Reader) error {
	switch v.Kind() {
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			if err := unmarshal(t.Field(i).Type, v.Field(i), t.Field(i).Tag, r); err != nil {
				return err
			}
		}
		return nil
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
				return fmt.Errorf("%w: unsupported len_prefix type %s. allowed types: uint8, uint16", ErrUnmarshalFailure, lenTag)
			}
		} else {
			return fmt.Errorf("%w: missing len_prefix tag", ErrUnmarshalFailure)
		}
		buf := make([]byte, bufLen)
		if bufLen > 0 {
			if _, err := io.ReadFull(r, buf); err != nil {
				return err
			}
		}
		// todo is there a more efficient way?
		v.SetString(string(buf))
		return nil
	case reflect.Uint8:
		var l uint8
		if err := binary.Read(r, binary.BigEndian, &l); err != nil {
			return err
		}
		v.Set(reflect.ValueOf(l))
		return nil
	case reflect.Uint16:
		var l uint16
		if err := binary.Read(r, binary.BigEndian, &l); err != nil {
			return err
		}
		v.Set(reflect.ValueOf(l))
		return nil
	case reflect.Uint32:
		var l uint32
		if err := binary.Read(r, binary.BigEndian, &l); err != nil {
			return err
		}
		v.Set(reflect.ValueOf(l))
		return nil
	case reflect.Uint64:
		var l uint64
		if err := binary.Read(r, binary.BigEndian, &l); err != nil {
			return err
		}
		v.Set(reflect.ValueOf(l))
		return nil
	case reflect.Slice:
		if lenTag, ok := tag.Lookup("len_prefix"); ok {
			var bufLen int
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
				return fmt.Errorf("%w: unsupported len_prefix type %s. allowed types: uint8, uint16", ErrUnmarshalFailure, lenTag)
			}

			buf := make([]byte, bufLen)
			if bufLen > 0 {
				if _, err := io.ReadFull(r, buf); err != nil {
					return err
				}
			}
			b := bytes.NewBuffer(buf)
			slice := reflect.New(v.Type()).Elem()
			// todo: if this is a slice of scalars, there should be no need to
			//  call Unmarshal on each element. it should be possible to just
			//  call binary.Read(r, binary.BigEndian, []byte)
			for b.Len() > 0 {
				v1 := reflect.New(v.Type().Elem()).Interface()
				if err := Unmarshal(v1, b); err != nil {
					return err
				}
				slice = reflect.Append(slice, reflect.ValueOf(v1).Elem())
			}
			v.Set(slice)
		} else if countTag, ok := tag.Lookup("count_prefix"); ok {
			var count int
			switch countTag {
			case "uint8":
				var l uint8
				if err := binary.Read(r, binary.BigEndian, &l); err != nil {
					return err
				}
				count = int(l)
			case "uint16":
				var l uint16
				if err := binary.Read(r, binary.BigEndian, &l); err != nil {
					return err
				}
				count = int(l)
			default:
				return fmt.Errorf("%w: unsupported count_prefix type %s. allowed types: uint8, uint16", ErrUnmarshalFailure, lenTag)
			}

			slice := reflect.New(v.Type()).Elem()
			for i := 0; i < count; i++ {
				v1 := reflect.New(v.Type().Elem()).Interface()
				if err := Unmarshal(v1, r); err != nil {
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
		return nil
	default:
		return fmt.Errorf("%w: unsupported type %v", ErrUnmarshalFailure, t.Kind())
	}
}
