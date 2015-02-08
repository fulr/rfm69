package main

import (
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/davecheney/gpio"
	"github.com/fulr/rfm69"
)

func main() {
	log.Print("Start")

	pin, err := gpio.OpenPin(gpio.GPIO25, gpio.ModeInput)
	if err != nil {
		panic(err)
	}
	defer pin.Close()

	spiBus, err := rfm69.NewSPIDevice()
	if err != nil {
		panic(err)
	}
	defer spiBus.Close()

	rfm, err := rfm69.NewDevice(spiBus, pin, 1, 10, true)
	if err != nil {
		log.Fatal(err)
	}
	log.Print(rfm)

	err = rfm.Encrypt([]byte("0123456789012345"))
	if err != nil {
		log.Fatal(err)
	}

	rxChan, txChan, quit := rfm.Loop()

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt, os.Kill)

	for {
		select {
		case data := <-rxChan:
			log.Print("main got data")
			log.Print(data)
		case <-sigint:
			quit <- 1
			<-quit
			return
		case <-time.After(3 * time.Second):
			txChan <- rfm69.Data{
				ToAddress:   99,
				FromAddress: 1,
				Data:        []byte{1, 2, 3},
				RequestAck:  true,
			}
		}
	}
}
