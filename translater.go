package openminder

import (
	"fmt"
)

var translatableFields = []string{
	"irrig_ph",
	"irrig_ec",
	"runoff_ph",
	"runoff_ec",
	"irrig_volume",
	"runoff_volume",
}

// ErrNotTranslatable is returned by the Translater when a field is not translatable
var ErrNotTranslatable = fmt.Errorf("non-translatable field")

// DefaultDBPath is the default path to the database
var DefaultDBPath = "minder.db"

// NewTranslater returns a new translater object
func NewTranslater() (*Translater, error) {
	db, err := NewBoltedJSON(DefaultDBPath, "minder")

	return &Translater{db}, err
}

// Translater is an object that translates readings
type Translater struct {
	jdb *BoltedJSON
}

type calibration struct {
	Scale  float64 `json:"scale"`
	Offset float64 `json:"offset"`
}

func (c calibration) Transform(v float64) float64 {
	return (v * c.Scale) + c.Offset
}

// SetCalibration sets the calibration for the given field
func (tr *Translater) SetCalibration(field string, scale, offset float64) error {
	var translatable bool
	for _, f := range translatableFields {
		if f == field {
			translatable = true
		}
	}

	if !translatable {
		return ErrNotTranslatable
	}

	return tr.jdb.Set(field, calibration{scale, offset})
}

// getCalibration gets the calibration for the given field
func (tr *Translater) getCalibration(field string) (calibration, error) {
	calib := calibration{}
	err := tr.jdb.Get(field, &calib)
	return calib, err
}

// Translate will convert the given value as per the calibrations stored for the given field
func (tr *Translater) Translate(field string, value float64) (float64, error) {
	c, err := tr.getCalibration(field)
	if err != nil {
		return value, fmt.Errorf("can't find calibration constant for %s", field)
	}

	return c.Transform(float64(value)), nil
}
