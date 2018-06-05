package aslbus

import (
	"testing"
	"time"

	"github.com/jacobsa/go-serial/serial"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSlave(t *testing.T) {
	Convey("test the slave side of the communications port", t, func() {
		opts := serial.OpenOptions{
			PortName:        "/dev/ttyUSB0",
			BaudRate:        57200,
			DataBits:        8,
			StopBits:        1,
			ParityMode:      serial.PARITY_NONE,
			MinimumReadSize: 9,
		}

		rxChan := make(chan string)
		slave := NewSlave(opts, rxChan)

		So(slave, ShouldNotBeNil)
		So(slave.running, ShouldNotBeNil)
		So(slave.port.Options.PortName, ShouldEqual, "/dev/ttyUSB0")
		So(slave.rxChan, ShouldNotBeNil)
		So(slave.IsOpen(), ShouldBeFalse)
		Convey("test starting and stopping the port", func() {
			go slave.Listen()
			time.Sleep(time.Second)
			So(slave.Running(), ShouldBeTrue)
			slave.Quit()
			time.Sleep(2 * time.Second)
			So(slave.running, ShouldBeFalse)
		})

	})
}
