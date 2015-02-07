package rfm69

import (
	"log"

	"github.com/davecheney/gpio"
)

// Loop is the main receive and transmit handling loop
func (r *Device) Loop() chan int {
	quit := make(chan int)
	c := make(chan Data, 5)
	go func() {
		irq := make(chan int)

		r.gpio.BeginWatch(gpio.EdgeRising, func() {
			log.Print("irq")
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
				r.waitForMode()
				r.writeFifo(&dataToTransmit)
				log.Print("transmit")
				log.Print(dataToTransmit)
				r.SetMode(RF_OPMODE_TRANSMITTER)
				r.writeReg(REG_DIOMAPPING1, RF_DIOMAPPING1_DIO0_00)
				<-irq
				log.Print("transmit done")
				r.SetMode(RF_OPMODE_RECEIVER)
				r.writeReg(REG_DIOMAPPING1, RF_DIOMAPPING1_DIO0_01)
			case <-irq:
				data, err := r.readFifo()
				if err != nil {
					log.Print(err)
					return
				}
				log.Print(data)
				if data.ToAddress != 255 && data.RequestAck {
					resp := Data{
						FromAddress: r.address,
						ToAddress:   data.FromAddress,
						SendAck:     true,
					}
					c <- resp
				}
			case <-quit:
				quit <- 1
				return
			}
		}
	}()
	return quit
}
