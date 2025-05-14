package otp_test

import (
	"fmt"
	"time"

	"github.com/penndev/gopkg/otp"
)

func ExampleGenerateOTPSecret() {
	s, _ := otp.GenerateSecret()
	fmt.Println(s)
	u := otp.GenerateOTPURI("totp", "gopkg", "test", s)
	fmt.Println(u)
	// Output:
	// RQ3Z3PE56KYBK2ND
	// otpauth://totp/gopkg:test?secret=RQ3Z3PE56KYBK2ND&issuer=gopkg&algorithm=SHA1&digits=6&period=30
}

func ExampleGenerateOTPWithTime() {
	code, _ := otp.GenerateOTPWithTime("RQ3Z3PE56KYBK2ND", time.Now())
	fmt.Println(code)
	// Output:
	// 767240
}
