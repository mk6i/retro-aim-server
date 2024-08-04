package wire

import (
	"crypto/md5"
	"errors"
	"io"
	"unicode"
)

// WeakMD5PasswordHash hashes password and authKey for AIM v3.5-v4.7.
//
//goland:noinspection ALL
func WeakMD5PasswordHash(pass, authKey string) []byte {
	hash := md5.New()
	io.WriteString(hash, authKey)
	io.WriteString(hash, pass)
	io.WriteString(hash, "AOL Instant Messenger (SM)")
	return hash.Sum(nil)
}

// StrongMD5PasswordHash hashes password and authKey for AIM v4.8+.
//
//goland:noinspection ALL
func StrongMD5PasswordHash(pass, authKey string) []byte {
	top := md5.New()
	io.WriteString(top, pass)
	bottom := md5.New()
	io.WriteString(bottom, authKey)
	bottom.Write(top.Sum(nil))
	io.WriteString(bottom, "AOL Instant Messenger (SM)")
	return bottom.Sum(nil)
}

// RoastPassword toggles password obfuscation using a roasting algorithm for
// AIM v1.0-v3.0 auth. The first call obfuscates the password, and the second
// call de-obfuscates the password, and so on.
func RoastPassword(roastedPass []byte) []byte {
	var roastTable = [16]byte{
		0xF3, 0x26, 0x81, 0xC4, 0x39, 0x86, 0xDB, 0x92,
		0x71, 0xA3, 0xB9, 0xE6, 0x53, 0x7A, 0x95, 0x7C,
	}
	clearPass := make([]byte, len(roastedPass))
	for i := range roastedPass {
		clearPass[i] = roastedPass[i] ^ roastTable[i%len(roastTable)]
	}
	return clearPass
}

// ValidateAIMHandle returns an error if the AIM screen name is invalid.
// A valid screen name meets the following criteria:
//   - 3-16 letters/numbers
//   - must start with a letter
//   - doesn't end with a space
func ValidateAIMHandle(screenName string) error {
	if len(screenName) < 3 || len(screenName) > 16 {
		return errors.New("screen name must be between 3 and 16 characters")
	}

	// Must start with a letter
	if !unicode.IsLetter(rune(screenName[0])) {
		return errors.New("screen name must start with a letter")
	}

	// Cannot end with a space
	if screenName[len(screenName)-1] == ' ' {
		return errors.New("screen name cannot end with a space")
	}

	// Must contain only letters, numbers, and spaces
	for _, ch := range screenName {
		if !unicode.IsLetter(ch) && !unicode.IsDigit(ch) && ch != ' ' {
			return errors.New("screen name must contain only letters, numbers, and spaces")
		}
	}

	return nil
}

// ValidateAIMPassword returns an error if the AIM password is invalid.
// A valid password is 4-16 characters long. The minimum password length is
// set here for software preservation purposes; operators should set more
// stringent password requirements.
func ValidateAIMPassword(pass string) error {
	if len(pass) < 4 || len(pass) > 16 {
		return errors.New("password must be between 4 and 16 characters")
	}
	return nil
}

// ValidateICQHandle returns an error if the ICQ UIN is invalid.
// A valid UIN is a number in the range 10000-2147483646.
func ValidateICQHandle(UIN uint32) error {
	if UIN < 10000 || UIN > 2147483646 {
		return errors.New("UIN must be between 10000 and 2147483646")
	}
	return nil
}

// ValidateICQPassword returns an error if the ICQ password is invalid.
// A valid password is 1-8 characters long. The minimum password length is set
// here for software preservation purposes; operators should set more stringent
// password requirements.
func ValidateICQPassword(pass string) error {
	if len(pass) < 1 || len(pass) > 8 {
		return errors.New("password must be between 1 and 8 characters")
	}
	return nil
}
