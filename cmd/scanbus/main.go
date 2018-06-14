package main

import (
	"flag"
	"log"
	"os"

	"github.com/autogrow/openminder/aslbus"
)

func main() {
	var port string
	var verbose bool

	flag.StringVar(&port, "port", "/dev/ttyUSB0", "port to use")
	flag.BoolVar(&verbose, "v", false, "verbose")
	flag.Parse()

	log.Printf("using port %s", port)
	bus := aslbus.New(port)
	scanner := aslbus.NewScanner(bus, 2, 60)

	bus.OnConnect(func() {
		log.Printf("bus connected")
		sns, scanned, err := scanner.Scan()
		log.Printf("%d probes found of %d scanned: error(%v)", len(sns), scanned, err)
	})

	if verbose {
		bus.OnPacket(func(pkt *aslbus.Packet) {
			log.Printf("packet received: %s", string(pkt.Bytes()))
		})
	}

	bus.OnError(func(err error) {
		log.Printf("ERROR: %s", err)
	})

	var probes []*aslbus.ECProbe
	scanner.OnDetect(func(sn string) {
		log.Printf("probe detected: %s", sn)
		probes = append(probes, aslbus.NewECProbe(sn).AttachBus(bus))
	})

	scanner.OnScanDone(func(sns []string, err error) {
		log.Printf("scan complete: %v", sns)
		os.Exit(0)
	})

	bus.Run()
}
