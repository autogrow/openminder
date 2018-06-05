package openminder

import (
	"time"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
)

// ContactClosure models a contact switch
type ContactClosure struct {
	Pin         gpio.PinIn
	onClosureCB func()
	stop        bool
}

// NewContactClosure creates a new polling contact closure at the given pin.
// The pin should be specified in the way the periph.io/x/gpioreg expects, GPIO6, GPIO7 etc
func NewContactClosure(pin string) (*ContactClosure, error) {
	cc := &ContactClosure{}

	p := gpioreg.ByName(pin)
	if err := p.In(gpio.PullUp, gpio.NoEdge); err != nil {
		return cc, err
	}

	cc.Pin = p
	return cc, nil
}

// OnClosure takes a function to call when a contact closure is detected
func (cc *ContactClosure) OnClosure(cb func()) {
	cc.onClosureCB = cb
}

// Stop will stop the contact closure from polling the pin
func (cc *ContactClosure) Stop() {
	cc.stop = true
}

// Start will start polling the contact closure pin
func (cc *ContactClosure) Start() {
	cc.stop = false
	go cc.loop()
}

func (cc *ContactClosure) loop() {
	lowCount := 0
	for {
		if cc.stop {
			return
		}

		level := cc.Pin.Read()

		if level == gpio.Low {
			lowCount++
		}

		if lowCount >= 4 && level == gpio.High {
			cc.onClosureCB()
			lowCount = 0
		}

		time.Sleep(time.Second / 10)
	}
}
