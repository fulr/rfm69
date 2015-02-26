package rfm69

import "github.com/davecheney/gpio"

const (
	irqPin = gpio.GPIO25
)

func getPin() (gpio.Pin, error) {
	return gpio.OpenPin(irqPin, gpio.ModeInput)
}
