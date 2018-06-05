package openminder

// CalculateRunoffRatio will use the readings and config to calculate the runoff ratio
func CalculateRunoffRatio(r *Readings, cfg Config) {
	if r.IrrigVolume == 0 {
		return
	}

	if cfg.RunoffDrippers == 0 || cfg.IrrigDrippers == 0 || cfg.DrippersPerPlant == 0 {
		r.RunoffRatio = float64(r.RunoffVolume) / float64(r.IrrigVolume)
		return
	}

	iVol := (r.IrrigVolume / float64(cfg.IrrigDrippers)) * float64(cfg.DrippersPerPlant)
	rVol := (r.RunoffVolume / float64(cfg.RunoffDrippers)) * float64(cfg.DrippersPerPlant)

	r.RunoffRatio = rVol / iVol
}
