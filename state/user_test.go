package state

import (
	"testing"
	"time"

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
			name:              "Valid AIM password",
			user:              User{AuthKey: "someAuthKey", IsICQ: false},
			password:          "validPassword",
			expectedWeakMD5:   wire.WeakMD5PasswordHash("validPassword", "someAuthKey"),
			expectedStrongMD5: wire.StrongMD5PasswordHash("validPassword", "someAuthKey"),
			wantError:         false,
		},
		{
			name:              "Empty AIM password",
			user:              User{AuthKey: "someAuthKey", IsICQ: false},
			password:          "",
			expectedWeakMD5:   nil,
			expectedStrongMD5: nil,
			wantError:         true,
		},
		{
			name:              "AIM password too short",
			user:              User{AuthKey: "someAuthKey", IsICQ: false},
			password:          "abc",
			expectedWeakMD5:   nil,
			expectedStrongMD5: nil,
			wantError:         true,
		},
		{
			name:              "AIM password too long",
			user:              User{AuthKey: "someAuthKey", IsICQ: false},
			password:          "thispasswordistoolong",
			expectedWeakMD5:   nil,
			expectedStrongMD5: nil,
			wantError:         true,
		},
		{
			name:              "Valid ICQ password",
			user:              User{AuthKey: "someAuthKey", IsICQ: true},
			password:          "validICQ",
			expectedWeakMD5:   wire.WeakMD5PasswordHash("validICQ", "someAuthKey"),
			expectedStrongMD5: wire.StrongMD5PasswordHash("validICQ", "someAuthKey"),
			wantError:         false,
		},
		{
			name:              "Empty ICQ password",
			user:              User{AuthKey: "someAuthKey", IsICQ: true},
			password:          "",
			expectedWeakMD5:   nil,
			expectedStrongMD5: nil,
			wantError:         true,
		},
		{
			name:              "ICQ password too long",
			user:              User{AuthKey: "someAuthKey", IsICQ: true},
			password:          "icqpass89",
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

func TestAge(t *testing.T) {
	tests := []struct {
		name        string
		user        User
		timeNow     func() time.Time
		expectedAge uint16
	}{
		{
			name: "Valid birthday, only year is set",
			user: User{
				ICQMoreInfo: ICQMoreInfo{
					BirthYear: 1990,
				},
			},
			timeNow: func() time.Time {
				return time.Date(2024, 8, 1, 0, 0, 0, 0, time.UTC)
			},
			expectedAge: 34,
		},
		{
			name: "Valid birthday, birthday passed this year",
			user: User{
				ICQMoreInfo: ICQMoreInfo{
					BirthYear:  1990,
					BirthMonth: 5,
					BirthDay:   10,
				},
			},
			timeNow: func() time.Time {
				return time.Date(2024, 8, 1, 0, 0, 0, 0, time.UTC)
			},
			expectedAge: 34,
		},
		{
			name: "Valid birthday, birthday not yet passed this year",
			user: User{
				ICQMoreInfo: ICQMoreInfo{
					BirthYear:  1990,
					BirthMonth: 12,
					BirthDay:   10,
				},
			},
			timeNow: func() time.Time {
				return time.Date(2024, 8, 1, 0, 0, 0, 0, time.UTC)
			},
			expectedAge: 33,
		},
		{
			name: "Birthday is today",
			user: User{
				ICQMoreInfo: ICQMoreInfo{
					BirthYear:  1990,
					BirthMonth: 8,
					BirthDay:   1,
				},
			},
			timeNow: func() time.Time {
				return time.Date(2024, 8, 1, 0, 0, 0, 0, time.UTC)
			},
			expectedAge: 34,
		},
		{
			name: "Invalid birthday, year is zero",
			user: User{
				ICQMoreInfo: ICQMoreInfo{
					BirthYear:  0,
					BirthMonth: 8,
					BirthDay:   1,
				},
			},
			timeNow: func() time.Time {
				return time.Date(2024, 8, 1, 0, 0, 0, 0, time.UTC)
			},
			expectedAge: 0,
		},
		{
			name: "Invalid birthday, day is zero",
			user: User{
				ICQMoreInfo: ICQMoreInfo{
					BirthYear:  1990,
					BirthMonth: 8,
					BirthDay:   0,
				},
			},
			timeNow: func() time.Time {
				return time.Date(2024, 8, 1, 0, 0, 0, 0, time.UTC)
			},
			expectedAge: 0,
		},
		{
			name: "Invalid birthday, month is zero",
			user: User{
				ICQMoreInfo: ICQMoreInfo{
					BirthYear:  1990,
					BirthMonth: 0,
					BirthDay:   1,
				},
			},
			timeNow: func() time.Time {
				return time.Date(2024, 8, 1, 0, 0, 0, 0, time.UTC)
			},
			expectedAge: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			age := tt.user.Age(tt.timeNow)
			if age != tt.expectedAge {
				t.Errorf("expected age %d, got %d", tt.expectedAge, age)
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
		{"Valid handle no spaces", "User123", nil},
		{"Valid handle with min character count and space", "U SR", nil},
		{"Valid handle with max character count", "JustTheRightSize", nil},
		{"Valid handle with max character count and spaces", "Just   RightSize", nil},
		{"Too short", "Us", ErrAIMHandleLength},
		{"Too short due to spaces", "U S", ErrAIMHandleLength},
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
			err := tt.input.ValidateUIN()
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr, "ValidateUIN() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				assert.NoError(t, err, "ValidateUIN() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
