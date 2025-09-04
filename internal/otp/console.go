package otp

import (
	"fmt"
	"io"
)

type ConsoleOTP struct {
	base   *BaseOTPProvider
	output io.Writer
	len    int
}

func NewConsoleOTP(sm OTPStateManager, output io.Writer, otpLen int) *ConsoleOTP {
	return &ConsoleOTP{
		base:   &BaseOTPProvider{stateManager: sm},
		len:    otpLen,
		output: output,
	}
}

func (c *ConsoleOTP) Send(pn string) error {
	otp := c.base.createRandomInt(c.len)
	err := c.base.stateManager.SetX(pn, otp)

	if err != nil {
		return err
	}

	line := fmt.Sprintf("Sending OTP :: { PhoneNumber = %s, OTP = %s }\n", pn, otp)
	_, err = c.output.Write([]byte(line))

	return err
}

func (c *ConsoleOTP) Check(pn string, otp string) bool {
	storedOtp, err := c.base.stateManager.Get(pn)
	if err != nil || storedOtp != otp {
		return false
	}

	return true
}
