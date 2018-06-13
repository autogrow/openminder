package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"bitbucket.org/autogrowsystems/openminder"
	"bitbucket.org/autogrowsystems/openminder/aslbus"
	calibutil "bitbucket.org/autogrowsystems/openminder/calib"
)

var version = "1.0.0"

func main() {
	var calibDef, port, cfgFile string
	var printReadings, calib, ecProbe, phProbe, moistureProbe, runoffSide, irrigSide, printVersion, detectProbes, scanbus bool
	var ecBuffer float64

	flag.BoolVar(&calib, "calib", false, "calibrate something")
	flag.StringVar(&calibDef, "set", "", "set reading calibration: reading,scale,offset (e.g. runoff_volume,5.0,0)")
	flag.Float64Var(&ecBuffer, "buffer", 2.77, "EC buffer")
	flag.BoolVar(&ecProbe, "ec", false, "calibrate an EC probe")
	flag.BoolVar(&phProbe, "ph", false, "calibrate a pH probe")
	flag.BoolVar(&moistureProbe, "moisture", false, "calibrate a moisture probe")
	flag.BoolVar(&runoffSide, "runoff", false, "calibrate a probe for the runoff side")
	flag.BoolVar(&irrigSide, "irrig", false, "calibrate a probe for the irrig side")
	flag.BoolVar(&printReadings, "readings", false, "print readings")
	flag.BoolVar(&detectProbes, "detectprobes", false, "start the probe detection wizard")
	flag.BoolVar(&scanbus, "scanbus", false, "scan the bus for probes wihout saving to config")
	flag.StringVar(&port, "p", "3232", "the port to talk to the API on")
	flag.StringVar(&cfgFile, "c", "", "the config file to use/write to")
	flag.BoolVar(&printVersion, "v", false, "print the version")
	flag.Parse()

	client := openminder.NewClient("http://localhost:" + port + "/" + apiVersion())

	switch {
	case printVersion:
		fmt.Println(version)
		os.Exit(0)

	case printReadings:
		r, err := client.Readings()
		if err != nil {
			panic(err)
		}
		dumpJSONR(r)

	case scanbus:
		if err := scanProbes(cfgFile); err != nil {
			log.Fatalf("ERROR: failed to scan probes: %s", err)
		}

	case detectProbes:
		if err := detectProbesWizard(cfgFile); err != nil {
			log.Fatalf("failed to complete probe detection: %s", err)
		}

	case calib && calibDef != "":
		bits := strings.Split(calibDef, ",")
		if len(bits) != 3 {
			log.Fatalf("ERROR: calib must be specified as reading,scale,offset")
		}

		s, err := strconv.ParseFloat(bits[1], 64)
		if err != nil {
			log.Fatalf("ERROR: %s", err)
		}

		o, err := strconv.ParseFloat(bits[2], 64)
		if err != nil {
			log.Fatalf("ERROR: %s", err)
		}

		err = client.SetCalibration(bits[0], s, o)
		if err != nil {
			log.Fatalf("ERROR: %s", err)
		}

	case calib && !runoffSide && !irrigSide:
		log.Fatalf("must specify the side to calibrate with -runoff or -irrig")

	case calib && ecProbe:
		if err := calibrateEC(client, runoffSide, ecBuffer); err != nil {
			log.Fatalf("ERROR: %s", err)
		}

	case calib && phProbe:
		if err := calibratePH(client, runoffSide); err != nil {
			log.Fatalf("ERROR: %s", err)
		}

	case calib && moistureProbe:
		if err := calibrateMoisture(client); err != nil {
			log.Fatalf("ERROR: %s", err)
		}

	default:
		flag.Usage()
		os.Exit(1)

	}

}

func apiVersion() string {
	return "v" + strings.Split(version, ".")[0]
}

func dumpJSONR(r interface{}) error {
	data, err := json.MarshalIndent(r, "", "  ")

	if err != nil {
		return err
	}

	fmt.Println(string(data))
	return nil
}

