package openminder

import (
	"fmt"
	"log"
	"time"

	"bitbucket.org/autogrowsystems/openminder/aslbus"
)

// Minder is a model that holds the objects that when
// combined make up the minder logic
type Minder struct {
	stopped       bool
	tr            *Translater
	cfg           *Config
	bus           *aslbus.Manager
	irrigPH       *PHCircuit
	runoffPH      *PHCircuit
	irrigTB       *TippingBucket
	runoffTB      *TippingBucket
	moisture      *MoistureInput
	Readings      *Readings
	errors        *errorStore
	onCfgChangeCB func(Config)
}

// NewMinder returns a new minder object with the default comprising
// objects already setup
func NewMinder(cfg *Config) (mdr *Minder, err error) {
	mdr = &Minder{
		Readings:      newReadings(),
		cfg:           cfg,
		onCfgChangeCB: func(cfg Config) {},
		errors:        newErrorStore(),
	}

	if mdr.tr, err = NewTranslater(); err != nil {
		return nil, err
	}

	mdr.init()

	return mdr, nil
}

// OnConfigChange registers a function to be called when the config changes
func (mdr *Minder) OnConfigChange(cb func(Config)) {
	mdr.onCfgChangeCB = cb
}

func (mdr *Minder) init() {
	mdr.initBus()
	mdr.initTBs()
	mdr.initADCs()
}

func (mdr *Minder) initBus() {
	mdr.bus = aslbus.NewManager(mdr.cfg.TTY, mdr.cfg.ScanTimeout, mdr.cfg.IrrigECProbe, mdr.cfg.RunoffECProbe)

	mdr.bus.OnError(func(err error) {
		log.Printf("ERROR: bus: %s", err)
		mdr.errors.Add(fmt.Errorf("bus error: %s", err))
	})

	mdr.bus.OnScanDone(func(serials []string, err error) {
		if err != nil {
			err = fmt.Errorf("scan failed: %s", err)
			mdr.errors.Add(err)
			log.Printf("ERROR: %s", err)
		}

		mdr.cfg.AssignProbeSerials(serials...)
		go mdr.onCfgChangeCB(*mdr.cfg)
	})

	go mdr.bus.Run()
}

func (mdr *Minder) initTBs() {
	go func() {
		for {
			tb, err := NewTippingBucket(mdr.cfg.IrrigTBGPIO)
			if err != nil {
				mdr.errors.Add(fmt.Errorf("ERROR: failed to connect to TB on %s: %s", mdr.cfg.IrrigTBGPIO, err))
				time.Sleep(time.Second)
				continue
			}
			mdr.irrigTB = tb

			break
		}

		mdr.irrigTB.OnTip(func() {
			var err error
			mdr.Readings.IrrigTips++
			mdr.Readings.IrrigVolume, err = mdr.tr.Translate("irrig_volume", float64(mdr.Readings.IrrigTips))
			CalculateRunoffRatio(mdr.Readings, *mdr.cfg)
			mdr.errors.Add(err)
		})
	}()

	go func() {
		for {
			tb, err := NewTippingBucket(mdr.cfg.RunoffTBGPIO)
			if err != nil {
				mdr.errors.Add(fmt.Errorf("ERROR: failed to connect to TB on %s: %s", mdr.cfg.RunoffTBGPIO, err))
				time.Sleep(time.Second)
				continue
			}
			mdr.runoffTB = tb
			break
		}

		mdr.runoffTB.OnTip(func() {
			var err error
			mdr.Readings.RunoffTips++
			mdr.Readings.RunoffVolume, err = mdr.tr.Translate("irrig_volume", float64(mdr.Readings.RunoffTips))
			CalculateRunoffRatio(mdr.Readings, *mdr.cfg)
			mdr.errors.Add(err)
		})
	}()
}

