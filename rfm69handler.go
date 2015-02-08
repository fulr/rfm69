package rfm69

import (
	"log"

	"github.com/davecheney/gpio"
)

// Loop is the main receive and transmit handling loop
func (r *Device) Loop() (chan Data, chan Data, chan int) {
	quit := make(chan int)
	txChan := make(chan Data, 5)
	rxChan := make(chan Data, 5)

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
			case dataToTransmit := <-txChan:
				// can send?
				err = r.SetMode(RF_OPMODE_STANDBY)
				if err != nil {
					log.Fatal(err)
				}

				err = r.waitForMode()
				if err != nil {
					log.Fatal(err)
				}

				err = r.writeFifo(&dataToTransmit)
				if err != nil {
					log.Fatal(err)
				}

				log.Print("transmit")
				log.Print(dataToTransmit)

				err = r.writeReg(REG_DIOMAPPING1, RF_DIOMAPPING1_DIO0_00)
				if err != nil {
					log.Fatal(err)
				}

				err = r.SetMode(RF_OPMODE_TRANSMITTER)
				if err != nil {
					log.Fatal(err)
				}

				<-irq
				log.Print("transmit done")

				err = r.SetMode(RF_OPMODE_STANDBY)
				if err != nil {
					log.Fatal(err)
				}

				err = r.waitForMode()
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

				log.Print("receive")
				log.Print(data)

				if data.ToAddress != 255 && data.RequestAck {
					resp := Data{
						FromAddress: r.address,
						ToAddress:   data.FromAddress,
						Data:        data.Data,
						SendAck:     true,
					}
					txChan <- resp
				}

				rxChan <- data

				err = r.SetMode(RF_OPMODE_RECEIVER)
				if err != nil {
					log.Fatal(err)
				}

			case <-quit:
				quit <- 1
				return
			}
		}
	}()

	return rxChan, txChan, quit
}
