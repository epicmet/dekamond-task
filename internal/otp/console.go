package otp

import "io"

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

	// TODO: Log a better structured line. Also add \n
	_, err = c.output.Write([]byte(otp))

	return err
}

func (c *ConsoleOTP) Check(pn string, otp string) bool {
	storedOtp, err := c.base.stateManager.Get(pn)
	if err != nil || storedOtp != otp {
		return false
	}

	return true
}
