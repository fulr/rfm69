package rfm69

import (
	"log"

	"github.com/davecheney/gpio"
)

// Loop is the main receive and transmit handling loop
func (r *Device) Loop() chan int {
	quit := make(chan int)
	c := make(chan Data)
	go func() {
		irq := make(chan int)

		r.gpio.BeginWatch(gpio.EdgeRising, func() {
			irq <- 1
		})

		err := r.SetMode(RF_OPMODE_RECEIVER)
		if err != nil {
			log.Print(err)
			return
		}
		defer r.SetMode(RF_OPMODE_STANDBY)

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
					return
				}
				log.Print(data)
			case <-quit:
				quit <- 1
				return
			}
		}
	}()
	return quit
}
