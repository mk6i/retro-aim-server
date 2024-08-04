package wire

import (
	"errors"
	"fmt"
	"testing"
)

func TestValidateAIMHandle(t *testing.T) {
	tests := []struct {
		screenName string
		expected   error
	}{
		{"ValidName", nil},
		{"Valid Name", nil},
		{"Val1dName", nil},
		{"Va", errors.New("screen name must be between 3 and 16 characters")},
		{"ThisIsAVeryLongScreenName", errors.New("screen name must be between 3 and 16 characters")},
		{"1Invalid", errors.New("screen name must start with a letter")},
		{"Invalid ", errors.New("screen name cannot end with a space")},
		{"Inval!dName", errors.New("screen name must contain only letters, numbers, and spaces")},
	}

	for _, tt := range tests {
		t.Run(tt.screenName, func(t *testing.T) {
			err := ValidateAIMHandle(tt.screenName)
			if err != nil && err.Error() != tt.expected.Error() {
				t.Errorf("got %v, want %v", err, tt.expected)
			}
			if err == nil && tt.expected != nil {
				t.Errorf("got %v, want %v", err, tt.expected)
			}
		})
	}
}

func TestValidateAIMPassword(t *testing.T) {
	tests := []struct {
		password string
		expected error
	}{
		{"pass1234", nil},
		{"1234", nil},
		{"pa", errors.New("password must be between 4 and 16 characters")},
		{"thisisaverylongpassword", errors.New("password must be between 4 and 16 characters")},
	}

	for _, tt := range tests {
		t.Run(tt.password, func(t *testing.T) {
			err := ValidateAIMPassword(tt.password)
			if err != nil && err.Error() != tt.expected.Error() {
				t.Errorf("got %v, want %v", err, tt.expected)
			}
			if err == nil && tt.expected != nil {
				t.Errorf("got %v, want %v", err, tt.expected)
			}
		})
	}
}

func TestValidateICQHandle(t *testing.T) {
	tests := []struct {
		UIN      uint32
		expected error
	}{
		{123456, nil},
		{10000, nil},
		{2147483646, nil},
		{9999, errors.New("UIN must be between 10000 and 2147483646")},
		{2147483647, errors.New("UIN must be between 10000 and 2147483646")},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%d", tt.UIN), func(t *testing.T) {
			err := ValidateICQHandle(tt.UIN)
			if err != nil && err.Error() != tt.expected.Error() {
				t.Errorf("got %v, want %v", err, tt.expected)
			}
			if err == nil && tt.expected != nil {
				t.Errorf("got %v, want %v", err, tt.expected)
			}
		})
	}
}

func TestValidateICQPassword(t *testing.T) {
	tests := []struct {
		password string
		expected error
	}{
		{"pass", nil},
		{"12345678", nil},
		{"", errors.New("password must be between 1 and 8 characters")},
		{"123456789", errors.New("password must be between 1 and 8 characters")},
	}

	for _, tt := range tests {
		t.Run(tt.password, func(t *testing.T) {
			err := ValidateICQPassword(tt.password)
			if err != nil && err.Error() != tt.expected.Error() {
				t.Errorf("got %v, want %v", err, tt.expected)
			}
			if err == nil && tt.expected != nil {
				t.Errorf("got %v, want %v", err, tt.expected)
			}
		})
	}
}
