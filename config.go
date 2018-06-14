package openminder

import (
	"encoding/json"
	"io/ioutil"
)

// Config is the configuration for the OpenMinder
type Config struct {
	// IrrigTBGPIO is the GPIO port that should be used for the irrigation tipping bucket
	IrrigTBGPIO string `json:"irrig_tb_gpio"`

	// RunoffTBGPIO is the GPIO port that should be used for the runoff tipping bucket
	RunoffTBGPIO string `json:"runoff_tb_gpio"`

	// Port is the port that the API should run on
	Port string `json:"port"`

	// ScanTimeout dictates how long the probe scan should run for
	ScanTimeout int `json:"scan_timeout"`

	// IrrigECProbe contains the serial number of the EC probe on the irrigation side
	IrrigECProbe string `json:"irrig_ec_probe"`

	// RunoffECProbe contains the serial number of the EC probe on the runoff side
	RunoffECProbe string `json:"runoff_ec_probe"`

	// TTY is the bus to use for the ASL Bus comms
	TTY string `json:"tty"`

	// MoistureGain is the gain to use with the moisture probe this should be 1,2,4 or 8
	MoistureGain int `json:"moisture_gain"`

	DrippersPerPlant int `json:"drippers_per_plant"`
	RunoffDrippers   int `json:"runoff_drippers"`
	IrrigDrippers    int `json:"irrig_drippers"`
}

// AssignProbeSerials assigns the probes in a way that preserves the order that the
// probes may have been set to before
func (cfg *Config) AssignProbeSerials(serials ...string) {
	if len(serials) != 2 {
		return
	}

	i := cfg.IrrigECProbe
	r := cfg.RunoffECProbe
	f := serials[0]
	l := serials[1]

	// assign the run-off serial
	// is runoff already assigned
	if r != f && r != l {
		// we need to assign a serial
		if f != i {
			r = f
		} else {
			r = l
		}
	}

	if i != f && i != l {
		// we need to assign a serial
		if f != r {
			i = f
		} else {
			i = l
		}
	}

	cfg.IrrigECProbe = i
	cfg.RunoffECProbe = r
}

// LoadFrom will load the config from the given filename
func (cfg *Config) LoadFrom(fn string) error {
	data, err := ioutil.ReadFile(fn)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, cfg)
}

// SaveTo will load the config from the given filename
func (cfg *Config) SaveTo(fn string) error {
	data, err := json.Marshal(cfg)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(fn, data, 0644)
}
