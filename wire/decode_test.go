package wire

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnmarshal(t *testing.T) {
	tests := []struct {
		name      string
		prototype any
		given     []byte
		want      any
		wantErr   error
	}{
		{
			name: "uint8",
			prototype: &struct {
				Val uint8
			}{},
			want: &struct {
				Val uint8
			}{
				Val: 100,
			},
			given: []byte{0x64},
		},
		{
			name: "uint8 with read error",
			prototype: &struct {
				Val uint8
			}{},
			wantErr: io.EOF,
			given:   []byte{},
		},
		{
			name: "uint16",
			prototype: &struct {
				Val uint16
			}{},
			want: &struct {
				Val uint16
			}{
				Val: 100,
			},
			given: []byte{0x0, 0x64},
		},
		{
			name: "uint16 with read error",
			prototype: &struct {
				Val uint16
			}{},
			wantErr: io.EOF,
			given:   []byte{},
		},
		{
			name: "uint32",
			prototype: &struct {
				Val uint32
			}{},
			want: &struct {
				Val uint32
			}{
				Val: 100,
			},
			given: []byte{0x0, 0x0, 0x0, 0x64},
		},
		{
			name: "uint32 with read error",
			prototype: &struct {
				Val uint32
			}{},
			wantErr: io.EOF,
			given:   []byte{},
		},
		{
			name: "uint64",
			prototype: &struct {
				Val uint64
			}{},
			want: &struct {
				Val uint64
			}{
				Val: 100,
			},
			given: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x64},
		},
		{
			name: "uint64 with read error",
			prototype: &struct {
				Val uint64
			}{},
			wantErr: io.EOF,
			given:   []byte{},
		},
		{
			name: "string8",
			prototype: &struct {
				Val string `len_prefix:"uint8"`
			}{},
			want: &struct {
				Val string `len_prefix:"uint8"`
			}{
				Val: "test-value",
			},
			given: append(
				[]byte{0xa}, /* len prefix */
				[]byte{0x74, 0x65, 0x73, 0x74, 0x2d, 0x76, 0x61, 0x6c, 0x75, 0x65}...), /* str val */
		},
		{
			name: "string8 read error",
			prototype: &struct {
				Val string `len_prefix:"uint8"`
			}{},
			given:   []byte{},
			wantErr: io.EOF,
		},
		{
			name: "string16",
			prototype: &struct {
				Val string `len_prefix:"uint16"`
			}{},
			want: &struct {
				Val string `len_prefix:"uint16"`
			}{
				Val: "test-value",
			},
			given: append(
				[]byte{0x0, 0xa}, /* len prefix */
				[]byte{0x74, 0x65, 0x73, 0x74, 0x2d, 0x76, 0x61, 0x6c, 0x75, 0x65}...), /* str val */
		},
		{
			name: "string16 read error",
			prototype: &struct {
				Val string `len_prefix:"uint16"`
			}{},
			given:   []byte{},
			wantErr: io.EOF,
		},
		{
			name: "unsupported string prefix type",
			prototype: &struct {
				Val string `len_prefix:"uint128"`
			}{},
			wantErr: ErrUnmarshalFailure,
			given: append(
				[]byte{0x0, 0xa}, /* len prefix */
				[]byte{0x74, 0x65, 0x73, 0x74, 0x2d, 0x76, 0x61, 0x6c, 0x75, 0x65}...), /* str val */
		},
		{
			name: "string with missing len_prefix",
			prototype: &struct {
				Val string
			}{},
			wantErr: ErrUnmarshalFailure,
			given: append(
				[]byte{0x0, 0xa}, /* len prefix */
				[]byte{0x74, 0x65, 0x73, 0x74, 0x2d, 0x76, 0x61, 0x6c, 0x75, 0x65}...), /* str val */
		},
		{
			name: "partial string8",
			prototype: &struct {
				Val string `len_prefix:"uint8"`
			}{},
			wantErr: io.EOF,
			given: append(
				[]byte{0xa},  /* len prefix */
				[]byte{}...), /* truncated payload */
		},
		{
			name: "byte slice with uint8 len_prefix",
			prototype: &struct {
				Val []byte `len_prefix:"uint8"`
			}{},
			want: &struct {
				Val []byte `len_prefix:"uint8"`
			}{
				Val: []byte(`hello`),
			},
			given: append(
				[]byte{0x05},                             /* len prefix */
				[]byte{0x68, 0x65, 0x6c, 0x6c, 0x6f}...), /* slice val */
		},
		{
			name: "slice of invalid type with uint8 len_prefix",
			prototype: &struct {
				Val []int `len_prefix:"uint8"`
			}{},
			wantErr: ErrUnmarshalFailure,
			given:   []byte{0x04, 0x65, 0x6c, 0x6c, 0x6f},
		},
		{
			name: "byte slice with uint8 len_prefix with read error",
			prototype: &struct {
				Val []byte `len_prefix:"uint8"`
			}{},
			wantErr: io.EOF,
			given: append(
				[]byte{0x05}, /* len prefix */
				[]byte{}...), /* slice val */
		},
		{
			name: "byte slice with uint8 len_prefix read error",
			prototype: &struct {
				Val []byte `len_prefix:"uint8"`
			}{},
			wantErr: io.EOF,
			given:   []byte{},
		},
		{
			name: "byte slice with uint16 len_prefix",
			prototype: &struct {
				Val []byte `len_prefix:"uint16"`
			}{},
			want: &struct {
				Val []byte `len_prefix:"uint16"`
			}{
				Val: []byte(`hello`),
			},
			given: append(
				[]byte{0x00, 0x05},                       /* len prefix */
				[]byte{0x68, 0x65, 0x6c, 0x6c, 0x6f}...), /* slice val */
		},
		{
			name: "byte slice with uint16 len_prefix read error",
			prototype: &struct {
				Val []byte `len_prefix:"uint16"`
			}{},
			wantErr: io.EOF,
			given:   []byte{},
		},
		{
			name: "byte slice with invalid len_prefix",
			prototype: &struct {
				Val []byte `len_prefix:"uint128"`
			}{},
			wantErr: ErrUnmarshalFailure,
			given: append(
				[]byte{0x00, 0x05},                       /* len prefix */
				[]byte{0x68, 0x65, 0x6c, 0x6c, 0x6f}...), /* slice val */
		},
		{
			name: "struct slice without prefix",
			prototype: &struct {
				Val []TLV
			}{},
			want: &struct {
				Val []TLV
			}{
				Val: []TLV{
					NewTLV(10, uint16(1234)),
					NewTLV(20, uint16(1234)),
				},
			},
			given: []byte{0x0, 0xa, 0x0, 0x2, 0x4, 0xd2, 0x0, 0x14, 0x0, 0x2, 0x4, 0xd2},
		},
		{
			name: "slice of unsupported type without prefix",
			prototype: &struct {
				Val []int
			}{},
			wantErr: ErrUnmarshalFailure,
			given:   []byte{0x0, 0xa, 0x0, 0x2},
		},
		{
			name: "struct slice with uint8 count_prefix",
			prototype: &struct {
				Val []TLV `count_prefix:"uint8"`
			}{},
			want: &struct {
				Val []TLV `count_prefix:"uint8"`
			}{
				Val: []TLV{
					NewTLV(10, uint16(1234)),
					NewTLV(20, uint16(1234)),
				},
			},
			given: append(
				[]byte{0x02}, /* count prefix */
				[]byte{0x0, 0xa, 0x0, 0x2, 0x4, 0xd2, 0x0, 0x14, 0x0, 0x2, 0x4, 0xd2}...), /* slice val */
		},
		{
			name: "struct slice with uint8 count_prefix and unsupported type",
			prototype: &struct {
				Val []struct {
					Val int16
				} `count_prefix:"uint8"`
			}{},
			wantErr: ErrUnmarshalFailure,
			given: append(
				[]byte{0x02}, /* count prefix */
				[]byte{0x0, 0xa, 0x0, 0x2, 0x4, 0xd2, 0x0, 0x14, 0x0, 0x2, 0x4, 0xd2}...), /* slice val */
		},
		{
			name: "struct slice with uint8 count_prefix read error",
			prototype: &struct {
				Val []TLV `count_prefix:"uint8"`
			}{},
			wantErr: io.EOF,
			given:   []byte{},
		},
		{
			name: "struct slice with uint16 count_prefix",
			prototype: &struct {
				Val []TLV `count_prefix:"uint16"`
			}{},
			want: &struct {
				Val []TLV `count_prefix:"uint16"`
			}{
				Val: []TLV{
					NewTLV(10, uint16(1234)),
					NewTLV(20, uint16(1234)),
				},
			},
			given: append(
				[]byte{0x0, 0x02}, /* count prefix */
				[]byte{0x0, 0xa, 0x0, 0x2, 0x4, 0xd2, 0x0, 0x14, 0x0, 0x2, 0x4, 0xd2}...), /* slice val */
		},
		{
			name: "struct slice with uint16 count_prefix and unsupported type",
			prototype: &struct {
				Val []struct {
					Val int16
				} `count_prefix:"uint16"`
			}{},
			wantErr: ErrUnmarshalFailure,
			given: append(
				[]byte{0x0, 0x02}, /* count prefix */
				[]byte{0x0, 0xa, 0x0, 0x2, 0x4, 0xd2, 0x0, 0x14, 0x0, 0x2, 0x4, 0xd2}...), /* slice val */
		},
		{
			name: "struct slice with uint16 count_prefix read error",
			prototype: &struct {
				Val []TLV `count_prefix:"uint16"`
			}{},
			wantErr: io.EOF,
			given:   []byte{},
		},
		{
			name: "struct slice with invalid count_prefix",
			prototype: &struct {
				Val []TLV `count_prefix:"uint128"`
			}{},
			wantErr: ErrUnmarshalFailure,
			given: append(
				[]byte{0x0, 0x02}, /* count prefix */
				[]byte{0x0, 0xa, 0x0, 0x2, 0x4, 0xd2, 0x0, 0x14, 0x0, 0x2, 0x4, 0xd2}...), /* slice val */
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			r := bytes.NewBuffer(tt.given)

			err := Unmarshal(tt.prototype, r)
			assert.ErrorIs(t, err, tt.wantErr)
			if tt.wantErr == nil {
				assert.Equal(t, tt.want, tt.prototype)
			}
		})
	}
}