func calibrateEC(client *openminder.Client, runoffSide bool, ecBuffer float64) error {
	fmt.Printf("wash the probe and put it in the %0.2f buffer solution, then push enter...\n", ecBuffer)
	waitForEnter()
	for i := 30; i > 0; i-- {
		fmt.Print("\rwaiting for the readings to settle... ", i, " ")
		time.Sleep(time.Second)
	}

	fmt.Println()
	fmt.Println("taking reading...")

	r, err := client.Readings()
	if err != nil {
		return err
	}

	side, reading := func() (string, float64) {
		switch runoffSide {
		case true:
			return "runoff", r.RunoffECRaw.Value()
		default:
			return "irrig", r.IrrigECRaw.Value()
		}
	}()

	fmt.Printf("reading was %0.2f EC\n", reading)

	scale, offset, err := calibutil.EC(ecBuffer, reading)
	if err != nil {
		return err
	}

	fmt.Printf("calibrated probe, scale=%0.2f offset=%0.2f - save this? [Y/n]: ", scale, offset)
	if !waitForAnswer(true) {
		fmt.Println("not saving...")
		os.Exit(0)
	}

	err = client.SetCalibration(side+"_ec", scale, offset)
	if err != nil {
		return err
	}

	fmt.Println("saved")

	return nil
}

func calibratePH(client *openminder.Client, runoffSide bool) error {
	fmt.Println("wash the probe and put it in the pH7 buffer solution, then push enter...")
	waitForEnter()
	for i := 30; i > 0; i-- {
		fmt.Print("\rwaiting for the readings to settle... ", i, " ")
		time.Sleep(time.Second)
	}

	fmt.Println()
	fmt.Println("taking reading...")

	r, err := client.Readings()
	if err != nil {
		return err
	}

	side, ph7 := func() (string, float64) {
		switch runoffSide {
		case true:
			return "runoff", r.RunoffPHRaw
		default:
			return "irrig", r.IrrigPHRaw
		}
	}()
	fmt.Printf("reading was %0.2f pH\n", ph7)

	fmt.Println("wash the probe and put it in the pH4 buffer solution, then push enter...")
	waitForEnter()
	for i := 30; i > 0; i-- {
		fmt.Print("\rwaiting for the readings to settle... ", i, " ")
		time.Sleep(time.Second)
	}

	fmt.Println()
	fmt.Println("taking reading...")

	r, err = client.Readings()
	if err != nil {
		return err
	}

	ph4 := func() float64 {
		switch runoffSide {
		case true:
			return r.RunoffPHRaw
		default:
			return r.IrrigPHRaw
		}
	}()
	fmt.Printf("reading was %0.2f pH\n", ph4)

	fmt.Println("calibrating...")

	scale, offset, err := calibutil.PH(ph7, ph4)
	if err != nil {
		return err
	}

	fmt.Printf("calibrated probe, scale=%0.2f offset=%0.2f - save this? [Y/n]: ", scale, offset)
	if !waitForAnswer(true) {
		fmt.Println("not saving...")
		os.Exit(0)
	}

	err = client.SetCalibration(side+"_ph", scale, offset)
	if err != nil {
		return err
	}

	fmt.Println("saved")

	return nil
}

