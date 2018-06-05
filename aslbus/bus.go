package aslbus

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/jacobsa/go-serial/serial"
)

// Bus defines the master serial port object
type Bus struct {
	master            *Master
	slave             *Slave
	ReadingsChan      chan string
	onErrorCB         func(error)
	onConnectCB       func()
	onPacketCBs       []func(*Packet)
	onProbesClearedCB func()
	probes            []Probe
	running           bool
}

// New creates a new serial port master based on the config supplied
func New(tty string) *Bus {
	opts := serial.OpenOptions{
		PortName:        tty,
		BaudRate:        19200,
		DataBits:        8,
		StopBits:        2,
		ParityMode:      serial.PARITY_NONE,
		MinimumReadSize: 9,
	}

	bus := new(Bus)
	bus.master = NewMaster(opts)
	rxChan := make(chan string)
	bus.ReadingsChan = rxChan
	bus.slave = NewSlave(opts, rxChan)

	// blank out all the callbacks
	bus.onErrorCB = func(err error) {}
	bus.onProbesClearedCB = func() {}
	bus.onConnectCB = func() {}

	return bus
}

// Transmit - used to transmit data out on the bus
func (bus *Bus) Transmit(addr, serial, cmd, data string) {
	bus.master.TransmitPacket(addr, serial, cmd, data)
}

func (bus *Bus) enableAllPings() {
	log.Printf("enable all pings")
	bus.master.TransmitPacket(masterAddress, pingSerial, enablePingCommand, "")
}

func (bus *Bus) disableAllPings() {
	log.Printf("disable all pings")
	bus.master.TransmitPacket(masterAddress, pingSerial, disablePingCommand, "")
}

func (bus *Bus) disableProbePing(serial string) {
	log.Printf("disable ping for %s", serial)
	bus.master.TransmitPacket(masterAddress, serial, disablePingCommand, "")
}

// Run starts the master and slave loops and waits for packets so it can
// porcess these into the correct device.
func (bus *Bus) Run() error {
	for {
		if bus.slave == nil {
			bus.onErrorCB(fmt.Errorf("no bus slave attached"))
			time.Sleep(time.Second / 2)
			continue
		}

		if bus.master == nil {
			bus.onErrorCB(fmt.Errorf("no bus master attached"))
			time.Sleep(time.Second / 2)
			continue
		}

		break
	}

	go bus.slave.Listen()
	go bus.master.Run()

	bus.running = true
	defer func() { bus.running = false }()
	go bus.onConnectCB()

	// Maintain open port
	for {
		newPkt, ok := <-bus.ReadingsChan
		if !ok {
			// Handle Error
			bus.slave.Quit()
			rxChan := make(chan string)
			bus.ReadingsChan = rxChan
			bus.slave.rxChan = rxChan
			go bus.slave.Listen()
			continue
		}

		err := bus.processPacket(newPkt)
		if err != nil {
			bus.onErrorCB(fmt.Errorf("processing packet failed: %s", err))
		}
	}
}

func (bus *Bus) processPacket(newPkt string) error {
	pkt, err := NewRxPkt(newPkt)

	if err != nil {
		if strings.Contains(err.Error(), "from a master") { // ignore tx packets
			return nil
		}
		return err
	}

	switch pkt.address {
	case ecProbeAddress:
		go bus.sendPacket(pkt)
		return nil
	}

	addrInt, err := strconv.ParseInt(pkt.address, 10, 64)
	return fmt.Errorf("packet from un-supported device (%s) %d,%s recieved", pkt.address, addrInt, err)
}

func (bus *Bus) sendPacket(pkt *Packet) {
	var sent = true

	for _, p := range bus.probes {
		if p.SN() == pkt.serial {
			err := p.Update(pkt)
			if err != nil {
				bus.onErrorCB(fmt.Errorf("failed to send packet to %s: %s", pkt.serial, err))
				sent = false
			}
		}
	}

	for _, cb := range bus.onPacketCBs {
		if cb != nil {
			go cb(pkt)
		}
	}

	if !sent {
		bus.onErrorCB(fmt.Errorf("sent package for unregistered probe: %s", pkt.serial))
	}
}

// Probes returns the probes registered to the this bus
func (bus *Bus) Probes() []Probe {
	return bus.probes
}

// Serials will return the serial numbers of all registered probes
func (bus *Bus) Serials() []string {
	var sns []string
	for _, p := range bus.probes {
		sns = append(sns, p.SN())
	}

	return sns
}

// ClearProbes will clear all probes by calling their DetachBus method
func (bus *Bus) ClearProbes() {
	for _, p := range bus.probes {
		p.DetachBus() // this stops the probe and unregisters it from the bus
	}

	bus.probes = []Probe{}
	bus.onProbesClearedCB()
}

// HasProbe will return true if the probe with the given serial has been registered
func (bus *Bus) HasProbe(serial string) bool {
	var have bool
	for _, p := range bus.probes {
		if p.SN() == serial {
			have = true
		}
	}

	return have
}

func (bus *Bus) unregisterProbe(serial string) {
	log.Println("unregisterProbe", serial, len(bus.probes))
	for i, _p := range bus.probes {
		log.Printf("i: %d/%d  p %v", i, len(bus.probes), _p)
		if _p == nil || serial == _p.SN() {
			log.Println("got him!", len(bus.probes))

			bus.probes[i] = bus.probes[len(bus.probes)-1]
			bus.probes[len(bus.probes)-1] = nil
			bus.probes = bus.probes[:len(bus.probes)-1]
		}
	}
}

func (bus *Bus) registerProbe(p Probe) {
	if bus.HasProbe(p.SN()) {
		return
	}

	bus.probes = append(bus.probes, p)
}

// OnConnect takes a function to call when the bus is successfully connected
// to both the master and slave ports
func (bus *Bus) OnConnect(cb func()) {
	bus.onConnectCB = cb
}

// OnProbesCleared will register a function to be called when the probes are cleared
// (likely prior to a detection)
func (bus *Bus) OnProbesCleared(cb func()) {
	bus.onProbesClearedCB = cb
}

// OnError takes a function to call when an error is detected
func (bus *Bus) OnError(cb func(error)) {
	bus.onErrorCB = cb
}

// OnPacket registers a func to call when a packet is received
func (bus *Bus) OnPacket(cb func(*Packet)) int {
	bus.onPacketCBs = append(bus.onPacketCBs, cb)
	return len(bus.onPacketCBs) - 1
}

// UnregisterOnPacket will unregister an on packet callback by the given index
func (bus *Bus) UnregisterOnPacket(i int) {
	bus.onPacketCBs[i] = nil
}
