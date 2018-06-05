package openminder

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestECProbeSwap(t *testing.T) {
	Convey("given a minder with a config and change callback", t, func() {
		m := &Minder{cfg: &Config{}}

		var calledCfg Config
		m.OnConfigChange(func(cfg Config) {
			calledCfg = cfg
		})

		Convey("and two EC probe serials in the config", func() {
			m.cfg.IrrigECProbe = "A"
			m.cfg.RunoffECProbe = "B"

			Convey("when the probes are swapped", func() {
				m.swapECProbes()

				Convey("the serials should be reversed", func() {
					m.cfg.IrrigECProbe = "B"
					m.cfg.RunoffECProbe = "A"
				})

				Convey("it should pass the reversed config to the callback", func() {
					calledCfg.IrrigECProbe = "B"
					calledCfg.RunoffECProbe = "A"
				})
			})
		})
	})
}
