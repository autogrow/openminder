package openminder

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

type fakeADC struct {
	v   float64
	err error
}

func (fadc fakeADC) Read() (float64, error) {
	return fadc.v, fadc.err
}

func (fadc fakeADC) AnalogRead() (int, error) {
	return 0, fadc.err
}

func TestMoistureCircuit(t *testing.T) {
	Convey("given a moisture circuit with an adc", t, func() {
		fadc := &fakeADC{}
		mc := MoistureCircuit{fadc}

		Convey("when the ADC has an error", func() {
			fadc.err = fmt.Errorf("test")

			Convey("when the voltage is read", func() {
				f, err := mc.Value()
				So(err, ShouldNotBeNil)
				So(f, ShouldEqual, 0)
			})
		})

		Convey("when the ADC is outputting 1v", func() {
			fadc.v = 1.0

			Convey("when the voltage is read", func() {
				f, err := mc.Value()
				So(err, ShouldBeNil)

				Convey("it should be the correct value", func() {
					So(f, ShouldAlmostEqual, 5.5, 0.01)
				})
			})
		})
	})
}
