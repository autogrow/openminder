package openminder

// ADC is an interface to an ADC chip
type ADC interface {
	AnalogRead() (int, error)
	Read() (float64, error)
}
