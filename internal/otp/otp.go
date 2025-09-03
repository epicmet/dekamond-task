package otp

import (
	"math/rand/v2"
	"strconv"
	"strings"
)

type OTPProvider interface {
	Send(pn string) error
	Check(pn string, otp string) bool
}

type BaseOTPProvider struct {
	stateManager OTPStateManager
}

// TODO: No repeat int?
func (b *BaseOTPProvider) createRandomInt(len int) string {
	res := make([]string, len)

	for range len {
		i := rand.IntN(10)
		res = append(res, strconv.Itoa(i))
	}

	return strings.Join(res[:], "")
}
