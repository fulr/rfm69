package rfm69

import (
	"log"

	"github.com/davecheney/gpio"
)

const (
	irqPin = gpio.GPIO25
)

func main() {
	log.Print("Start")

	pin, err := gpio.OpenPin(irqPin, gpio.ModeInput)
	if err != nil {
		panic(err)
	}
	defer pin.Close()

	spiBus, err := NewSPIDevice("/dev/spidev0.0")
	if err != nil {
		panic(err)
	}
	defer spiBus.Close()

	rfm, err := NewDevice(spiBus, pin, 1, 10, false)
	if err != nil {
		log.Fatal(err)
	}
	log.Print(rfm)
}
