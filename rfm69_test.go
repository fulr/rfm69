package rfm69

import (
	"testing"

	"github.com/kidoman/embd"
	_ "github.com/kidoman/embd/host/rpi"
)

func TestRfm69(t *testing.T) {
	t.Log("Test")
	if err := embd.InitSPI(); err != nil {
		t.Error(err)
	}
	defer embd.CloseSPI()

	spiBus := embd.NewSPIBus(embd.SPIMode0, 0, 4000000, 8, 0)
	defer spiBus.Close()

	rfm, err := NewDevice(spiBus, 1, 10, true)
	if err != nil {
		t.Error(err)
	}
	t.Log(rfm)
}
