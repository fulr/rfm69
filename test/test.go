package main

import (
	"fmt"

	"github.com/fulr/rfm69"
	"github.com/kidoman/embd"
	_ "github.com/kidoman/embd/host/all"
)

func main() {
	fmt.Print("Test")

	if err := embd.InitSPI(); err != nil {
		panic(err)
	}
	defer embd.CloseSPI()

	spiBus := embd.NewSPIBus(embd.SPIMode0, 0, 4000000, 8, 0)
	defer spiBus.Close()

	rfm, err := rfm69.NewDevice(spiBus, 1, 10, true)
	if err != nil {
		panic(err)
	}
}
