package state

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mk6i/retro-aim-server/wire"
)

func TestUser_HashPassword(t *testing.T) {
	tests := []struct {
		name              string
		user              User
		password          string
		expectedWeakMD5   []byte
		expectedStrongMD5 []byte
		wantError         bool
	}{
		{
			name:              "Valid password",
			user:              User{AuthKey: "someAuthKey"},
			password:          "validPassword",
			expectedWeakMD5:   wire.WeakMD5PasswordHash("validPassword", "someAuthKey"),
			expectedStrongMD5: wire.StrongMD5PasswordHash("validPassword", "someAuthKey"),
			wantError:         false,
		},
		{
			name:              "Empty password",
			user:              User{AuthKey: "someAuthKey"},
			password:          "",
			expectedWeakMD5:   nil,
			expectedStrongMD5: nil,
			wantError:         true,
		},
		{
			name:              "Password too short",
			user:              User{AuthKey: "someAuthKey"},
			password:          "abc",
			expectedWeakMD5:   nil,
			expectedStrongMD5: nil,
			wantError:         true,
		},
		{
			name:              "Password too long",
			user:              User{AuthKey: "someAuthKey"},
			password:          "thispasswordistoolong",
			expectedWeakMD5:   nil,
			expectedStrongMD5: nil,
			wantError:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.user.HashPassword(tt.password)

			if tt.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedWeakMD5, tt.user.WeakMD5Pass)
				assert.Equal(t, tt.expectedStrongMD5, tt.user.StrongMD5Pass)
			}
		})
	}
}

func TestDisplayScreenName_ValidateAIMHandle(t *testing.T) {
	tests := []struct {
		name    string
		input   DisplayScreenName
		wantErr error
	}{
		{"Valid handle", "User123", nil},
		{"Too short", "Us", ErrAIMHandleLength},
		{"Too long", "ThisIsAReallyLongScreenName", ErrAIMHandleLength},
		{"Starts with number", "1User", ErrAIMHandleInvalidFormat},
		{"Ends with space", "User123 ", ErrAIMHandleInvalidFormat},
		{"Contains invalid character", "User@123", ErrAIMHandleInvalidFormat},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.ValidateAIMHandle()
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr, "ValidateAIMHandle() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				assert.NoError(t, err, "ValidateAIMHandle() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDisplayScreenName_ValidateICQHandle(t *testing.T) {
	tests := []struct {
		name    string
		input   DisplayScreenName
		wantErr error
	}{
		{"Valid UIN", "123456", nil},
		{"Too low", "9999", ErrICQUINInvalidFormat},
		{"Too high", "2147483647", ErrICQUINInvalidFormat},
		{"Non-numeric", "abcd", ErrICQUINInvalidFormat},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.ValidateICQHandle()
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr, "ValidateICQHandle() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				assert.NoError(t, err, "ValidateICQHandle() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
