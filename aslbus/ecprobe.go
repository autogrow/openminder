package aslbus

import (
	"fmt"
	"log"
	"time"
)

const (
	ecProbeAddress = "x"
	ecProbeDevice  = "ec_probe"
)

// ErrProbeNotAttached - message gerenated when the probe is not attached
var ErrProbeNotAttached = fmt.Errorf("probe not attached to any bus")

// Probe is an interface for the EC probe struct
type Probe interface {
	Update(*Packet) error
	Start() error
	DetachBus()
	Stop()
	SN() string
	IsValid() bool
	GetTemp() float64
	GetEC() float64
}

// ECProbe represents an Autogrow Smart EC probe
type ECProbe struct {
	Serial          string `json:"address"`
	LastSeen        int64  `json:"last_seen"`
	master          *Master
	bus             *Bus
	running         bool
	quit            chan bool
	EC              float64 `json:"ec"`
	ECReal          float64 `json:"ec_real"`
	Temp            float64 `json:"temp"`
	TempReal        float64 `json:"temp_real"`
	FirmwareVersion string  `json:"firmware_version"`
}

// NewECProbe - returns a pointer for the device with the serial number specified
func NewECProbe(serial string) *ECProbe {
	return &ECProbe{
		Serial: serial,
		quit:   make(chan bool),
	}
}

func (d *ECProbe) attachMaster(master *Master) *ECProbe {
	d.master = master
	return d
}

// GetTemp will return the current temperature of the probe
func (d *ECProbe) GetTemp() float64 {
	return d.Temp
}

// GetEC will return the current EC of the probe
func (d *ECProbe) GetEC() float64 {
	return d.EC
}

// DetachBus will stop the readings loop
func (d *ECProbe) DetachBus() {
	d.Stop()
	for {
		if !d.running {
			break
		}

		time.Sleep(time.Second / 10)
	}

	log.Println("probe unregisterProbe", d.Serial, d)
	d.bus.unregisterProbe(d.Serial)
}

// AttachBus will attach the given bus to the EC probe
func (d *ECProbe) AttachBus(bus *Bus) *ECProbe {
	d.bus = bus

	d.attachMaster(bus.master)

	// start the probe readings in a thread
	go func() {
		if err := d.Start(); err != nil {
			bus.onErrorCB(fmt.Errorf("probe %s failed to start: %s", d.Serial, err))
		}
	}()

	// register the probe with the bus
	bus.registerProbe(d)
	return d
}

// Update the probe readings with an rx packet
func (d *ECProbe) Update(pkt *Packet) error {
	if pkt.serial != d.Serial {
		log.Printf("WARN: packet for %s recieved by %s", pkt.serial, d.Serial)
		return nil // ignore packets not for this device
	}

	d.LastSeen = pkt.timestamp

	if pkt.cmd != readingCommand {
		return nil
	}

	// Decode Payload
	if len(pkt.data) == 0 {
		return fmt.Errorf("Readings packet contains no data")
	}

	d.process(pkt.data)
	return nil
}

// Quit - issues a stop condition to the loop
func (d *ECProbe) Quit() {
	d.Stop()
}

func (d *ECProbe) enablePings() {
	d.master.TransmitPacket(ecProbeAddress, d.Serial, enablePingCommand, "")
}

func (d *ECProbe) disablePings() {
	d.master.TransmitPacket(ecProbeAddress, d.Serial, disablePingCommand, "")
}

// Start will setup the quit chan and start the interrogation loop.  An error will be
// returned if there were problems starting the loop
func (d *ECProbe) Start() error {
	d.quit = make(chan bool, 1)
	return d.interrogate(5, 1)
}

// Stop will close the quit chan triggering the interrogation loop to bail
func (d *ECProbe) Stop() {
	close(d.quit)
	d.quit = nil
}

// SN returns the serial number of the device
func (d *ECProbe) SN() string {
	return d.Serial
}

func (d *ECProbe) requestReading() error {
	if d.master == nil {
		return ErrProbeNotAttached
	}

	d.master.TransmitPacket(ecProbeAddress, d.Serial, readingCommand, "")
	return nil
}

// interrogate request readings and if readings stop then try to ping
func (d *ECProbe) interrogate(every, retryCnt int) error {
	if d.master == nil {
		return ErrProbeNotAttached
	}

	ticker := time.NewTicker(time.Duration(every) * time.Second)
	d.running = true
	defer func() { d.running = false }()

	for {
		select {
		case <-ticker.C:
			if err := d.requestReading(); err != nil {
				return err
			}

		case _, _ = <-d.quit:
			return nil
		}
	}
}

// IsValid returns true if the probe has been seen in the
// last 2 minutes
func (d *ECProbe) IsValid() bool {
	if (time.Now().Unix() - d.LastSeen) > 120 {
		return false
	}
	return true
}

var packetPayloadFormat = []byteDef{
	{"asl_status", clByte},
	{"spare1", clByte},
	{"firmware_version_lo", clByte},
	{"firmware_version_hi", clByte},
	{"status_bool0_lo", clByte},
	{"status_bool0_hi", clByte},
	{"status_bool1_lo", clByte},
	{"status_bool1_hi", clByte},
	{"ec_lo", clByte},
	{"ec_hi", clByte},
	{"temp_lo", clByte},
	{"temp_hi", clByte},
	{"spare2", clLong},
	{"spare3", clLong},
	{"spare4", clLong},
	{"spare5", clLong},
	{"spare6", clLong},
	{"spare7", clLong},
	{"spare8", clLong},
	{"spare9", clLong},
	{"spare10", clLong},
	{"spare11", clLong},
	{"spare12", clWord},
	{"ec_real_0", clByte},
	{"ec_real_1", clByte},
	{"ec_real_2", clByte},
	{"ec_real_3", clByte},
	{"temp_real_0", clByte},
	{"temp_real_1", clByte},
	{"temp_real_2", clByte},
	{"temp_real_3", clByte},
	{"spare13", clByte},
	{"sig_pc", clByte},
}

func (d *ECProbe) process(data string) {
	raw := d.convertData(data)

	// Firmware Version
	fw := raw["firmware_version_lo"] + (raw["firmware_version_hi"] << 8)
	d.FirmwareVersion = fmt.Sprintf("V%.2f", float64(fw)/100)

	// EC
	ec := raw["ec_lo"] + (raw["ec_hi"] << 8)
	d.EC = float64(ec) / 100.0
	ecReal := raw["ec_real_0"] + (raw["ec_real_1"] << 8) + (raw["ec_real_2"] << 16) + (raw["ec_real_3"] << 24)
	d.ECReal = float64(ecReal)

	// Temp
	temp := raw["temp_lo"] + (raw["temp_hi"] << 8)
	d.Temp = float64(temp) / 100.0
	tempReal := raw["temp_real_0"] + (raw["temp_real_1"] << 8) + (raw["temp_real_2"] << 16) + (raw["temp_real_3"] << 24)
	d.TempReal = float64(tempReal)
}

func (d *ECProbe) convertData(data string) map[string]int {
	index := 0
	out := make(map[string]int)
	for _, bytedef := range packetPayloadFormat {
		end := index + bytedef.l
		if end > len(data) {
			break
		}
		out[bytedef.n] = Hextobin(data[index:end])
		index += bytedef.l
	}
	return out
}
