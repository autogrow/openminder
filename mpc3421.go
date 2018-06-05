package openminder

import (
	"fmt"
	"io"
	"log"
	"math"
	"time"

	"gobot.io/x/gobot/sysfs"
)

type byteReadWriter interface {
	io.Reader
	WriteByte(byte) error
}

// MPC3421VRef is the voltage reference for the MPC3421 chip
const MPC3421VRef = 2.048

// MPC3421Driver is the driver for the MPC3421 chip
type MPC3421Driver struct {
	bus        byteReadWriter
	gain       int
	configByte byte
}

// NewMPC3421 returns a object representing the MPC3421 ADC chip
func NewMPC3421(addr int) (*MPC3421Driver, error) {
	adc := &MPC3421Driver{}
	b, err := sysfs.NewI2cDevice("/dev/i2c-1")
	if err != nil {
		return adc, err
	}

	err = b.SetAddress(addr)
	if err != nil {
		return adc, err
	}

	adc.bus = b
	err = adc.SetGain(1)
	return adc, err
}

// SetGain sets the gain of the ADC
func (adc *MPC3421Driver) SetGain(gain int) error {
	var cfgByte byte
	switch gain {
	case 1:
		cfgByte = 0x1C

	case 2:
		cfgByte = 0x1D

	case 4:
		cfgByte = 0x1E

	case 8:
		cfgByte = 0x1F

	default:
		cfgByte = 0x1C
		log.Println("gain needs to be 1,2,4, or 8 otherwise 1 will be used")
	}

	err := adc.bus.WriteByte(cfgByte)
	if err != nil {
		return err
	}

	adc.gain = gain
	adc.configByte = cfgByte
	time.Sleep(time.Second / 10)
	return nil
}

// AnalogRead returns the analog ADC value
func (adc *MPC3421Driver) AnalogRead() (int, error) {
	var val int

	reply := make([]byte, 4)
	_, err := adc.bus.Read(reply)
	if err != nil {
		return val, err
	}

	if (reply[3] & 0x1F) != adc.configByte {
		log.Printf("ERROR: config byte was different %x", reply)
		return val, fmt.Errorf("config byte was different")
	}

	val = (int(reply[0]&0x03) << 16) + (int(reply[1]) << 8) + int(reply[2])
	return val, nil
}

// Read returns the ADC value in volts
func (adc *MPC3421Driver) Read() (float64, error) {
	v := 0.0

	bits, err := adc.AnalogRead()
	if err != nil {
		return 0.0, err
	}

	sign := (bits & 0x20000) >> 17
	if sign != 0 {
		bits = (bits - 0x3FFFF) * -1
	}

	v = (float64(bits) / math.Pow(2, 17)) * MPC3421VRef
	v = v / float64(adc.gain)
	return v, nil
}
