package openminder

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCalculateRunoffRatio(t *testing.T) {
	Convey("given a default config", t, func() {
		cfg := Config{}

		Convey("and some volume readings", func() {
			r := newReadings()
			r.IrrigVolume = 40
			r.RunoffVolume = 30

			Convey("it should calculate the ratio by volume only", func() {
				CalculateRunoffRatio(r, cfg)
				So(r.RunoffRatio, ShouldEqual, 0.75)
			})
		})

		Convey("and no volume readings", func() {
			r := newReadings()

			Convey("it should not calculate the runoff ratio", func() {
				CalculateRunoffRatio(r, cfg)
				So(r.RunoffRatio, ShouldEqual, 0)
			})
		})
	})

	Convey("given a config", t, func() {
		cfg := Config{
			DrippersPerPlant: 4,
			RunoffDrippers:   8,
			IrrigDrippers:    2,
		}

		Convey("and some volume readings", func() {
			r := newReadings()
			r.IrrigVolume = 80
			r.RunoffVolume = 320

			Convey("it should calculate the runoff ratio", func() {
				CalculateRunoffRatio(r, cfg)
				So(r.RunoffRatio, ShouldEqual, 1)
			})
		})

		Convey("and no volume readings", func() {
			r := newReadings()

			Convey("it should not calculate the runoff ratio", func() {
				CalculateRunoffRatio(r, cfg)
				So(r.RunoffRatio, ShouldEqual, 0)
			})
		})
	})
}
