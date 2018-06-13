package openminder

const (
	resistorDivider = 1 / 11
)

// MoistureInput models a PH circuit that contains an ADC
type MoistureInput struct {
	ADC
}

// NewMoistureInput returns a new PHCircuit with the given ADC
func NewMoistureInput(adc ADC) *MoistureInput {
	return &MoistureInput{adc}
}

// Value returns the pH value that this circuit reports
func (m *MoistureInput) Value() (float64, error) {
	volts, err := m.Read()
	if err != nil {
		return 0, err
	}

	voltsOut := volts / opampGain
	volts = voltsOut / resistorDivider
	return volts, nil
}
