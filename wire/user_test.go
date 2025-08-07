package wire

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWeakMD5PasswordHash(t *testing.T) {
	tests := []struct {
		name     string
		password string
		authKey  string
		want     []byte
	}{
		{
			name:     "empty password and auth key",
			password: "",
			authKey:  "",
			want:     []byte{0x13, 0xfd, 0x0b, 0x9e, 0x89, 0xf4, 0xb8, 0x36, 0xa7, 0x65, 0x8b, 0x9d, 0xca, 0xad, 0x2a, 0xd4},
		},
		{
			name:     "password and auth key",
			password: "password123",
			authKey:  "authkey456",
			want:     []byte{0x04, 0x79, 0x63, 0x82, 0x0d, 0xa7, 0xbb, 0xfe, 0x6a, 0x9b, 0x41, 0xa4, 0x5c, 0x47, 0xcb, 0xcb},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WeakMD5PasswordHash(tt.password, tt.authKey)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestStrongMD5PasswordHash(t *testing.T) {
	tests := []struct {
		name     string
		password string
		authKey  string
		want     []byte
	}{
		{
			name:     "empty password and auth key",
			password: "",
			authKey:  "",
			want:     []byte{0x1f, 0xa2, 0xb6, 0x99, 0x59, 0x84, 0xb0, 0x14, 0x68, 0xa3, 0x7c, 0x77, 0x42, 0x90, 0x0a, 0xc9},
		},
		{
			name:     "password and auth key",
			password: "password123",
			authKey:  "authkey456",
			want:     []byte{0xb9, 0x07, 0x91, 0xcc, 0xcb, 0x5c, 0x57, 0x71, 0xbd, 0xcb, 0xc9, 0x39, 0x82, 0xf7, 0x94, 0x84},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StrongMD5PasswordHash(tt.password, tt.authKey)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRoastOSCARPassword(t *testing.T) {
	tests := []struct {
		name        string
		roastedPass []byte
		want        []byte
	}{
		{
			name:        "empty password",
			roastedPass: []byte{},
			want:        []byte{},
		},
		{
			name:        "single byte password",
			roastedPass: []byte{0xF3},
			want:        []byte{0x00},
		},
		{
			name:        "multiple bytes password",
			roastedPass: []byte{0xF3, 0x26, 0x81, 0xC4},
			want:        []byte{0x00, 0x00, 0x00, 0x00},
		},
		{
			name:        "password longer than roast table",
			roastedPass: []byte{0xF3, 0x26, 0x81, 0xC4, 0x39, 0x86, 0xDB, 0x92, 0x71, 0xA3, 0xB9, 0xE6, 0x53, 0x7A, 0x95, 0x7C, 0xF3, 0x26},
			want:        []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		{
			name:        "non-zero roasted password",
			roastedPass: []byte{0xE3, 0x16, 0x91, 0xD4},
			want:        []byte{0x10, 0x30, 0x10, 0x10},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RoastOSCARPassword(tt.roastedPass)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRoastKerberosPassword(t *testing.T) {
	tests := []struct {
		name        string
		roastedPass []byte
		want        []byte
	}{
		{
			name:        "empty password",
			roastedPass: []byte{},
			want:        []byte{},
		},
		{
			name:        "single byte password",
			roastedPass: []byte{0x76},
			want:        []byte{0x00},
		},
		{
			name:        "multiple bytes password",
			roastedPass: []byte{0x76, 0x91, 0xc5, 0xe7},
			want:        []byte{0x00, 0x00, 0x00, 0x00},
		},
		{
			name:        "password longer than roast table",
			roastedPass: []byte{0x76, 0x91, 0xc5, 0xe7, 0xd0, 0xd9, 0x95, 0xdd, 0x9e, 0x2F, 0xea, 0xd8, 0x6B, 0x21, 0xc2, 0xbc, 0x76, 0x91},
			want:        []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		{
			name:        "non-zero roasted password",
			roastedPass: []byte{0x66, 0x81, 0xd5, 0xf7},
			want:        []byte{0x10, 0x10, 0x10, 0x10},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RoastKerberosPassword(tt.roastedPass)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRoastOSCARJavaPassword(t *testing.T) {
	tests := []struct {
		name        string
		roastedPass []byte
		want        []byte
	}{
		{
			name:        "empty password",
			roastedPass: []byte{},
			want:        []byte{},
		},
		{
			name:        "single byte password",
			roastedPass: []byte{0xF3},
			want:        []byte{0x00},
		},
		{
			name:        "multiple bytes password",
			roastedPass: []byte{0xF3, 0xB3, 0x6C, 0x99},
			want:        []byte{0x00, 0x00, 0x00, 0x00},
		},
		{
			name:        "password longer than roast table",
			roastedPass: []byte{0xF3, 0xB3, 0x6C, 0x99, 0x95, 0x3F, 0xAC, 0xB6, 0xC5, 0xFA, 0x6B, 0x63, 0x69, 0x6C, 0xC3, 0x9A, 0xF3, 0xB3},
			want:        []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		{
			name:        "non-zero roasted password",
			roastedPass: []byte{0xE3, 0xA3, 0x5C, 0x89},
			want:        []byte{0x10, 0x10, 0x30, 0x10},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RoastOSCARJavaPassword(tt.roastedPass)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRoastTOCPassword(t *testing.T) {
	tests := []struct {
		name        string
		roastedPass []byte
		want        []byte
	}{
		{
			name:        "empty password",
			roastedPass: []byte{},
			want:        []byte{},
		},
		{
			name:        "single byte password",
			roastedPass: []byte{0x54}, // 'T'
			want:        []byte{0x00},
		},
		{
			name:        "multiple bytes password",
			roastedPass: []byte{0x54, 0x69, 0x63, 0x2F, 0x54, 0x6F, 0x63}, // "Tic/Toc"
			want:        []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		{
			name:        "password longer than roast table",
			roastedPass: []byte{0x54, 0x69, 0x63, 0x2F, 0x54, 0x6F, 0x63, 0x54, 0x69, 0x63, 0x2F, 0x54, 0x6F, 0x63},
			want:        []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		{
			name:        "non-zero roasted password",
			roastedPass: []byte{0x44, 0x59, 0x53, 0x1F},
			want:        []byte{0x10, 0x30, 0x30, 0x30},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RoastTOCPassword(tt.roastedPass)
			assert.Equal(t, tt.want, got)
		})
	}
}

// Test that roasting and unroasting are reversible operations
func TestRoastPasswordReversibility(t *testing.T) {
	tests := []struct {
		name       string
		clearPass  []byte
		roastTable []byte
		roastFunc  func([]byte) []byte
	}{
		{
			name:       "OSCAR password reversibility",
			clearPass:  []byte("password123"),
			roastTable: []byte{0xF3, 0x26, 0x81, 0xC4, 0x39, 0x86, 0xDB, 0x92, 0x71, 0xA3, 0xB9, 0xE6, 0x53, 0x7A, 0x95, 0x7C},
			roastFunc:  RoastOSCARPassword,
		},
		{
			name:       "Kerberos password reversibility",
			clearPass:  []byte("password123"),
			roastTable: []byte{0x76, 0x91, 0xc5, 0xe7, 0xd0, 0xd9, 0x95, 0xdd, 0x9e, 0x2F, 0xea, 0xd8, 0x6B, 0x21, 0xc2, 0xbc},
			roastFunc:  RoastKerberosPassword,
		},
		{
			name:       "OSCAR Java password reversibility",
			clearPass:  []byte("password123"),
			roastTable: []byte{0xF3, 0xB3, 0x6C, 0x99, 0x95, 0x3F, 0xAC, 0xB6, 0xC5, 0xFA, 0x6B, 0x63, 0x69, 0x6C, 0xC3, 0x9A},
			roastFunc:  RoastOSCARJavaPassword,
		},
		{
			name:       "TOC password reversibility",
			clearPass:  []byte("password123"),
			roastTable: []byte("Tic/Toc"),
			roastFunc:  RoastTOCPassword,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// First, roast the clear password
			roasted := roastPass(tt.clearPass, tt.roastTable)

			// Then, unroast it using the roast function
			unroasted := tt.roastFunc(roasted)

			// The result should be the original clear password
			assert.Equal(t, tt.clearPass, unroasted)
		})
	}
}