func (mdr *Minder) initADCs() {
	go func() {
		for {
			adc, err := NewMPC3421(0x68)
			if err != nil {
				mdr.errors.Add(fmt.Errorf("ERROR: failed to connect to ADC 1: %s", err))
				time.Sleep(time.Second)
				continue
			}
			mdr.irrigPH = NewPHCircuit(adc)
			return
		}
	}()

	go func() {
		for {
			adc, err := NewMPC3421(0x69)
			if err != nil {
				mdr.errors.Add(fmt.Errorf("ERROR: failed to connect to ADC 2: %s", err))
				time.Sleep(time.Second)
				continue
			}
			mdr.runoffPH = NewPHCircuit(adc)
			return
		}
	}()

	go func() {
		for {
			adc, err := NewMPC3421(0x70)
			if err != nil {
				mdr.errors.Add(fmt.Errorf("ERROR: failed to connect to ADC 3: %s", err))
				time.Sleep(time.Second)
				continue
			}
			err = adc.SetGain(mdr.cfg.MoistureGain)
			mdr.errors.Add(fmt.Errorf("ERROR: moisture adc gain error: %s", err))
			mdr.moisture = NewMoistureInput(adc)
			return
		}
	}()
}

// Start the minder loop
func (mdr *Minder) Start() {
	for {
		if mdr.stopped {
			return
		}

		mdr.readECProbes()
		mdr.readPHProbes()
		mdr.readMoistureProbe()
		time.Sleep(time.Second)
	}
}

// Stop the minder loop
func (mdr *Minder) Stop() {
	mdr.stopped = true
}

func (mdr *Minder) readPHProbes() {
	var err error

	if mdr.irrigPH != nil {
		mdr.Readings.IrrigADC, err = mdr.irrigPH.AnalogRead()
		mdr.errors.Add(err)
		mdr.Readings.IrrigPHVoltage, err = mdr.irrigPH.Read()
		mdr.errors.Add(err)
		mdr.Readings.IrrigPHRaw, err = mdr.irrigPH.Value()
		mdr.errors.Add(err)
		mdr.Readings.IrrigPH, err = mdr.tr.Translate("irrig_ph", mdr.Readings.IrrigPHRaw)
		mdr.errors.Add(err)
	}

	if mdr.runoffPH != nil {
		mdr.Readings.RunoffADC, err = mdr.runoffPH.AnalogRead()
		mdr.errors.Add(err)
		mdr.Readings.RunoffPHVoltage, err = mdr.runoffPH.Read()
		mdr.errors.Add(err)
		mdr.Readings.RunoffPHRaw, err = mdr.runoffPH.Value()
		mdr.errors.Add(err)
		mdr.Readings.RunoffPH, err = mdr.tr.Translate("runoff_ph", mdr.Readings.RunoffPHRaw)
		mdr.errors.Add(err)
	}
}

func (mdr *Minder) readMoistureProbe() {
	var err error

	if mdr.moisture != nil {
		mdr.Readings.MoistureADC, err = mdr.moisture.AnalogRead()
		mdr.errors.Add(err)
		mdr.Readings.MoistureVoltage, err = mdr.moisture.Read()
		mdr.errors.Add(err)
		mdr.Readings.Moisture, err = mdr.tr.Translate("moisture", mdr.Readings.MoistureVoltage)
		mdr.errors.Add(err)
	}
}

func (mdr *Minder) readECProbes() {
	var err error

	mdr.Readings.IrrigECRaw, mdr.Readings.IrrigECTemp = mdr.bus.ProbeReadings(mdr.cfg.IrrigECProbe)
	ec, err := mdr.tr.Translate("irrig_ec", mdr.Readings.IrrigECRaw.Value())
	mdr.errors.Add(err)
	mdr.Readings.IrrigEC.SetValue(ec)
	mdr.Readings.IrrigEC.Valid = mdr.Readings.IrrigECRaw.IsValid()

	mdr.Readings.RunoffECRaw, mdr.Readings.RunoffECTemp = mdr.bus.ProbeReadings(mdr.cfg.RunoffECProbe)
	ec, err = mdr.tr.Translate("runoff_ec", mdr.Readings.RunoffECRaw.Value())
	mdr.errors.Add(err)
	mdr.Readings.RunoffEC.SetValue(ec)
	mdr.Readings.RunoffEC.Valid = mdr.Readings.RunoffECRaw.IsValid()
}

func (mdr *Minder) swapECProbes() {
	a := mdr.cfg.IrrigECProbe
	b := mdr.cfg.RunoffECProbe
	mdr.cfg.IrrigECProbe = b
	mdr.cfg.RunoffECProbe = a

	mdr.onCfgChangeCB(*mdr.cfg)
}
