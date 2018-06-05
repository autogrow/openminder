package openminder

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpiotest"
)

func TestContactClosure(t *testing.T) {
	Convey("given a cc with a high pin", t, func() {

		p := new(gpiotest.Pin)
		p.L = gpio.High
		cc := &ContactClosure{}
		cc.Pin = p

		Convey("when the cc is started with an interrupt callback", func() {
			closed := false
			cc.onClosureCB = func() {
				closed = true
			}

			cc.Start()

			Convey("and the pin reads low for < 400ms", func() {
				p.L = gpio.Low
				time.Sleep(300 * time.Millisecond)
				p.L = gpio.High

				Convey("it should not have closed", func() {
					So(closed, ShouldBeFalse)
				})
			})

			Convey("and the pin reads low for > 400ms", func() {
				p.L = gpio.Low
				time.Sleep(500 * time.Millisecond)

				Convey("and the pin goes high again", func() {
					p.L = gpio.High
					time.Sleep(100 * time.Millisecond)

					Convey("it should have closed", func() {
						So(closed, ShouldBeTrue)
					})
				})

				Convey("and the pin stays low", func() {
					time.Sleep(100 * time.Millisecond)

					Convey("it should not have closed", func() {
						So(closed, ShouldBeFalse)
					})
				})
			})
		})

		cc.Stop()
	})
}
