package rfm69

/*
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <errno.h>
#include <fcntl.h>
#include <sys/ioctl.h>
#include <unistd.h>
#include <linux/types.h>
#include <linux/spi/spidev.h>

#define SPI_SPEED 4000000

uint8_t mode=0;
uint8_t bits=8;
uint32_t speed=SPI_SPEED;
uint16_t delay=5;

int spi_open(const char *device) {
                int fd = open(device, O_RDWR);
                int ret;

                if (fd < 0) {
                                printf("can't open device");
                                return -1;
                }
                ret = ioctl(fd, SPI_IOC_WR_MODE, &mode);
                if (ret == -1) {
                                printf("can't set spi mode");
                                return -1;
                }

                ret = ioctl(fd, SPI_IOC_RD_MODE, &mode);
                if (ret == -1) {
                                printf("can't get spi mode");
                                return -1;
                }

                ret = ioctl(fd, SPI_IOC_WR_BITS_PER_WORD, &bits);
                if (ret == -1) {
                                printf("can't set bits per word");
                                return -1;
                }

                ret = ioctl(fd, SPI_IOC_RD_BITS_PER_WORD, &bits);
                if (ret == -1) {
                                printf("can't get bits per word");
                                return -1;
                }

                ret = ioctl(fd, SPI_IOC_WR_MAX_SPEED_HZ, &speed);
                if (ret == -1) {
                                printf("can't set max speed hz");
                                return -1;
                }

                ret = ioctl(fd, SPI_IOC_RD_MAX_SPEED_HZ, &speed);
                if (ret == -1) {
                                printf("can't get max speed hz");
                                return -1;
                }

                return fd;
}

int spi_xfer(int fd, char* tx, char* rx, int length) {
                struct spi_ioc_transfer tr = {
                                .tx_buf = (unsigned long)tx,
                                .rx_buf = (unsigned long)rx,
                                .len = length,
                                .delay_usecs = delay,
                                .speed_hz = speed,
                                .bits_per_word = bits,
                };

                int ret = ioctl(fd, SPI_IOC_MESSAGE(1), &tr);
                if (ret < 1)
                                return -1;

                return 0;
}
*/
import "C"
import "unsafe"

import "errors"

// SPIDevice device
type SPIDevice struct {
	fd C.int
}

// NewSPIDevice creates a new device
func NewSPIDevice() (*SPIDevice, error) {
	name := C.CString("/dev/spidev0.0")
	defer C.free(unsafe.Pointer(name))
	i := C.spi_open(name)
	if i < 0 {
		return nil, errors.New("could not open")
	}
	return &SPIDevice{i}, nil
}

// Xfer cross transfer
func (d *SPIDevice) Xfer(tx []byte) ([]byte, error) {
	length := len(tx)
	rx := make([]byte, length)
	//log.Print("sending", tx)
	ret := C.spi_xfer(d.fd, (*C.char)(unsafe.Pointer(&tx[0])), (*C.char)(unsafe.Pointer(&rx[0])), C.int(length))
	//log.Print("got", rx)
	if ret < 0 {
		return nil, errors.New("could not xfer")
	}
	return rx, nil
}

// Close closes the fd
func (d *SPIDevice) Close() {
	C.close(d.fd)
}
