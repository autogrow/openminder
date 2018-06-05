package openminder

import "math"

const (
	phPerVolt = 0.059
	phRef     = 1.220
	opampGain = 2
)

// PHCircuit models a PH circuit that contains an ADC
type PHCircuit struct {
	ADC
}

// NewPHCircuit returns a new PHCircuit with the given ADC
func NewPHCircuit(adc ADC) *PHCircuit {
	return &PHCircuit{adc}
}

// Value returns the pH value that this circuit reports
func (phc *PHCircuit) Value() (float64, error) {
	var ph float64
	volts, err := phc.Read()
	if err != nil {
		return ph, err
	}

	volts = (volts * -1) / opampGain
	ph = voltsToPH(phPerVolt, volts)
	return ph, nil
}

func voltsToPH(phPerVolt, volts float64) float64 {
	ph := 7.0 + (volts / phPerVolt)
	return math.Abs(math.Round(ph/0.1) * 0.1)
}
