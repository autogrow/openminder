package aslbus

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestPacketing(t *testing.T) {
	Convey("convert a BCD string to integer works correclty", t, func() {
		str1 := "10"
		str2 := "3C40"
		str3 := "34B9A4DE"

		num1 := Hextobin(str1)
		num2 := Hextobin(str2)
		num3 := Hextobin(str3)

		So(num1, ShouldEqual, 16)
		So(num2, ShouldEqual, 15424)
		So(num3, ShouldEqual, 884581598)
	})

	Convey("packet CRC is correctly calculated", t, func() {
		packet := ":ASL1805180001$0000123"
		crc := calculateCRC(packet)
		So(crc, ShouldEqual, "7010")
	})

	Convey("new tx packet is correclty generated", t, func() {
		pkt := NewTxPkt(masterAddress, "ASL1805180001", "$0", "23")

		So(pkt.serial, ShouldEqual, "ASL1805180001")
		So(pkt.cmd, ShouldEqual, "$0")
		So(pkt.datacnt, ShouldEqual, "0100")
		So(pkt.data, ShouldEqual, "23")
		So(pkt.crc, ShouldEqual, "043E")
		So(pkt.raw, ShouldEqual, ":!ASL1805180001$0010023043E"+string(pktEOF))
		So(pkt.timestamp, ShouldEqual, time.Now().Unix())

		txBytes := pkt.Bytes()
		So(len(txBytes), ShouldNotEqual, 0)
		So(len(txBytes), ShouldEqual, len(pkt.raw)+2)
	})
}

func TestRxPackets(t *testing.T) {
	Convey("given a valid packet", t, func() {
		pkt, err := NewRxPkt(":xASL1805180001$00100235732")
		So(err, ShouldBeNil)

		Convey("it should be correctly parsed", func() {
			So(pkt.serial, ShouldEqual, "ASL1805180001")
			So(pkt.cmd, ShouldEqual, "$0")
			So(pkt.datacnt, ShouldEqual, "0001")
			So(pkt.data, ShouldEqual, "23")
			So(pkt.crc, ShouldEqual, "5732")
			So(pkt.timestamp, ShouldEqual, time.Now().Unix())
		})
	})

	Convey("given invalid packets", t, func() {
		Convey("errors should be detected", func() {
			_, err := NewRxPkt("ASL1805180001")
			So(err, ShouldEqual, ErrNoStartChar)

			_, err = NewRxPkt(":ASL1805180001")
			So(err, ShouldEqual, ErrTooShort)

			_, err = NewRxPkt(":ASL1805180001$0")
			So(err, ShouldEqual, ErrTooShort)

			_, err = NewRxPkt(":xASL1805180001$00001")
			So(err, ShouldEqual, ErrNoCRC)

			_, err = NewRxPkt(":xASL1805180001$001004CF3")
			So(err, ShouldEqual, ErrSizeMismatch)

			_, err = NewRxPkt(":!ASL1805180001r001002341B7")
			So(err, ShouldEqual, ErrMasterPacket)

			_, err = NewRxPkt(":xASL1805180001$001002344F3")
			So(err.Error(), ShouldContainSubstring, "CRC Failed")
		})
	})
}
