package otp_test

import (
	"testing"

	"github.com/penndev/gopkg/otp"
)

func TestGenerateOTPSecret(t *testing.T) {
	s, _ := otp.GenerateSecret()
	if len(s) != 16 {
		t.Logf("%s", s)
	}
}

func ExampleGenerateOTPSecret() {
	// s := otp.GenerateOTPSecret()
	// fmt.Println(s)
	// LR3PUVPAWSWYGVFX
	// ditigs, _ := otp.GenerateOTPWithTime("LR3PUVPAWSWYGVFX", time.Now())
	// fmt.Println(ditigs)
	// Output:
	// 123456
}
