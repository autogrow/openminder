package aslbus

import (
	"runtime"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func wait(ms int) {
	time.Sleep(time.Duration(ms) * time.Millisecond)
}

func testPacket() string {
	serial := ":xASL1805180000"
	cmd := "r0"
	dataCnt := "4000"
	data := "00"       // ASLStatus
	data += "00"       // spare1
	data += "0301"     // Firmware Version
	data += "01010101" // Status Bool
	data += "1501"     // ECx100
	data += "F609"     // Tempx100
	for i := 0; i < 41; i++ {
		data += "0" // Spares
	}
	data += "83445587" // EC Real
	data += "83445587" // Temp Real
	data += "00"       // Spare
	data += "01"       // SigPC

	msg := serial + cmd + dataCnt + data
	crc := calculateCRC(msg)
	msg += crc

	return msg
}

func TestECProbeReadings(t *testing.T) {
	Convey("given an ECProbe", t, func() {
		probe := NewECProbe("ASL1805180000")

		Convey("it should have the defaults", func() {
			So(probe.Serial, ShouldEqual, "ASL1805180000")
			So(probe.LastSeen, ShouldBeZeroValue)
			So(probe.running, ShouldBeFalse)
			So(probe.master, ShouldBeNil)
			So(probe.quit, ShouldNotBeNil)
			So(probe.EC, ShouldBeZeroValue)
			So(probe.Temp, ShouldBeZeroValue)
			So(probe.IsValid(), ShouldBeFalse)
		})

		Convey("when given a valid packet", func() {
			pkt, err := NewRxPkt(testPacket())
			So(err, ShouldBeNil)
			err = probe.Update(pkt)
			So(err, ShouldBeNil)

			Convey("it should update the probe readings", func() {
				So(probe.EC, ShouldEqual, 2.77)
				So(probe.Temp, ShouldEqual, 25.5)
				So(probe.LastSeen, ShouldEqual, pkt.timestamp)
			})
		})

		Convey("it should not start without a master", func() {
			So(probe.Start(), ShouldEqual, ErrProbeNotAttached)
		})

		Convey("with a master", func() {
			probe.master = &Master{}

			Convey("it should start and stop", func() {
				go func() {
					wait(100)
					probe.Stop()
				}()

				So(probe.Start(), ShouldBeNil)
				So(probe.quit, ShouldBeNil)
			})

			Convey("it should request readings", func() {
				So(probe.requestReading(), ShouldBeNil)
				So(probe.master.txRequests, ShouldEqual, 1)
			})
		})

		Convey("when attached to a bus", func() {
			bus := &Bus{master: &Master{}}
			threads := runtime.NumGoroutine()

			probe.AttachBus(bus)

			Convey("it should attach the master from the bus", func() {
				So(probe.master, ShouldEqual, bus.master)
			})

			Convey("it should register the probe with the bus", func() {
				So(bus.Serials(), ShouldContain, probe.SN())
			})

			Convey("it should spawn the start thread", func() {
				wait(50)
				So(runtime.NumGoroutine(), ShouldEqual, threads+1)
				So(probe.running, ShouldBeTrue)
				So(probe.quit, ShouldNotBeNil)
			})

			probe.Stop()
		})

		// 		Convey("updating with a invalid packet should return an error", func() {
		// 			serial := ":ASL1805180000"
		// 			cmd := "R0"
		// 			dataCnt := "0000"

		// 			msg := serial + cmd + dataCnt
		// 			crc := calculateCRC(msg)
		// 			msg += crc
		// 			rxPkt, err := NewRxPkt(msg)
		// 			So(err, ShouldBeNil)
		// 			master.running = true
		// 			err = probe.Update(rxPkt)
		// 			So(err.Error(), ShouldContainSubstring, "no data")
		// 			txPkt := <-master.TxChannel
		// 			So(txPkt, ShouldNotBeNil)
		// 			So(probe.LastSeen, ShouldEqual, time.Now().Unix())
		// 			So(probe.IsValid(), ShouldBeTrue)
		// 		})

		// 		Convey("updating with a valid new ping command should update probe", func() {
		// 			serial := ":ASL1805180000"
		// 			cmd := "$0"
		// 			dataCnt := "0000"

		// 			msg := serial + cmd + dataCnt
		// 			crc := calculateCRC(msg)
		// 			msg += crc
		// 			rxPkt, err := NewRxPkt(msg)
		// 			So(err, ShouldBeNil)
		// 			master.running = true
		// 			err = probe.Update(rxPkt)
		// 			So(err, ShouldBeNil)
		// 			txPkt := <-master.TxChannel
		// 			So(txPkt, ShouldNotBeNil)
		// 			So(probe.LastSeen, ShouldEqual, time.Now().Unix())
		// 			So(probe.IsValid(), ShouldBeTrue)
		// 		})

		// 		Convey("updating with a valid new packet should update probe", func() {

		// 			txPkt := <-master.TxChannel
		// 			So(txPkt, ShouldNotBeNil)
		// 			So(probe.LastSeen, ShouldEqual, time.Now().Unix())

		// 			So(probe.IsValid(), ShouldBeTrue)

		// 			probe.LastSeen = time.Now().Unix() - 119
		// 			go probe.interrogate(1, 1)

		// 			var pkts []Packet
		// 			for {
		// 				pkt := <-master.TxChannel
		// 				pkts = append(pkts, *pkt)
		// 				if len(pkts) > 5 {
		// 					probe.Update(rxPkt)
		// 				}
		// 				if len(pkts) > 7 {
		// 					probe.Quit()
		// 					break
		// 				}
		// 			}
		// 			readPktDetected := false
		// 			enablePingDetected := false
		// 			disablePingDetected := false
		// 			pingDetected := false
		// 			for _, pkt := range pkts {
		// 				switch pkt.cmd {
		// 				case readingCommand:
		// 					readPktDetected = true
		// 				case enablePingCommand:
		// 					enablePingDetected = true
		// 				case disablePingCommand:
		// 					disablePingDetected = true
		// 				case pingCommand:
		// 					pingDetected = true
		// 				}
		// 			}
		// 			So(readPktDetected, ShouldBeTrue)
		// 			So(enablePingDetected, ShouldBeTrue)
		// 			So(disablePingDetected, ShouldBeTrue)
		// 			So(pingDetected, ShouldBeTrue)
		// 		})
	})
}
