package oscar

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

// errWriter is a writer that always returns an error.
type errWriter struct{}

func (errWriter) Write(p []byte) (n int, err error) {
	return 0, io.EOF
}

func TestMarshal(t *testing.T) {
	tests := []struct {
		name    string
		w       io.Writer
		given   any
		want    []byte
		wantErr error
	}{
		{
			name: "marshal uint8",
			w:    &bytes.Buffer{},
			given: struct {
				Val uint8
			}{
				Val: 100,
			},
			want: []byte{0x64},
		},
		{
			name: "marshal uint16",
			w:    &bytes.Buffer{},
			given: struct {
				Val uint16
			}{
				Val: 100,
			},
			want: []byte{0x0, 0x64},
		},
		{
			name: "marshal uint32",
			w:    &bytes.Buffer{},
			given: struct {
				Val uint32
			}{
				Val: 100,
			},
			want: []byte{0x0, 0x0, 0x0, 0x64},
		},
		{
			name: "marshal uint64",
			w:    &bytes.Buffer{},
			given: struct {
				Val uint64
			}{
				Val: 100,
			},
			want: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x64},
		},
		{
			name: "unsupported type",
			w:    &bytes.Buffer{},
			given: struct {
				Val int16
			}{
				Val: 100,
			},
			wantErr: ErrMarshalFailure,
		},
		{
			name: "string8",
			w:    &bytes.Buffer{},
			given: struct {
				Val string `len_prefix:"uint8"`
			}{
				Val: "test-value",
			},
			want: append(
				[]byte{0xa}, /* len prefix */
				[]byte{0x74, 0x65, 0x73, 0x74, 0x2d, 0x76, 0x61, 0x6c, 0x75, 0x65}...), /* str val */
		},
		{
			name: "string8 write error",
			w:    errWriter{},
			given: struct {
				Val string `len_prefix:"uint8"`
			}{
				Val: "test-value",
			},
			wantErr: io.EOF,
		},
		{
			name: "string16",
			w:    &bytes.Buffer{},
			given: struct {
				Val string `len_prefix:"uint16"`
			}{
				Val: "test-value",
			},
			want: append(
				[]byte{0x0, 0xa}, /* len prefix */
				[]byte{0x74, 0x65, 0x73, 0x74, 0x2d, 0x76, 0x61, 0x6c, 0x75, 0x65}...), /* str val */
		},
		{
			name: "string16 write error",
			w:    errWriter{},
			given: struct {
				Val string `len_prefix:"uint16"`
			}{
				Val: "test-value",
			},
			wantErr: io.EOF,
		},
		{
			name: "string with unknown prefix type",
			w:    &bytes.Buffer{},
			given: struct {
				Val string `len_prefix:"uint128"`
			}{
				Val: "test-value",
			},
			wantErr: ErrMarshalFailure,
		},
		{
			name: "byte slice with no prefix",
			w:    &bytes.Buffer{},
			given: struct {
				Val []byte
			}{
				Val: []byte(`hello`),
			},
			want: []byte{0x68, 0x65, 0x6c, 0x6c, 0x6f},
		},
		{
			name: "byte slice with no prefix with write error",
			w:    &bytes.Buffer{},
			given: struct {
				Val []int
			}{
				Val: []int{1, 2, 3, 4},
			},
			wantErr: ErrMarshalFailure,
		},
		{
			name: "byte slice with no prefix with write error",
			w:    errWriter{},
			given: struct {
				Val []byte
			}{
				Val: []byte(`hello`),
			},
			wantErr: io.EOF,
		},
		{
			name: "empty byte slice",
			w:    &bytes.Buffer{},
			given: struct {
				Val []byte
			}{
				Val: []byte{},
			},
			want: []byte(nil),
		},
		{
			name: "byte slice with uint8 len_prefix",
			w:    &bytes.Buffer{},
			given: struct {
				Val []byte `len_prefix:"uint8"`
			}{
				Val: []byte(`hello`),
			},
			want: append(
				[]byte{0x05},                             /* len prefix */
				[]byte{0x68, 0x65, 0x6c, 0x6c, 0x6f}...), /* slice val */
		},
		{
			name: "byte slice with uint8 len_prefix with error",
			w:    errWriter{},
			given: struct {
				Val []byte `len_prefix:"uint8"`
			}{
				Val: []byte(`hello`),
			},
			wantErr: io.EOF,
		},
		{
			name: "byte slice with uint16 len_prefix",
			w:    &bytes.Buffer{},
			given: struct {
				Val []byte `len_prefix:"uint16"`
			}{
				Val: []byte(`hello`),
			},
			want: append(
				[]byte{0x00, 0x05},                       /* len prefix */
				[]byte{0x68, 0x65, 0x6c, 0x6c, 0x6f}...), /* slice val */
		},
		{
			name: "byte slice with uint16 len_prefix with error",
			w:    errWriter{},
			given: struct {
				Val []byte `len_prefix:"uint16"`
			}{
				Val: []byte(`hello`),
			},
			wantErr: io.EOF,
		},
		{
			name: "empty struct slice",
			w:    &bytes.Buffer{},
			given: struct {
				Val []TLV
			}{
				Val: []TLV{},
			},
			want: []byte(nil),
		},
		{
			name: "struct slice with invalid type in struct",
			w:    &bytes.Buffer{},
			given: struct {
				Val []struct {
					Val int16
				}
			}{
				Val: []struct {
					Val int16
				}{
					{
						Val: 1234,
					},
				},
			},
			wantErr: ErrMarshalFailure,
		},
		{
			name: "struct slice with uint8 count_prefix",
			w:    &bytes.Buffer{},
			given: struct {
				Val []TLV `count_prefix:"uint8"`
			}{
				Val: []TLV{
					NewTLV(10, uint16(1234)),
					NewTLV(20, uint16(1234)),
				},
			},
			want: append(
				[]byte{0x02}, /* count prefix */
				[]byte{0x0, 0xa, 0x0, 0x2, 0x4, 0xd2, 0x0, 0x14, 0x0, 0x2, 0x4, 0xd2}...), /* slice val */
		},
		{
			name: "struct slice with uint8 count_prefix with error",
			w:    errWriter{},
			given: struct {
				Val []TLV `count_prefix:"uint8"`
			}{
				Val: []TLV{
					NewTLV(10, uint16(1234)),
				},
			},
			wantErr: io.EOF,
		},
		{
			name: "struct slice with uint16 count_prefix",
			w:    &bytes.Buffer{},
			given: struct {
				Val []TLV `count_prefix:"uint16"`
			}{
				Val: []TLV{
					NewTLV(10, uint16(1234)),
					NewTLV(20, uint16(1234)),
				},
			},
			want: append(
				[]byte{0x00, 0x02}, /* count prefix */
				[]byte{0x0, 0xa, 0x0, 0x2, 0x4, 0xd2, 0x0, 0x14, 0x0, 0x2, 0x4, 0xd2}...), /* slice val */
		},
		{
			name: "struct slice with uint16 count_prefix with error",
			w:    errWriter{},
			given: struct {
				Val []TLV `count_prefix:"uint16"`
			}{
				Val: []TLV{
					NewTLV(10, uint16(1234)),
				},
			},
			wantErr: io.EOF,
		},
		{
			name: "byte slice with uint16 len_prefix and uint16 count_prefix",
			w:    &bytes.Buffer{},
			given: struct {
				Val []byte `len_prefix:"uint16" count_prefix:"uint16"`
			}{
				Val: []byte(`hello`),
			},
			wantErr: ErrMarshalFailure,
		},
		{
			name: "byte slice with unknown len_prefix type",
			w:    &bytes.Buffer{},
			given: struct {
				Val []byte `len_prefix:"uint128"`
			}{
				Val: []byte(`hello`),
			},
			wantErr: ErrMarshalFailure,
		},
		{
			name: "byte slice with unknown count_prefix type",
			w:    &bytes.Buffer{},
			given: struct {
				Val []byte `count_prefix:"uint128"`
			}{
				Val: []byte(`hello`),
			},
			wantErr: ErrMarshalFailure,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Marshal(tt.given, tt.w)
			assert.ErrorIs(t, err, tt.wantErr)
			if tt.wantErr == nil {
				if w, ok := tt.w.(*bytes.Buffer); ok {
					assert.Equal(t, tt.want, w.Bytes())
				}
			}
		})
	}
}
