package openminder

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// AttachAPI will attach an api to the minder so it can setup
// the appropriate endpoints
func (mdr *Minder) AttachAPI(api gin.IRouter) {
	api.GET("/errors", mdr.errorsHandler())
	api.GET("/calibrations", mdr.calibrationsHandler())
	api.PUT("/calibrations/:field/:scale/:offset", mdr.calibrateHandler())
	api.GET("/config", mdr.configHandler())
	api.GET("/readings", mdr.readingsHandler())
	api.PUT("/readings/calibrate/:field/:scale/:offset", mdr.calibrateHandler())
	api.GET("/bus", mdr.busHandler())
	api.PUT("/bus/scan", mdr.busScanHandler())
	api.PUT("/bus/swap", mdr.busSwapHandler())
}

func (mdr *Minder) busScanHandler() func(*gin.Context) {
	return func(c *gin.Context) {
		go mdr.bus.Rescan()
		c.JSON(202, map[string]string{"bus_url": "/v1/bus"})
	}
}

func (mdr *Minder) busSwapHandler() func(*gin.Context) {
	return func(c *gin.Context) {
		mdr.swapECProbes()
		c.JSON(200, map[string]string{"bus_url": "/v1/bus"})
	}
}

func (mdr *Minder) busHandler() func(*gin.Context) {
	return func(c *gin.Context) {
		var scanstart interface{}
		var scandone interface{}

		if !mdr.bus.LastScanStart.IsZero() {
			scanstart = time.Since(mdr.bus.LastScanStart).String()
		}

		if !mdr.bus.LastScanDone.IsZero() {
			scanstart = time.Since(mdr.bus.LastScanDone).String()
		}

		data := map[string]interface{}{
			"available": mdr.bus.Serials(),
			"configured": map[string]string{
				"irrig":  mdr.cfg.IrrigECProbe,
				"runoff": mdr.cfg.RunoffECProbe,
			},
			"last_scan_start": scanstart,
			"last_scan_done":  scandone,
		}

		c.JSON(200, data)
	}
}

func (mdr *Minder) errorsHandler() func(*gin.Context) {
	return func(c *gin.Context) {
		c.JSON(200, mdr.errors.Map())
	}
}

func (mdr *Minder) calibrationsHandler() func(*gin.Context) {
	return func(c *gin.Context) {
		data := map[string]calibration{}
		for _, f := range translatableFields {
			c, _ := mdr.tr.getCalibration(f)
			data[f] = c
		}

		c.JSON(200, data)
	}
}

func (mdr *Minder) configHandler() func(*gin.Context) {
	return func(c *gin.Context) {
		c.JSON(200, mdr.cfg)
	}
}

func (mdr *Minder) readingsHandler() func(*gin.Context) {
	return func(c *gin.Context) {
		c.JSON(200, mdr.Readings)
	}
}

func (mdr *Minder) calibrateHandler() func(*gin.Context) {
	return func(c *gin.Context) {
		field := c.Param("field")
		if field == "" {
			c.AbortWithStatusJSON(400, errmsg("field cannot be empty"))
			return
		}

		scale, err := strconv.ParseFloat(c.Param("scale"), 64)
		if err != nil {
			c.AbortWithStatusJSON(400, errmsg("scale must be a float"))
			return
		}

		offset, err := strconv.ParseFloat(c.Param("offset"), 64)
		if err != nil {
			c.AbortWithStatusJSON(400, errmsg("offset must be a float"))
			return
		}

		err = mdr.tr.SetCalibration(field, scale, offset)

		if err == ErrNotTranslatable {
			c.AbortWithStatusJSON(400, errmsg("that field is not translatable"))
			return
		}

		if err != nil {
			c.AbortWithError(500, err)
			return
		}

		c.Status(204)
	}
}

func errmsg(msg string) interface{} {
	return struct {
		Msg string `json:"error"`
	}{msg}
}
