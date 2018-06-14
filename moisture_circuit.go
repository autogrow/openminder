package openminder

const (
	resistorDivider = 0.09090
)

// MoistureCircuit models a moisture circuit that contains an ADC
type MoistureCircuit struct {
	ADC
}

// NewMoistureCircuit returns a new MoistureCircuit with the given ADC
func NewMoistureCircuit(adc ADC) *MoistureCircuit {
	return &MoistureCircuit{adc}
}

// Value returns the moisture value that this circuit reports
func (m *MoistureCircuit) Value() (float64, error) {
	volts, err := m.Read()
	if err != nil {
		return 0, err
	}

	voltsOut := volts / opampGain
	volts = voltsOut / resistorDivider
	return volts, nil
}
