package aslbus

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// Scanner is for scanning the bus for probes
type Scanner struct {
	bus     *Bus
	serials []string
	count   int
	timeout int
	done    bool
	scan    chan string
	lmscan  chan bool

	onScanDone func([]string, error)
	onDetectCB func(string)
}

// NewScanner returns a new scanner
func NewScanner(bus *Bus, count, timeout int) *Scanner {
	s := &Scanner{
		bus:        bus,
		count:      count,
		timeout:    timeout,
		serials:    bus.Serials(),
		scan:       make(chan string),
		lmscan:     make(chan bool),
		onScanDone: func([]string, error) {},
		onDetectCB: func(string) {},
	}

	return s
}

// Scan is a package level scanner that internally uses the scanner
func Scan(bus *Bus, count, timeout int) ([]string, int, error) {
	return NewScanner(bus, count, timeout).Scan()
}

// OnScanDone registers a func to call when the scan function finds all the probes
// it was requested to find
func (scnr *Scanner) OnScanDone(cb func([]string, error)) {
	scnr.onScanDone = cb
}

// OnDetect registers a callback to run when an unknown probe is detected
func (scnr *Scanner) OnDetect(cb func(string)) {
	scnr.onDetectCB = cb
}

func (scnr *Scanner) allFound() bool {
	return len(scnr.serials) >= scnr.count
}

func (scnr *Scanner) lastManScanner() {
	for {
		if scnr.lastManStanding() {
			scnr.lmscan <- true
		}

		time.Sleep(time.Second)
	}
}

func (scnr *Scanner) lastManStanding() bool {
	return scnr.count-len(scnr.serials) == 1
}

func (scnr *Scanner) feedSerialNumbers() {
	for maskSize := 1; maskSize <= 5; maskSize++ { // start with lowest mask
		if scnr.done || scnr.lastManStanding() {
			return
		}

		// get the possible serial numbers for the mask
		wildcardSNs := wildcardSerials(maskSize)
		for _, sn := range wildcardSNs {
			if scnr.done || scnr.lastManStanding() {
				return
			}

			scnr.scan <- sn
			time.Sleep(time.Second)
		}
	}
}

func (scnr *Scanner) periodicBroadcaster() {
	for {
		time.Sleep(10 * time.Second)
		if scnr.done || scnr.lastManStanding() {
			return
		}

		log.Println("broadcasting global pings")
		scnr.bus.master.TransmitPacket(masterAddress, pingSerial, pingCommand, "")
	}
}

func (scnr *Scanner) packetListener(pkt *Packet) {
	// if we know about this serial, ignore the packet
	for _, sn := range scnr.serials {
		if sn == pkt.serial {
			return
		}
	}

	// add the serial to the list and tell any listeners it was detected
	scnr.serials = append(scnr.serials, pkt.serial)
	scnr.onDetectCB(pkt.serial)

	// turn off pings for this serial now that we found it
	scnr.bus.disableProbePing(pkt.serial)
}

// Scan will scan the bus to find as many probes as specified in the count.  The
// detection will run until all probes are found or the given timeout (in seconds)
// is reached.
func (scnr *Scanner) Scan() ([]string, int, error) {
	scnr.done = false
	defer func() { scnr.done = true }()

	once := new(sync.Once)

	scanned := 0
	if !scnr.bus.running {
		err := fmt.Errorf("bus is not running")
		scnr.onScanDone(scnr.serials, err)
		return scnr.serials, scanned, err
	}

	// if we already have all the devices we need, bail out
	if scnr.allFound() {
		scnr.onScanDone(scnr.serials, nil)
		return scnr.serials, scanned, nil
	}

	timer := time.NewTimer(time.Duration(scnr.timeout) * time.Second)
	scnr.bus.enableAllPings()

	// scan over all possible serial numbers to send to the ping loop
	go scnr.feedSerialNumbers()

	// listen for a packet from the device
	i := scnr.bus.OnPacket(scnr.packetListener)
	defer scnr.bus.UnregisterOnPacket(i)

	// periodically re-enable broadcast pings in case someone just attached
	// a probe that was previously ping disabled
	go scnr.periodicBroadcaster()

	// watch for the last man standing condition
	go scnr.lastManScanner()

	for {
		select {
		case <-timer.C:
			err := fmt.Errorf("probe detection timed out")
			scnr.onScanDone(scnr.serials, err)
			return scnr.serials, scanned, err

		// if there's one probe left to detect, ping anyone left who still has pings enabled
		case <-scnr.lmscan:
			once.Do(func() { log.Println("one probe left, ping everybody") })
			scnr.bus.master.TransmitPacket(masterAddress, pingSerial, pingCommand, "")

		case sn := <-scnr.scan:
			// try to ping the next serial
			log.Printf("pinging %s", sn)
			scnr.bus.master.TransmitPacket(masterAddress, sn, pingCommand, "")
			scanned++

		default:
			if scnr.allFound() { // if we found all the probes get out of here
				scnr.onScanDone(scnr.serials, nil)
				return scnr.serials, scanned, nil
			}

		}
	}
}

// wildcardSerials returns the masked wildcard serials
func wildcardSerials(maskSize int) []string {
	serials := []string{}
	serial := pingSerial

	// only get as many serials as we need to for the mask size
	switch maskSize {
	case 1: // will return 10 serials
		for addrCntr := 0; addrCntr < 10; addrCntr++ {
			serial = serial[:len(serial)-1] + fmt.Sprintf("%1d", addrCntr)
			serials = append(serials, serial)
		}

	case 2: // will return 90 serials
		for addrCntr := 10; addrCntr < 100; addrCntr++ {
			serial = serial[:len(serial)-2] + fmt.Sprintf("%2d", addrCntr)
			serials = append(serials, serial)
		}

	case 3: // will return 900 serials
		for addrCntr := 100; addrCntr < 999; addrCntr++ {
			serial = serial[:len(serial)-3] + fmt.Sprintf("%3d", addrCntr)
			serials = append(serials, serial)
		}

	case 4: // will return 9000 serials
		for addrCntr := 1000; addrCntr < 9999; addrCntr++ {
			serial = serial[:len(serial)-4] + fmt.Sprintf("%4d", addrCntr)
			serials = append(serials, serial)
		}

	default:
		return serials
	}

	return serials
}
