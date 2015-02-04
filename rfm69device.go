// Package rfm69 RFM69 Implementation in Go
package rfm69

import (
	"log"

	"github.com/fulr/embd"
)

// Device RFM69 Device
type Device struct {
	SpiDevice  embd.SPIBus
	gpio       embd.DigitalPin
	mode       byte
	address    byte
	network    byte
	isRFM69HW  bool
	powerLevel byte
}

// Global settings
const (
	CsmaLimit  = -80
	MaxDataLen = 66
)

// NewDevice creates a new device
func NewDevice(spi embd.SPIBus, gpio embd.DigitalPin, nodeID, networkID byte, isRfm69HW bool) (*Device, error) {
	ret := &Device{
		SpiDevice: spi,
		gpio:      gpio,
		network:   networkID,
		address:   nodeID,
		isRFM69HW: isRfm69HW,
	}

	log.Println("before setup")
	err := ret.setup()
	log.Println("after setup")

	return ret, err
}

func (r *Device) writeReg(addr, data byte) error {
	tx := []byte{addr | 0x80, data}
	log.Printf("write %x: %x", addr, data)
	err := r.SpiDevice.TransferAndRecieveData(tx)
	if err != nil {
		log.Println(err)
	}
	return err
}

func (r *Device) readReg(addr byte) (byte, error) {
	tx := []byte{addr & 0x7f, 0}
	log.Printf("read %x", addr)
	err := r.SpiDevice.TransferAndRecieveData(tx)
	if err != nil {
		log.Println(err)
	}
	return tx[1], err
}

