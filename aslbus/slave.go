package aslbus

import (
	"bufio"
	"time"

	"github.com/jacobsa/go-serial/serial"
)

// Slave defines the slave serial port object
type Slave struct {
	running bool
	rxChan  chan string
	port    *SerialPort
}

// NewSlave creates a new serial port slave based on the supplied config
func NewSlave(options serial.OpenOptions, rxChan chan string) *Slave {
	return &Slave{false, rxChan, NewPort(options)}
}

// Running - returns a true if the slave is running its listen loop
func (s *Slave) Running() bool {
	return s.running
}

// Quit - closes the current slave
func (s *Slave) Quit() {
	s.running = false
}

// IsOpen returns if the port is open
func (s *Slave) IsOpen() bool {
	return s.port.IsOpen()
}

// Close closes the slave serial port
func (s *Slave) Close() {
	s.port.Close()
}

// Closes the  port
func (s *Slave) stop() {
	if s.port.IsOpen() {
		s.port.Close()
	}
}

// Listen read incoming messages from the serial port, also closes the port
// and re-opens if any errors occur
func (s *Slave) Listen() {
	ticker := time.NewTicker(1 * time.Second)

	s.running = true
	defer s.stop()

	for {
		<-ticker.C

		if !s.running {
			return
		}

		if s.port.IsClosed() {
			if err := s.port.Open(); err != nil {
				continue
			}
		}

		reader := bufio.NewReaderSize(s.port.Port, 10240)

		for {
			if !s.running {
				return
			}

			// wait to read a packet
			reply, err := reader.ReadString(pktEOF) // EOF is excluded

			if err != nil {
				reader.Reset(s.port.Port)
				break
			}

			go func() {
				s.rxChan <- reply
			}()
		}
	}
}
