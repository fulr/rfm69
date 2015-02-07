package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/fulr/rfm69"
	"github.com/kidoman/embd"

	_ "github.com/kidoman/embd/host/rpi"
)

func main() {
	log.Print("Start")

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

	spiBus, err := rfm69.NewSPIDevice()
	if err != nil {
		panic(err)
	}
	defer spiBus.Close()

	rfm, err := rfm69.NewDevice(spiBus, gpio, 1, 10, true)
	if err != nil {
		log.Fatal(err)
	}
	log.Print(rfm)

	err = rfm.Encrypt("0123456789012345")
	if err != nil {
		log.Fatal(err)
	}

	quit := rfm.Loop()

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)

	for {
		select {
		case <-sigint:
			quit <- 1
			<-quit
			return
		}
	}
}