func (r *Device) setup() error {
	config := [][]byte{
		/* 0x01 */ {REG_OPMODE, RF_OPMODE_SEQUENCER_ON | RF_OPMODE_LISTEN_OFF | RF_OPMODE_STANDBY},
		/* 0x02 */ {REG_DATAMODUL, RF_DATAMODUL_DATAMODE_PACKET | RF_DATAMODUL_MODULATIONTYPE_FSK | RF_DATAMODUL_MODULATIONSHAPING_00}, // no shaping
		/* 0x03 */ {REG_BITRATEMSB, RF_BITRATEMSB_55555}, // default: 4.8 KBPS
		/* 0x04 */ {REG_BITRATELSB, RF_BITRATELSB_55555},
		/* 0x05 */ {REG_FDEVMSB, RF_FDEVMSB_50000}, // default: 5KHz, (FDEV + BitRate / 2 <= 500KHz)
		/* 0x06 */ {REG_FDEVLSB, RF_FDEVLSB_50000},

		/* 0x07 */ {REG_FRFMSB, RF_FRFMSB_868},
		/* 0x08 */ {REG_FRFMID, RF_FRFMID_868},
		/* 0x09 */ {REG_FRFLSB, RF_FRFLSB_868},

		// looks like PA1 and PA2 are not implemented on RFM69W, hence the max output power is 13dBm
		// +17dBm and +20dBm are possible on RFM69HW
		// +13dBm formula: Pout = -18 + OutputPower (with PA0 or PA1**)
		// +17dBm formula: Pout = -14 + OutputPower (with PA1 and PA2)**
		// +20dBm formula: Pout = -11 + OutputPower (with PA1 and PA2)** and high power PA settings (section 3.3.7 in datasheet)
		///* 0x11 */ { REG_PALEVEL, RF_PALEVEL_PA0_ON | RF_PALEVEL_PA1_OFF | RF_PALEVEL_PA2_OFF | RF_PALEVEL_OUTPUTPOWER_11111},
		///* 0x13 */ { REG_OCP, RF_OCP_ON | RF_OCP_TRIM_95 }, // over current protection (default is 95mA)

		// RXBW defaults are { REG_RXBW, RF_RXBW_DCCFREQ_010 | RF_RXBW_MANT_24 | RF_RXBW_EXP_5} (RxBw: 10.4KHz)
		/* 0x19 */ {REG_RXBW, RF_RXBW_DCCFREQ_010 | RF_RXBW_MANT_16 | RF_RXBW_EXP_2}, // (BitRate < 2 * RxBw)
		//for BR-19200: /* 0x19 */ { REG_RXBW, RF_RXBW_DCCFREQ_010 | RF_RXBW_MANT_24 | RF_RXBW_EXP_3 },
		/* 0x25 */ {REG_DIOMAPPING1, RF_DIOMAPPING1_DIO0_01}, // DIO0 is the only IRQ we're using
		/* 0x26 */ {REG_DIOMAPPING2, RF_DIOMAPPING2_CLKOUT_OFF}, // DIO5 ClkOut disable for power saving
		/* 0x28 */ {REG_IRQFLAGS2, RF_IRQFLAGS2_FIFOOVERRUN}, // writing to this bit ensures that the FIFO & status flags are reset
		/* 0x29 */ {REG_RSSITHRESH, 220}, // must be set to dBm = (-Sensitivity / 2), default is 0xE4 = 228 so -114dBm
		///* 0x2D */ { REG_PREAMBLELSB, RF_PREAMBLESIZE_LSB_VALUE } // default 3 preamble bytes 0xAAAAAA
		/* 0x2E */ {REG_SYNCCONFIG, RF_SYNC_ON | RF_SYNC_FIFOFILL_AUTO | RF_SYNC_SIZE_2 | RF_SYNC_TOL_0},
		/* 0x2F */ {REG_SYNCVALUE1, 0x2D}, // attempt to make this compatible with sync1 byte of RFM12B lib
		/* 0x30 */ {REG_SYNCVALUE2, r.network}, // NETWORK ID
		/* 0x37 */ {REG_PACKETCONFIG1, RF_PACKET1_FORMAT_VARIABLE | RF_PACKET1_DCFREE_OFF | RF_PACKET1_CRC_ON | RF_PACKET1_CRCAUTOCLEAR_ON | RF_PACKET1_ADRSFILTERING_OFF},
		/* 0x38 */ {REG_PAYLOADLENGTH, 66}, // in variable length mode: the max frame size, not used in TX
		///* 0x39 */ { REG_NODEADRS, nodeID }, // turned off because we're not using address filtering
		/* 0x3C */ {REG_FIFOTHRESH, RF_FIFOTHRESH_TXSTART_FIFONOTEMPTY | RF_FIFOTHRESH_VALUE}, // TX on FIFO not empty
		/* 0x3D */ {REG_PACKETCONFIG2, RF_PACKET2_RXRESTARTDELAY_2BITS | RF_PACKET2_AUTORXRESTART_ON | RF_PACKET2_AES_OFF}, // RXRESTARTDELAY must match transmitter PA ramp-down time (bitrate dependent)
		//for BR-19200: /* 0x3D */ { REG_PACKETCONFIG2, RF_PACKET2_RXRESTARTDELAY_NONE | RF_PACKET2_AUTORXRESTART_ON | RF_PACKET2_AES_OFF }, // RXRESTARTDELAY must match transmitter PA ramp-down time (bitrate dependent)
		/* 0x6F */ {REG_TESTDAGC, RF_DAGC_IMPROVED_LOWBETA0}, // run DAGC continuously in RX mode for Fading Margin Improvement, recommended default for AfcLowBetaOn=0
	}

	//digitalWrite(_slaveSelectPin, HIGH)
	//pinMode(_slaveSelectPin, OUTPUT)
	//SPI.begin()

	log.Println("start setup")
	for data, err := r.readReg(REG_SYNCVALUE1); err == nil && data != 0xAA; data, err = r.readReg(REG_SYNCVALUE1) {
		err := r.writeReg(REG_SYNCVALUE1, 0xAA)
		if err != nil {
			return err
		}
	}

	for data, err := r.readReg(REG_SYNCVALUE1); err == nil && data != 0x55; data, err = r.readReg(REG_SYNCVALUE1) {
		r.writeReg(REG_SYNCVALUE1, 0x55)
		if err != nil {
			return err
		}
	}

	for _, c := range config {
		err := r.writeReg(c[0], c[1])
		if err != nil {
			return err
		}
	}

	// Encryption is persistent between resets and can trip you up during debugging.
	// Disable it during initialization so we always start from a known state.
	err := r.encrypt([]byte{})
	if err != nil {
		return err
	}

	// called regardless if it's a RFM69W or RFM69HW
	err = r.setHighPower(r.isRFM69HW)
	if err != nil {
		return err
	}

	err = r.SetMode(RF_OPMODE_STANDBY)
	if err != nil {
		return err
	}
	r.waitForMode()
	//while((readReg(REG_IRQFLAGS1) & RF_IRQFLAGS1_MODEREADY) == 0x00) // wait for ModeReady
	//attachInterrupt(_interruptNum, RFM69::isr0, RISING);

	//selfPointer = this
	//_address = nodeID
	return nil
}

func (r *Device) waitForMode() error {
	for {
		reg, err := r.readReg(REG_IRQFLAGS1)
		if err != nil {
			return err
		}
		if reg&RF_IRQFLAGS1_MODEREADY != 0 {
			break
		}
	}
	return nil
}

func (r *Device) encrypt(key []byte) error {
	var turnOn byte
	if len(key) == 16 {
		turnOn = 1
		tx := make([]byte, 17)
		tx[0] = REG_AESKEY1 | 0x80
		copy(tx[1:], key)
		if err := r.SpiDevice.TransferAndRecieveData(tx); err != nil {
			return err
		}
	}
	return r.readWriteReg(REG_PACKETCONFIG2, 0xFE, turnOn)
}

