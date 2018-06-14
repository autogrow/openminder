package main

import (
	"flag"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/autogrow/openminder/aslbus"

	"gobot.io/x/gobot/sysfs"
)

func main() {
	var mode string
	var device string
	var reg string
	var write string
	var read int
	var gain int

	flag.StringVar(&mode, "mode", "ph", "pH or DO")
	flag.StringVar(&device, "a", "0x68", "address of device")
	flag.StringVar(&reg, "reg", "", "register to start read from, if empty no address will be specifed")
	flag.StringVar(&write, "write", "", "bytes to write if empty no write will be done")
	flag.IntVar(&read, "read", 1, "number of bytes to read")
	flag.IntVar(&gain, "gain", 1, "adc gain")
	flag.Parse()

	// Open the first available I²C bus:
	d, err := sysfs.NewI2cDevice("/dev/i2c-1")
	if err != nil {
		fmt.Println(err)
		return
	}

	addr := strings.Trim(device, "0x")
	addrByte := aslbus.Hextobin(addr)
	err = d.SetAddress(addrByte)
	if err != nil {
		fmt.Println(err)
		return
	}

	var writeByte byte

	switch gain {
	case 1:
		writeByte = 0x1C

	case 2:
		writeByte = 0x1D

	case 4:
		writeByte = 0x1E

	case 8:
		writeByte = 0x1F

	default:
		writeByte = 0x1C
		fmt.Println("gain needs to be 1,2,4, or 8 otherwise 1 will be used")
	}

	err = d.WriteByte(writeByte)
	if err != nil {
		fmt.Println(err)
		return
	}

	bytes := make([]byte, 4)
	for {
		time.Sleep(time.Second)

		b, err := d.Read(bytes)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("Read: ", b, " bytes")
		fmt.Printf("0x%02X\n0x%02X\n0x%02X\n0x%02X\n", bytes[0], bytes[1], bytes[2], bytes[3])
		val := (int(bytes[0]&0x03) << 16) + (int(bytes[1]) << 8) + int(bytes[2])
		sign := (val & 0x20000) >> 17
		fmt.Println("VAL: ", val, "SIGN: ", sign)
		var voltage float64
		if sign != 0 {
			fmt.Println("Value is negative")
			// Number is negative
			val = (val - 0x3FFFF) * -1
			voltage = (float64(val) / math.Pow(2, 17)) * 2.048
			fmt.Println("Raw ADC: ", val)
			fmt.Println("VOLTAGE: ", voltage)
			phv := ((voltage / float64(gain)) / 2) / 0.059
			fmt.Println("PH V: ", phv)
			ph := phv + 7
			fmt.Println("pH: ", ph, " DO: ", phv)
		} else {
			voltage = (float64(val) / math.Pow(2, 17)) * 2.048
			fmt.Println("Raw ADC: ", val)
			fmt.Println("VOLTAGE: ", voltage)
			phv := ((voltage / float64(gain)) / 2) / 0.059
			fmt.Println("PH V: ", phv)
			ph := 7 - phv
			fmt.Println("pH: ", ph, " DO: ", phv)
		}
	}

}
