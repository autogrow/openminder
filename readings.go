package openminder

import "bitbucket.org/autogrowsystems/openminder/types"

// Readings represents the readings kept by the minder
type Readings struct {
	IrrigADC        int              `json:"irrig_adc"`
	IrrigPHVoltage  float64          `json:"irrig_ph_voltage"`
	IrrigPHRaw      float64          `json:"irrig_ph_raw"`
	IrrigPH         float64          `json:"irrig_ph"`
	RunoffADC       int              `json:"runoff_adc"`
	RunoffPHVoltage float64          `json:"runoff_ph_voltage"`
	RunoffPHRaw     float64          `json:"runoff_ph_raw"`
	RunoffPH        float64          `json:"runoff_ph"`
	IrrigTips       int64            `json:"irrig_tips"`
	IrrigVolume     float64          `json:"irrig_volume"`
	RunoffTips      int64            `json:"runoff_tips"`
	RunoffVolume    float64          `json:"runoff_volume"`
	IrrigEC         *types.NullFloat `json:"irrig_ec"`
	IrrigECRaw      *types.NullFloat `json:"irrig_ec_raw"`
	IrrigECTemp     *types.NullFloat `json:"irrig_ectemp"`
	RunoffEC        *types.NullFloat `json:"runoff_ec"`
	RunoffECRaw     *types.NullFloat `json:"runoff_ec_raw"`
	RunoffECTemp    *types.NullFloat `json:"runoff_ectemp"`
	RunoffRatio     float64          `json:"runoff_ratio"`
	MoistureADC     int              `json:"moisture_adc"`
	MoistureVoltage float64          `json:"moisture_voltage"`
	Moisture        float64          `json:"moisture"`
}

func newReadings() *Readings {
	return &Readings{
		IrrigEC:      &types.NullFloat{},
		IrrigECRaw:   &types.NullFloat{},
		IrrigECTemp:  &types.NullFloat{},
		RunoffEC:     &types.NullFloat{},
		RunoffECRaw:  &types.NullFloat{},
		RunoffECTemp: &types.NullFloat{},
	}
}
