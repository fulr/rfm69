// +build !linux

package rfm69

// SPIDevice is a device
type SPIDevice struct {
}

// NewSPIDevice creates a new device
func NewSPIDevice() (*SPIDevice, error) {
	return &SPIDevice{}, nil
}

// Xfer cross transfer
func (d *SPIDevice) Xfer(tx []byte) ([]byte, error) {
	length := len(tx)
	rx := make([]byte, length)
	return rx, nil
}

// Close closes the fd
func (d *SPIDevice) Close() {
}
