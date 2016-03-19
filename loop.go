package rfm69

import (
	"log"

	"github.com/davecheney/gpio"
)

// Send data
func (r *Device) Send(d *Data) {
	r.tx <- d
}

func (r *Device) loop() {
	irq := make(chan int)
	r.gpio.BeginWatch(gpio.EdgeRising, func() {
		irq <- 1
	})
	defer r.gpio.EndWatch()

	err := r.SetMode(RF_OPMODE_RECEIVER)
	if err != nil {
		log.Fatal(err)
	}
	defer r.SetMode(RF_OPMODE_STANDBY)

	for {
		select {
		case dataToTransmit := <-r.tx:
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
				log.Fatal(err)
			}
			if flags&RF_IRQFLAGS2_PAYLOADREADY == 0 {
				continue
			}
			data, err := r.readFifo()
			if err != nil {
				log.Fatal(err)
			}
			if r.OnReceive != nil {
				go r.OnReceive(&data)
			}
			err = r.SetMode(RF_OPMODE_RECEIVER)
			if err != nil {
				log.Fatal(err)
			}
		case <-r.quit:
			r.quit <- true
			return
		}
	}
}
