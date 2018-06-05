package aslbus

import (
	"testing"

	"github.com/jacobsa/go-serial/serial"
	. "github.com/smartystreets/goconvey/convey"
)

func TestPort(t *testing.T) {
	Convey("test an invalid port creation", t, func() {
		opts := serial.OpenOptions{
			PortName:        "/dev/ttyUSB1",
			BaudRate:        57200,
			DataBits:        8,
			StopBits:        1,
			ParityMode:      serial.PARITY_NONE,
			MinimumReadSize: 9,
		}

		port := NewPort(opts)

		So(port, ShouldNotBeNil)
		So(port.State, ShouldEqual, portClosed)
		So(port.Options.PortName, ShouldEqual, "/dev/ttyUSB1")
		Convey("test starting and stopping the port", func() {
			err := port.Open()
			So(err.Error(), ShouldContainSubstring, "no such file")
		})
	})
}
