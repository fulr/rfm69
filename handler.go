package rfm69

import (
	"log"

	"github.com/davecheney/gpio"
)

// Loop is the main receive and transmit handling loop
func (r *Device) Loop() (chan *Data, chan bool) {
	quit := make(chan bool)
	ch := make(chan *Data, 5)
	go r.loopInternal(ch, quit)
	return ch, quit
}

func (r *Device) loopInternal(ch chan *Data, quit chan bool) {
	irq := make(chan int)
	r.gpio.BeginWatch(gpio.EdgeRising, func() {
		irq <- 1
	})
	defer r.gpio.EndWatch()

	err := r.SetMode(RF_OPMODE_RECEIVER)
	if err != nil {
		log.Print(err)
		return
	}
	defer r.SetMode(RF_OPMODE_STANDBY)

	for {
		select {
		case dataToTransmit := <-ch:
			// TODO: can send?
			r.readWriteReg(REG_PACKETCONFIG2, 0xFB, RF_PACKET2_RXRESTART) // avoid RX deadlocks
			err = r.SetModeAndWait(RF_OPMODE_STANDBY)
			if err != nil {
				log.Fatal(err)
			}
			err = r.writeReg(REG_DIOMAPPING1, RF_DIOMAPPING1_DIO0_00)
			if err != nil {
				log.Fatal(err)
			}
			err = r.writeFifo(dataToTransmit)
			if err != nil {
				log.Fatal(err)
			}
			err = r.SetMode(RF_OPMODE_TRANSMITTER)
			if err != nil {
				log.Fatal(err)
			}

			<-irq

			err = r.SetModeAndWait(RF_OPMODE_STANDBY)
			if err != nil {
				log.Fatal(err)
			}
			err = r.writeReg(REG_DIOMAPPING1, RF_DIOMAPPING1_DIO0_01)
			if err != nil {
				log.Fatal(err)
			}
			err = r.SetMode(RF_OPMODE_RECEIVER)
			if err != nil {
				log.Fatal(err)
			}
		case <-irq:
			if r.mode != RF_OPMODE_RECEIVER {
				continue
			}
			flags, err := r.readReg(REG_IRQFLAGS2)
			if err != nil {
				return
			}
			if flags&RF_IRQFLAGS2_PAYLOADREADY == 0 {
				continue
			}
			data, err := r.readFifo()
			if err != nil {
				log.Print(err)
				return
			}
			ch <- &data
			err = r.SetMode(RF_OPMODE_RECEIVER)
			if err != nil {
				log.Fatal(err)
			}
		case <-quit:
			quit <- true
			return
		}
	}
}
