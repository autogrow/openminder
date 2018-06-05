package calib

import (
	"fmt"
	"math"
)

// PH will take a ph7 and ph4 value and calculate the scale and offset.
// an error will be returned if the value are too far out of range (indicating a bad probe)
func PH(ph7, ph4 float64) (scale, offset float64, err error) {
	// If the pH7 is below the pH4 value this is invlaid
	if ph7 < ph4 {
		err = fmt.Errorf("pH7 value (%.2f) is lower than pH4 (%.2f)", ph7, ph4)
		return
	}

	scale = 3.0 / (ph7 - ph4)
	offset = 7.0 - ph7

	absoffset := math.Abs(offset)
	if absoffset > 1.0 {
		err = fmt.Errorf("pH7 value (%.2f) produced an incorrect scale value (%.2f)", ph7, absoffset)
		return
	}

	// a pH7 needs to be between 6 and 8 to be valid
	if ph7 < 6.0 || ph7 > 8.0 {
		err = fmt.Errorf("pH7 value (%.2f) is out of range", ph7)
		return
	}

	// a pH4 needs to be between 3 and 5 to be valid
	if ph4 < 3.0 || ph4 > 5.0 {
		err = fmt.Errorf("pH4 value (%.2f) is out of range", ph4)
		return
	}

	// an inrange pH7 and ph4 should only produce a scale within this range
	if scale < 0.75 || scale > 1.5 {
		err = fmt.Errorf("pH7 value (%.2f) and pH4 value (%.2f) produces an out of range scale (%.2f)", ph7, ph4, scale)
		return
	}

	return
}

// TransformPH uses the special formula to transform the uncalibrated pH into
// the calibrated value using the scale and offset.
func TransformPH(reading, scale, offset float64) float64 {
	return (((reading - 7.0) + offset) * scale) + 7.0
}

// EC takes a buffer value and an EC reading from in that buffer and calculates
// the scale and offset to calibrate future values from.  OK will be false if
// the ec reading was too far out of range (indicating a bad probe)
func EC(buffer, ec float64) (scale, offset float64, err error) {
	scale = buffer / ec
	min, max := (buffer * 0.5), (buffer * 1.5)

	if ec <= max && ec >= min {
		err = fmt.Errorf("EC reading out of range")
	}

	return
}
