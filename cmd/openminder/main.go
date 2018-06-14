package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/autogrow/openminder"
	"github.com/gin-gonic/gin"
	"periph.io/x/periph/host"
)

var version = "1.0.0"

func main() {
	cfg := &openminder.Config{}
	var cfgFile string
	var printVersion bool

	flag.StringVar(&cfg.IrrigTBGPIO, "tb1", "GPIO5", "pin for irrigation tipping bucket")
	flag.StringVar(&cfg.RunoffTBGPIO, "tb2", "GPIO6", "pin for runoff tipping bucket")
	flag.StringVar(&cfg.Port, "p", "3232", "the port to serve the API on")
	flag.StringVar(&cfgFile, "c", "", "path to the config file to use")
	flag.BoolVar(&printVersion, "v", false, "print the version")
	flag.Parse()

	if printVersion {
		fmt.Println(version)
		os.Exit(0)
	}

	if cfgFile != "" {
		if err := cfg.LoadFrom(cfgFile); err != nil {
			panic(err)
		}
	}

	if _, err := host.Init(); err != nil {
		panic(err)
	}

	api := gin.Default()
	r := api.Group("/" + apiVersion())

	minder, err := openminder.NewMinder(cfg)
	if err != nil {
		panic(err)
	}
	go minder.Start()

	minder.OnConfigChange(func(cfg openminder.Config) {
		err := cfg.SaveTo(cfgFile)
		if err != nil {
			log.Printf("ERROR: failed to update config file: %s", err)
		} else {
			log.Printf("updated config file")
		}
	})

	minder.AttachAPI(r)
	api.Run(":" + cfg.Port)
}

func apiVersion() string {
	return "v" + strings.Split(version, ".")[0]
}
