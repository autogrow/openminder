package openminder

// TippingBucket models a tipping bucket for recording volume
type TippingBucket struct {
	cc *ContactClosure
}

// NewTippingBucket creates a new tipping bucket
func NewTippingBucket(pin string) (*TippingBucket, error) {
	cc, err := NewContactClosure(pin)
	return &TippingBucket{cc}, err
}

// OnTip will fire the given func when a tip is recorded
func (tb *TippingBucket) OnTip(cb func()) {
	go tb.cc.Start()
	tb.cc.OnClosure(cb)
}
