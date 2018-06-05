package types

import (
	"database/sql"
	"encoding/json"
)

// NullFloat represents a float that can be null when marshalled to JSON
type NullFloat struct {
	sql.NullFloat64
}

// Value returns the value of the NullFloat (even if it is invalid)
func (r *NullFloat) Value() float64 {
	return r.Float64
}

// SetValue sets the value of the NullFloat
func (r *NullFloat) SetValue(v float64) {
	r.Float64 = v
	r.Valid = true
}

// IsValid returns true or false if the NullFloat is valid
func (r *NullFloat) IsValid() bool {
	return r.Valid
}

// SetInvalid sets the NullFloat as invalid
func (r *NullFloat) SetInvalid() {
	r.Valid = false
}

// MarshalJSON allows to marshal this struct into JSON
func (r NullFloat) MarshalJSON() ([]byte, error) {
	var s interface{} = r.Float64
	if r.Valid == false {
		s = nil
	}
	return json.Marshal(s)
}

// UnmarshalJSON allows to marshal this struct into JSON
func (r *NullFloat) UnmarshalJSON(b []byte) error {
	var float interface{}

	if err := json.Unmarshal(b, &float); err != nil {
		return err
	}

	if float == nil {
		r.SetInvalid()
		return nil
	}

	r.SetValue(float.(float64))
	return nil
}
