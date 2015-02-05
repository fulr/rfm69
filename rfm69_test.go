package rfm69

import (
	"testing"

	"github.com/kidoman/embd"

	_ "github.com/kidoman/embd/host/rpi"
)

func TestRfm69(t *testing.T) {
	t.Log("Test")
	if err := embd.InitGPIO(); err != nil {
		panic(err)
	}
	defer embd.CloseGPIO()

	gpio, err := embd.NewDigitalPin(25)
	if err != nil {
		panic(err)
	}
	defer gpio.Close()

	if err := gpio.SetDirection(embd.In); err != nil {
		panic(err)
	}
	gpio.ActiveLow(false)

	spiBus, err := NewSPIDevice()
	if err != nil {
		panic(err)
	}
	defer spiBus.Close()

	rfm, err := NewDevice(spiBus, gpio, 1, 10, true)
	if err != nil {
		t.Error(err)
	}
	t.Log(rfm)
}