func calibrateMoisture(client *openminder.Client) error {
	fmt.Println("make sure the probe is completely dry or in dry media, then push enter...")
	waitForEnter()
	for i := 30; i > 0; i-- {
		fmt.Print("\rwaiting for the readings to settle... ", i, " ")
		time.Sleep(time.Second)
	}

	fmt.Println()
	fmt.Println("taking reading...")

	r, err := client.Readings()
	if err != nil {
		return err
	}

	offset := r.MoistureVoltage
	fmt.Printf("reading was %0.3f volts\n", offset)

	fmt.Println("place probe in water or completely saturated media, then push enter...")
	waitForEnter()
	for i := 30; i > 0; i-- {
		fmt.Print("\rwaiting for the readings to settle... ", i, " ")
		time.Sleep(time.Second)
	}

	fmt.Println()
	fmt.Println("taking reading...")

	r, err = client.Readings()
	if err != nil {
		return err
	}

	scale := r.MoistureVoltage
	fmt.Printf("reading was %0.3f pH\n", scale)

	fmt.Printf("calibrated probe, dry voltage=%0.3f wet voltage=%0.3f - save this? [Y/n]: ", scale, offset)
	if !waitForAnswer(true) {
		fmt.Println("not saving...")
		os.Exit(0)
	}

	err = client.SetCalibration("moisture", scale, offset)
	if err != nil {
		return err
	}

	fmt.Println("saved")

	return nil
}

func waitForEnter() {
	reader := bufio.NewReader(os.Stdin)
	_, _ = reader.ReadString('\n')
}

func waitForAnswer(dflt bool) bool {
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	text = strings.TrimSpace(text)
	text = strings.ToLower(text)

	if text == "" {
		return dflt
	}

	switch text {
	case "y", "yes":
		return true
	}
	return false
}

func detectProbesWizard(cfgFile string) error {
	var irrigSN, runoffSN string

	if cfgFile == "" {
		return fmt.Errorf("to detect the probes you need to specify the config file to save them to using -c")
	}

	cfg := new(openminder.Config)
	if err := cfg.LoadFrom(cfgFile); err != nil {
		return fmt.Errorf("failed to read config: %s", err)
	}

	bus := aslbus.New(cfg.TTY)

	fmt.Println("unplug all probes and push enter")
	waitForEnter()
	fmt.Println("plug in the probe to be used on the runoff side and push enter")
	waitForEnter()

	runoffSN = detectOneProbe(bus)

	fmt.Println("found EC probe with serial number", runoffSN)

	fmt.Println("plug in the probe to be used on the irrigation side and push enter")
	waitForEnter()

	irrigSN = detectOneProbe(bus)

	fmt.Println("found EC probe with serial number", irrigSN)

	fmt.Println("save these probes to your config? [Y/n]: ")
	if shouldSave := waitForAnswer(true); !shouldSave {
		fmt.Println("not saving...")
		return nil
	}

	cfg.IrrigECProbe = irrigSN
	cfg.RunoffECProbe = runoffSN
	if err := cfg.SaveTo(cfgFile); err != nil {
		return fmt.Errorf("failed to save config: %s", err)
	}

	fmt.Println("saved probes to config file", cfgFile)

	return nil
}

func detectOneProbe(bus *aslbus.Bus) string {
	sns, _, err := aslbus.Scan(bus, 1, 60)
	if err != nil {
		fmt.Println("ERROR:", err)
		os.Exit(1)
	}

	if len(sns) == 0 {
		fmt.Println("ERROR: no probes detected")
		os.Exit(1)
	}

	if len(sns) > 1 {
		fmt.Println("ERROR: more that one probe detected")
		os.Exit(1)
	}

	return sns[0]
}

func scanProbes(cfgFile string) error {
	cfg := new(openminder.Config)
	if err := cfg.LoadFrom(cfgFile); err != nil {
		return fmt.Errorf("failed to read config: %s", err)
	}

	log.Printf("using port %s", cfg.TTY)
	bus := aslbus.New(cfg.TTY)
	scanner := aslbus.NewScanner(bus, 2, 60)

	bus.OnConnect(func() {
		log.Printf("bus connected")
		sns, scanned, err := scanner.Scan()
		log.Printf("%d / 2 probes found of %d scanned: error(%v)", len(sns), scanned, err)
	})

	bus.OnError(func(err error) {
		log.Printf("ERROR: %s", err)
	})

	scanner.OnScanDone(func(sns []string, err error) {
		log.Printf("scan complete: %v", sns)
		os.Exit(0)
	})

	bus.Run()
	return nil
}