// SetMode sets operation mode
func (r *Device) SetMode(newMode byte) error {
	if newMode == r.mode {
		return nil
	}

	err := r.readWriteReg(REG_OPMODE, 0xE3, newMode)
	if err != nil {
		return err
	}

	if newMode == RF_OPMODE_RECEIVER || newMode == RF_OPMODE_TRANSMITTER {
		err = r.setHighPowerRegs(newMode == RF_OPMODE_RECEIVER)
		if err != nil {
			return err
		}
	}

	// we are using packet mode, so this check is not really needed
	// but waiting for mode ready is necessary when going from sleep because the FIFO may not be immediately available from previous mode
	if r.mode == RF_OPMODE_SLEEP {
		for {
			data, err := r.readReg(REG_IRQFLAGS1)
			if err != nil {
				return err
			}
			if data&RF_IRQFLAGS1_MODEREADY != 0 {
				break
			}
		}
	}

	r.mode = newMode
	return nil
}

func (r *Device) setHighPower(turnOn bool) error {
	r.isRFM69HW = turnOn

	ocp := byte(RF_OCP_ON)
	if r.isRFM69HW {
		ocp = RF_OCP_OFF
	}

	err := r.writeReg(REG_OCP, ocp)
	if err != nil {
		return err
	}

	if r.isRFM69HW { // turning ON
		// enable P1 & P2 amplifier stages
		err = r.readWriteReg(REG_PALEVEL, 0x1F, RF_PALEVEL_PA1_ON|RF_PALEVEL_PA2_ON)
	} else {
		// enable P0 only
		err = r.readWriteReg(REG_PALEVEL, 0, RF_PALEVEL_PA0_ON|RF_PALEVEL_PA1_OFF|RF_PALEVEL_PA2_OFF|r.powerLevel)
	}

	return err
}

func (r *Device) setHighPowerRegs(turnOn bool) (err error) {
	var (
		testPa1 byte = 0x55
		testPa2 byte = 0x70
	)

	if turnOn {
		testPa1 = 0x5D
		testPa2 = 0x7C
	}
	err = r.writeReg(REG_TESTPA1, testPa1)
	if err != nil {
		return
	}
	err = r.writeReg(REG_TESTPA2, testPa2)
	return
}

// SetNetwork sets the network ID
func (r *Device) SetNetwork(networkID byte) error {
	r.network = networkID
	return r.writeReg(REG_SYNCVALUE2, networkID)
}

// SetAddress sets the node address
func (r *Device) SetAddress(address byte) error {
	r.address = address
	return r.writeReg(REG_NODEADRS, address)
}

// SetPowerLevel sets the TX power
func (r *Device) SetPowerLevel(powerLevel byte) error {
	r.powerLevel = powerLevel
	if r.powerLevel > 31 {
		r.powerLevel = 31
	}
	return r.readWriteReg(REG_PALEVEL, 0xE0, r.powerLevel)
}

func (r *Device) canSend() (bool, error) {
	// if signal stronger than -100dBm is detected assume channel activity
	if r.mode == RF_OPMODE_RECEIVER {
		rssi, err := r.readRSSI(false)
		if err != nil {
			return false, err
		}
		if rssi < CsmaLimit {
			err = r.SetMode(RF_OPMODE_STANDBY)
			return true, err
		}
	}
	return false, nil
}

func (r *Device) readRSSI(forceTrigger bool) (rssi int, err error) {
	if forceTrigger {
		// RSSI trigger not needed if DAGC is in continuous mode
		err = r.writeReg(REG_RSSICONFIG, RF_RSSI_START)
		if err != nil {
			return
		}
		for {
			data, err := r.readReg(REG_RSSICONFIG)
			if err != nil {
				return 0, err
			}
			if data&RF_RSSI_DONE != 0 {
				break
			}
		}
	}
	var data byte
	data, err = r.readReg(REG_RSSIVALUE)
	if err != nil {
		return
	}
	rssi = -int(data) / 2
	return
}

func (r *Device) readWriteReg(reg, andMask, orMask byte) error {
	regValue, err := r.readReg(reg)
	if err != nil {
		return err
	}
	regValue = (regValue & andMask) | orMask
	return r.writeReg(reg, regValue)
}

func (r *Device) writeFifo(toAddress byte, buffer []byte, requestACK, sendACK bool) error {
	buffersize := len(buffer)
	if buffersize > MaxDataLen {
		buffersize = MaxDataLen
	}
	tx := make([]byte, buffersize+5)
	// write to FIFO
	tx[0] = REG_FIFO | 0x80
	tx[1] = byte(buffersize + 3)
	tx[2] = toAddress
	tx[3] = r.address

	if requestACK {
		tx[4] = 0x40
	}
	if sendACK {
		tx[4] = 0x80
	}

	copy(tx[5:], buffer[:buffersize])

	return r.SpiDevice.TransferAndRecieveData(tx)
}
