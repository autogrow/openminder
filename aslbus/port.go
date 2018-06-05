package aslbus

import (
	"io"

	"github.com/jacobsa/go-serial/serial"
)

const (
	portClosed = 0
	portOpen   = 1
)

// SerialPort wrapper for the jacobsa go-serial so the port can handle open and close better
type SerialPort struct {
	Options serial.OpenOptions
	Port    io.ReadWriteCloser
	State   int
}

// NewPort returns a serial port with the parameters specified
func NewPort(options serial.OpenOptions) *SerialPort {
	port := &SerialPort{
		Options: options,
		State:   portClosed,
	}

	return port
}

// Open - issued to open a serial port
func (s *SerialPort) Open() error {
	var err error
	s.Port, err = serial.Open(s.Options)
	if err != nil {
		s.State = portClosed
		return err
	}

	s.State = portOpen
	return nil
}

// IsClosed - returns a true if the serial port is closed
func (s *SerialPort) IsClosed() bool {
	return s.State == portClosed
}

// IsOpen - returns a true if the serial port is open
func (s *SerialPort) IsOpen() bool {
	return s.State == portOpen
}

// Close - closes the serial port
func (s *SerialPort) Close() {
	if s.Port != nil {
		s.Port.Close()
	}
	s.State = portClosed
}
