package rfm69

import (
	"log"

	"github.com/kidoman/embd"
)

// Loop is the main receive and transmit handling loop
func (r *Device) Loop() error {
	c := make(chan Data)
	irq := make(chan int)

	r.gpio.Watch(embd.EdgeRising, func(pin embd.DigitalPin) {
		irq <- 1
	})

	r.SetMode(RF_OPMODE_RECEIVER)

	for {
		select {
		case dataToTransmit := <-c:
			// can send?
			r.SetMode(RF_OPMODE_STANDBY)
			r.writeFifo(&dataToTransmit)
			r.writeReg(REG_DIOMAPPING1, RF_DIOMAPPING1_DIO0_00)
			r.SetMode(RF_OPMODE_TRANSMITTER)
			<-irq
			r.SetMode(RF_OPMODE_RECEIVER)
			r.writeReg(REG_DIOMAPPING1, RF_DIOMAPPING1_DIO0_01)
		case <-irq:
			data, err := r.readFifo()
			if err != nil {
				log.Print(err)
				return err
			}
			log.Print(data)
		}
	}
}
