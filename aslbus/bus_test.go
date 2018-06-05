package aslbus

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func mockECProbe(sn string) Probe {
	return &ECProbe{Serial: sn}
}

func TestBus(t *testing.T) {
	Convey("given a new bus", t, func() {
		tty := "/dev/ttyUSB0"
		bus := New(tty)

		So(bus.master, ShouldNotBeNil)
		So(bus.master.port.Options.PortName, ShouldEqual, tty)
		So(bus.slave, ShouldNotBeNil)
		So(bus.slave.port.Options.PortName, ShouldEqual, tty)
		So(bus.ReadingsChan, ShouldNotBeNil)
		So(bus.slave.rxChan, ShouldNotBeNil)
		So(bus.probes, ShouldBeEmpty)

		Convey("and two EC probes", func() {
			probe1 := mockECProbe("ASL1805180000")
			probe2 := mockECProbe("ASL1805180001")

			Convey("when the probes are registered with the bus", func() {
				bus.registerProbe(probe1)
				bus.registerProbe(probe2)

				Convey("they should be added to the bus", func() {
					So(bus.Serials(), ShouldContain, probe1.SN())
					So(bus.Serials(), ShouldContain, probe2.SN())
					So(len(bus.probes), ShouldEqual, 2)
				})

				Convey("an identical probe should not be added", func() {
					bus.registerProbe(probe1)
					So(len(bus.probes), ShouldEqual, 2)
				})
			})

		})
	})
}
