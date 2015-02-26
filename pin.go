// +build !linux

package rfm69

import (
	"errors"

	"github.com/davecheney/gpio"
)

func getPin() (gpio.Pin, error) {
	return nil, errors.New("not implemented")
}
