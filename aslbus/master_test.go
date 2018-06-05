package aslbus

import (
	"testing"
	"time"

	"github.com/jacobsa/go-serial/serial"
	. "github.com/smartystreets/goconvey/convey"
)

func TestMaster(t *testing.T) {
	Convey("test the master side of the communications port", t, func() {
		opts := serial.OpenOptions{
			PortName:        "/dev/ttyUSB0",
			BaudRate:        57200,
			DataBits:        8,
			StopBits:        1,
			ParityMode:      serial.PARITY_NONE,
			MinimumReadSize: 9,
		}

		master := NewMaster(opts)

		So(master, ShouldNotBeNil)
		So(master.running, ShouldNotBeNil)
		So(master.port.Options.PortName, ShouldEqual, "/dev/ttyUSB0")
		Convey("test starting and stopping the port", func() {
			go master.Run()
			time.Sleep(time.Second)
			So(master.running, ShouldBeTrue)
			master.TransmitPacket(string(0xff), "ASL1805180000", "$0", "")
			time.Sleep(time.Second)
			master.Quit()
		})
		Convey("test recovering from closing channel", func() {
			go master.Run()
			time.Sleep(time.Second)
			So(master.running, ShouldBeTrue)
			master.TransmitPacket(string(0xff), "ASL1805180000", "$0", "")
			time.Sleep(time.Second)
			close(master.TxChannel)
			time.Sleep(time.Second)
			So(master.TxChannel, ShouldNotBeNil)
			master.Quit()
		})
		Convey("test transmit function", func() {
			opts := serial.OpenOptions{
				PortName:        "/dev/ttyUSB1",
				BaudRate:        57200,
				DataBits:        8,
				StopBits:        1,
				ParityMode:      serial.PARITY_NONE,
				MinimumReadSize: 9,
			}
			master2 := NewMaster(opts)
			pkt := NewTxPkt(string(0xff), "ASL1805180001", "$0", "")
			err := master2.transmit(pkt)
			So(err, ShouldNotBeNil)
			pkt.raw = ""
			opts = serial.OpenOptions{
				PortName:        "/dev/ttyUSB0",
				BaudRate:        57200,
				DataBits:        8,
				StopBits:        1,
				ParityMode:      serial.PARITY_NONE,
				MinimumReadSize: 9,
			}
			master2 = NewMaster(opts)
			err = master2.transmit(pkt)
			So(err, ShouldNotBeNil)
		})
	})
}
