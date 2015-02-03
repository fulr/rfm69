package rfm69

import (
	"flag"
	"testing"

	"github.com/kidoman/embd"
	_ "github.com/kidoman/embd/host/rpi"
)

func TestRfm69(t *testing.T) {
	flag.Set("v", "23")

	t.Log("Test")
	if err := embd.InitSPI(); err != nil {
		t.Error(err)
	}
	defer embd.CloseSPI()

	if err := embd.InitGPIO(); err != nil {
		panic(err)
	}
	defer embd.CloseGPIO()

	gpio, err := embd.NewDigitalPin(10)
	if err != nil {
		panic(err)
	}
	defer gpio.Close()

	//if err := gpio.SetDirection(embd.In); err != nil {
	//	panic(err)
	//}
	//gpio.ActiveLow(false)

	spiBus := embd.NewSPIBus(embd.SPIMode0, 0, 4000000, 8, 0)
	defer spiBus.Close()

	rfm, err := NewDevice(spiBus, gpio, 1, 10, true)
	if err != nil {
		t.Error(err)
	}
	t.Log(rfm)
}
