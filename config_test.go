package openminder

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestAssignProbeSerials(t *testing.T) {
	Convey("given a config", t, func() {
		cfg := &Config{}

		Convey("with the runoff serial set", func() {
			cfg.RunoffECProbe = "1234"

			Convey("when a pair of serials is given, with the first identical to the runoff serial", func() {
				cfg.AssignProbeSerials("1234", "6789")

				Convey("it should not overwrite the runoff serial", func() {
					So(cfg.RunoffECProbe, ShouldEqual, "1234")
				})

				Convey("it should set the irrig serial", func() {
					So(cfg.IrrigECProbe, ShouldEqual, "6789")
				})
			})

			Convey("when a pair of serials is given, with the second identical to the runoff serial", func() {
				cfg.AssignProbeSerials("6789", "1234")

				Convey("it should not overwrite the runoff serial", func() {
					So(cfg.RunoffECProbe, ShouldEqual, "1234")
				})

				Convey("it should set the irrig serial", func() {
					So(cfg.IrrigECProbe, ShouldEqual, "6789")
				})
			})
		})

		Convey("with the irrig serial set", func() {
			cfg.IrrigECProbe = "1234"

			Convey("when a pair of serials is given, with the first identical to the irrig serial", func() {
				cfg.AssignProbeSerials("1234", "6789")

				Convey("it should not overwrite the irrig serial", func() {
					So(cfg.IrrigECProbe, ShouldEqual, "1234")
				})

				Convey("it should set the runoff serial", func() {
					So(cfg.RunoffECProbe, ShouldEqual, "6789")
				})
			})

			Convey("when a pair of serials is given, with the second identical to the irrig serial", func() {
				cfg.AssignProbeSerials("6789", "1234")

				Convey("it should not overwrite the irrig serial", func() {
					So(cfg.IrrigECProbe, ShouldEqual, "1234")
				})

				Convey("it should set the runoff serial", func() {
					So(cfg.RunoffECProbe, ShouldEqual, "6789")
				})
			})
		})

		Convey("with both serial numbers are set", func() {
			cfg.IrrigECProbe = "1234"
			cfg.RunoffECProbe = "6789"

			Convey("and two new serial numbers come in", func() {
				cfg.AssignProbeSerials("abcd", "efgh")

				Convey("it should assign the new serial numbers", func() {
					So(cfg.RunoffECProbe, ShouldEqual, "abcd")
					So(cfg.IrrigECProbe, ShouldEqual, "efgh")
				})
			})

			Convey("and identical serial numbers come in", func() {
				cfg.AssignProbeSerials("6789", "1234")

				Convey("they should not be reassigned", func() {
					So(cfg.RunoffECProbe, ShouldEqual, "6789")
					So(cfg.IrrigECProbe, ShouldEqual, "1234")
				})
			})

			Convey("and identical serial numbers come in, reversed", func() {
				cfg.AssignProbeSerials("1234", "6789")

				Convey("they should not be reassigned", func() {
					So(cfg.RunoffECProbe, ShouldEqual, "6789")
					So(cfg.IrrigECProbe, ShouldEqual, "1234")
				})
			})
		})

		Convey("with both serial numbers are not set", func() {
			cfg.IrrigECProbe = ""
			cfg.RunoffECProbe = ""

			Convey("and two new serial numbers come in", func() {
				cfg.AssignProbeSerials("abcd", "efgh")

				Convey("it should assign the new serial numbers", func() {
					So(cfg.RunoffECProbe, ShouldEqual, "abcd")
					So(cfg.IrrigECProbe, ShouldEqual, "efgh")
				})
			})
		})

	})
}
