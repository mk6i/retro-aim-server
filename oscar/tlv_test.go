package oscar

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTLVList_Append(t *testing.T) {
	want := TLVList{
		{
			Tag:   0,
			Value: []byte(`0`),
		},
		{
			Tag:   1,
			Value: []byte(`1`),
		},
		{
			Tag:   2,
			Value: []byte(`2`),
		},
	}

	have := TLVList{}
	have.Append(NewTLV(0, []byte(`0`)))
	have.Append(NewTLV(1, []byte(`1`)))
	have.Append(NewTLV(2, []byte(`2`)))

	assert.Equal(t, want, have)
}

func TestTLVList_AppendList(t *testing.T) {
	want := TLVList{
		{
			Tag:   0,
			Value: []byte(`0`),
		},
		{
			Tag:   1,
			Value: []byte(`1`),
		},
		{
			Tag:   2,
			Value: []byte(`2`),
		},
	}

	have := TLVList{}
	have.AppendList([]TLV{
		NewTLV(0, []byte(`0`)),
		NewTLV(1, []byte(`1`)),
		NewTLV(2, []byte(`2`)),
	})

	assert.Equal(t, want, have)
}

func TestTLVList_Getters(t *testing.T) {
	type args struct {
		tType uint16
	}
	tests := []struct {
		name   string
		given  []TLV
		ttype  any
		lookup func(TLVList) (any, bool)
		expect any
		found  bool
		panic  bool
	}{
		{
			name: "given a TLV of uint32, expect found value",
			given: []TLV{
				NewTLV(0, uint32(12)),
				NewTLV(1, uint32(34)),
				NewTLV(2, uint32(56)),
			},
			lookup: func(l TLVList) (any, bool) {
				return l.Uint32(1)
			},
			expect: uint32(34),
			found:  true,
		},
		{
			name: "given a TLV of uint32, expect not found value",
			given: []TLV{
				NewTLV(0, uint32(12)),
				NewTLV(1, uint32(34)),
				NewTLV(2, uint32(56)),
			},
			lookup: func(l TLVList) (any, bool) {
				return l.Uint32(3)
			},
			expect: uint32(0),
			found:  false,
		},
		{
			name: "given a TLV of uint16, expect found value",
			given: []TLV{
				NewTLV(0, uint16(12)),
				NewTLV(1, uint16(34)),
				NewTLV(2, uint16(56)),
			},
			lookup: func(l TLVList) (any, bool) {
				return l.Uint16(1)
			},
			expect: uint16(34),
			found:  true,
		},
		{
			name: "given a TLV of uint16, expect not found value",
			given: []TLV{
				NewTLV(0, uint16(12)),
				NewTLV(1, uint16(34)),
				NewTLV(2, uint16(56)),
			},
			lookup: func(l TLVList) (any, bool) {
				return l.Uint16(3)
			},
			expect: uint16(0),
			found:  false,
		},
		{
			name: "given a TLV of string, expect found value",
			given: []TLV{
				NewTLV(0, "12"),
				NewTLV(1, "34"),
				NewTLV(2, "56"),
			},
			lookup: func(l TLVList) (any, bool) {
				return l.String(1)
			},
			expect: "34",
			found:  true,
		},
		{
			name: "given a TLV of string, expect not found value",
			given: []TLV{
				NewTLV(0, "12"),
				NewTLV(1, "34"),
				NewTLV(2, "56"),
			},
			lookup: func(l TLVList) (any, bool) {
				return l.String(3)
			},
			expect: "",
			found:  false,
		},
		{
			name: "given a TLV of slice, expect found value",
			given: []TLV{
				NewTLV(0, []byte(`12`)),
				NewTLV(1, []byte(`34`)),
				NewTLV(2, []byte(`56`)),
			},
			lookup: func(l TLVList) (any, bool) {
				return l.Slice(1)
			},
			expect: []byte(`34`),
			found:  true,
		},
		{
			name: "given a TLV of string, expect not found value",
			given: []TLV{
				NewTLV(0, []byte(`12`)),
				NewTLV(1, []byte(`34`)),
				NewTLV(2, []byte(`56`)),
			},
			lookup: func(l TLVList) (any, bool) {
				return l.Slice(3)
			},
			expect: []byte(nil),
			found:  false,
		},
		{
			name: "expect a panic when there's a type mismatch",
			given: []TLV{
				NewTLV(0, uint16(12)),
				NewTLV(1, uint16(34)),
				NewTLV(2, uint16(56)),
			},
			lookup: func(l TLVList) (any, bool) {
				return l.Uint32(1)
			},
			panic: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.panic {
				assert.Panics(t, func() { tt.lookup(tt.given) })
				return
			}
			have, found := tt.lookup(tt.given)
			assert.Equal(t, tt.expect, have)
			assert.Equal(t, tt.found, found)
		})
	}
}

func TestTLVList_NewTLVPanic(t *testing.T) {
	// make sure NewTLV panics when it encounters an unsupported type, in this
	// case it's int.
	assert.Panics(t, func() {
		NewTLV(1, 30)
	})
}
