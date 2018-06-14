package aslbus

import (
	"fmt"
	"log"
	"time"

	"github.com/autogrow/openminder/types"
)

// Manager models an ASL bus manager who manages the bus
// and probes
type Manager struct {
	bus           *Bus
	LastScanStart time.Time
	LastScanDone  time.Time
	scanTimeout   int
	scanner       *Scanner
}

// NewManager creates a new ASL Bus manager that handles bus scanning and
// provides readings
func NewManager(tty string, scanTimeout int, cfgSerials ...string) *Manager {
	mgr := &Manager{}
	mgr.bus = New(tty)
	mgr.scanner = NewScanner(mgr.bus, 2, scanTimeout)

	// don't use empty strings for serials
	serials := []string{}
	for _, sn := range cfgSerials {
		if sn != "" {
			serials = append(serials, sn)
		}
	}

	log.Printf("%d known probes: %v", len(serials), serials)

	// watch for probe detections
	mgr.scanner.OnDetect(func(serial string) {
		log.Printf("detected new probe: %s", serial)
	})

	// when the bus is connected, create known probes
	mgr.bus.OnConnect(func() {
		log.Println("bus is connected")

		if len(serials) < 2 {
			serials = mgr.Scan()
		}

		for _, sn := range serials {
			log.Printf("attaching ec probe %s", sn)
			NewECProbe(sn).AttachBus(mgr.bus)
		}
	})

	return mgr
}

// Serials returns the serial numbers of the device connected to the bus
func (mgr *Manager) Serials() []string {
	return mgr.bus.Serials()
}

// OnScanDone registers a func to call when the scan function finds all the probes
// it was requested to find
func (mgr *Manager) OnScanDone(cb func([]string, error)) {
	mgr.scanner.OnScanDone(cb)
}

// ProbeReadings returns the readings for the probes
func (mgr *Manager) ProbeReadings(sn string) (*types.NullFloat, *types.NullFloat) {
	ec := &types.NullFloat{}
	temp := &types.NullFloat{}

	for _, p := range mgr.bus.probes {
		if p.SN() == sn && p.IsValid() {
			temp.SetValue(p.GetTemp())
			ec.SetValue(p.GetEC())
		}
	}

	return ec, temp
}

// Run starts the bus loop
func (mgr *Manager) Run() {
	mgr.bus.Run()
}

// OnError registers a function to call when there are any errors in the bus
func (mgr *Manager) OnError(cb func(error)) {
	mgr.bus.OnError(cb)
}

// Scan wraps the bus scan method
func (mgr *Manager) Scan() []string {
	mgr.LastScanStart = time.Now()
	log.Println("starting a bus scan")
	serials, scanned, err := mgr.scanner.Scan()
	mgr.LastScanDone = time.Now()

	if err != nil {
		mgr.bus.onErrorCB(fmt.Errorf("scan failed (%d found, %d scanned): %s", len(serials), scanned, err))
	}

	return serials
}

// Rescan will clear the probes registered on the bus before scanning it
// and re assigning any probes
func (mgr *Manager) Rescan() {
	mgr.bus.ClearProbes()
	serials := mgr.Scan()
	for _, sn := range serials {
		log.Printf("attaching ec probe %s", sn)
		NewECProbe(sn).AttachBus(mgr.bus)
	}
}
