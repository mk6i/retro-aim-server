package state

import (
	"bytes"
	"errors"
	"strings"

	"github.com/google/uuid"

	"github.com/mk6i/retro-aim-server/wire"
)

// BlockedState represents the blocked status between two users
type BlockedState int

var (
	// ErrDupUser indicates that a user already exists.
	ErrDupUser = errors.New("user already exists")
	// ErrNoUser indicates that a user does not exist.
	ErrNoUser = errors.New("user does not exist")
)

const (
	// BlockedNo indicates that neither user blocks the other.
	BlockedNo BlockedState = iota
	// BlockedA indicates that user A blocks user B.
	BlockedA
	// BlockedB indicates that user B blocks user A.
	BlockedB
)

// IdentScreenName struct stores the normalized version of a user's screen name.
// This format is used for uniformity in storage and comparison by removing spaces
// and converting all characters to lowercase.
type IdentScreenName struct {
	// screenName contains the identifier screen name value. Do not assign this
	// value directly. Rather, set it through NewIdentScreenName. This ensures
	// that when an instance of IdentScreenName is present, it's guaranteed to
	// have a normalized value.
	screenName string
}

// String returns the string representation of the IdentScreenName.
func (i IdentScreenName) String() string {
	return i.screenName
}

// NewIdentScreenName creates a new IdentScreenName.
func NewIdentScreenName(screenName string) IdentScreenName {
	str := strings.ReplaceAll(screenName, " ", "")
	str = strings.ToLower(str)
	return IdentScreenName{screenName: str}
}

// DisplayScreenName type represents the screen name in the user-defined format.
// This includes the original casing and spacing as defined by the user.
type DisplayScreenName string

// IdentScreenName converts the DisplayScreenName to an IdentScreenName by applying
// the normalization process defined in NewIdentScreenName.
func (s DisplayScreenName) IdentScreenName() IdentScreenName {
	return NewIdentScreenName(string(s))
}

// String returns the original display string of the screen name, preserving the user-defined
// casing and spaces.
func (s DisplayScreenName) String() string {
	return string(s)
}

// NewStubUser creates a new user with canned credentials. The default password
// is "welcome1". This is typically used for development purposes.
func NewStubUser(screenName DisplayScreenName) (User, error) {
	uid, err := uuid.NewRandom()
	if err != nil {
		return User{}, err
	}
	u := User{
		IdentScreenName:   NewIdentScreenName(string(screenName)),
		DisplayScreenName: screenName,
		AuthKey:           uid.String(),
	}
	err = u.HashPassword("welcome1")
	return u, err
}

// User represents a user account.
type User struct {
	// IdentScreenName is the AIM screen name.
	IdentScreenName IdentScreenName
	// DisplayScreenName is the formatted screen name.
	DisplayScreenName DisplayScreenName
	// AuthKey is the salt for the MD5 password hash.
	AuthKey string
	// StrongMD5Pass is the MD5 password hash format used by AIM v4.8-v5.9.
	StrongMD5Pass []byte
	// WeakMD5Pass is the MD5 password hash format used by AIM v3.5-v4.7. This
	// hash is used to authenticate roasted passwords for AIM v1.0-v3.0.
	WeakMD5Pass []byte
}

// ValidateHash checks if md5Hash is identical to one of the password hashes.
func (u *User) ValidateHash(md5Hash []byte) bool {
	return bytes.Equal(u.StrongMD5Pass, md5Hash) || bytes.Equal(u.WeakMD5Pass, md5Hash)
}

// ValidateRoastedPass checks if the provided roasted password matches the MD5
// hash of the user's actual password. A roasted password is a XOR-obfuscated
// form of the real password, intended to add a simple layer of security.
func (u *User) ValidateRoastedPass(roastedPass []byte) bool {
	clearPass := wire.RoastPassword(roastedPass)
	md5Hash := wire.WeakMD5PasswordHash(string(clearPass), u.AuthKey) // todo remove string conversion
	return bytes.Equal(u.WeakMD5Pass, md5Hash)
}

// HashPassword computes MD5 hashes of the user's password. It computes both
// weak and strong variants and stores them in the struct.
func (u *User) HashPassword(passwd string) error {
	u.WeakMD5Pass = wire.WeakMD5PasswordHash(passwd, u.AuthKey)
	u.StrongMD5Pass = wire.StrongMD5PasswordHash(passwd, u.AuthKey)
	return nil
}
