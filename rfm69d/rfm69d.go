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

	quit := rfm.Loop()

	sigint := make(chan os.Signal)
	signal.Notify(sigint, os.Interrupt)

	<-sigint

	quit <- 1
	<-quit
}
