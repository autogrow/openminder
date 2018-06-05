package aslbus

import (
	"fmt"
	"time"

	"github.com/jacobsa/go-serial/serial"
)

// Master - object struct for an ASL bus master
type Master struct {
	running    bool
	port       *SerialPort
	TxChannel  chan *Packet
	quit       chan bool
	pingActive bool
	txRequests int
	txSent     int
}

// NewMaster - returns a new asl bus master object
func NewMaster(options serial.OpenOptions) *Master {
	return &Master{
		false,
		NewPort(options),
		make(chan *Packet),
		make(chan bool),
		false,
		0,
		0,
	}
}

// Quit - closes the current slave
func (m *Master) Quit() {
	m.running = false
}

// Stop - stops the master transmit queue, run will need to be called again to start it
func (m *Master) stop() {
	if m.port.IsOpen() {
		m.port.Close()
	}
}

// TransmitPacket - creates a packet from the addr, command and paylod specifed and pushes it to the transmit queue
func (m *Master) TransmitPacket(address, serial, command, payload string) {
	tx := NewTxPkt(address, serial, command, payload)
	// fmt.Println("Transmit this: ", serial, " : ", address, " : ", command, " : ", payload)
	m.txRequests++
	if m.running {
		go func() {
			m.TxChannel <- tx
			m.txSent++
		}()
	}
}

// Run - starts the bus running - needs to be a go call
func (m *Master) Run() {
	// Maintain open port
	txTicker := time.NewTicker(1 * time.Second)
	m.running = true

	defer m.stop()
	txBuffer := NewFIFO(100)

	m.txRequests = 0
	m.txSent = 0

	for {
		select {
		case pkt, ok := <-m.TxChannel:
			if !ok {
				m.TxChannel = make(chan *Packet)
				continue
			}
			txBuffer.Push(pkt)

		case <-txTicker.C:
			if txBuffer.Length() > 0 {
				txPkt := txBuffer.Pop().(*Packet)
				m.transmit(txPkt)
			}
			if !m.running {
				return
			}
		}
	}
}

func (m *Master) transmit(packet *Packet) error {
	if m.port.IsClosed() {
		if err := m.port.Open(); err != nil {
			return err
		}
	}

	n, err := m.port.Port.Write(packet.Bytes())

	if err != nil {
		m.port.Close()
		return err
	} else if n == 0 {
		m.port.Close()
		return fmt.Errorf("No bytes written to port")
	}
	return nil
}
