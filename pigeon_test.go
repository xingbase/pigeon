package pigeon_test

import (
	"testing"

	"github.com/xingbase/pigeon"
)

func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		name      string
		args      string
		wantValid bool
		wantErr   bool
	}{
		{
			name:      "gmail.com OK",
			args:      "user@Gmail.COM ",
			wantValid: true,
			wantErr:   false,
		},
		{
			name:      "hanmail.ner OK",
			args:      "user@hanmail.ner",
			wantValid: false,
			wantErr:   true,
		},
		// {
		// 	name:      "gamil.com NG",
		// 	args:      "user@gamil.com",
		// 	wantValid: false,
		// 	wantErr:   true,
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call the function with actual DNS lookup
			valid, err := pigeon.IsValidEmail(tt.args)

			// Check results
			if valid != tt.wantValid {
				t.Errorf("IsValidDomain(%q) returned valid=%v, want %v", tt.args, valid, tt.wantValid)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("IsValidDomain(%q) returned error=%v, want error=%v", tt.args, err, tt.wantErr)
			}
		})
	}
}

func TestIsValidDomain(t *testing.T) {
	// Test cases
	tests := []struct {
		name      string
		args      string
		wantValid bool
		wantErr   bool
	}{
		{
			name:      "1111gmail.com",
			args:      "1111gmail.com",
			wantValid: true,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call the function with actual DNS lookup
			valid, err := pigeon.IsValidDomain(tt.args)

			// Check results
			if valid != tt.wantValid {
				t.Errorf("IsValidDomain(%q) returned valid=%v, want %v", tt.args, valid, tt.wantValid)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("IsValidDomain(%q) returned error=%v, want error=%v", tt.args, err, tt.wantErr)
			}
		})
	}
}
