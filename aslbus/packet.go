package aslbus

import (
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

const (
	// ClByte defines the size in chars of a byte
	clByte = 2
	// ClByte defines the size in chars of a word
	clWord = 4
	// ClByte defines the size in chars of a long
	clLong = 8

	pktHeaderSize = 19

	pktAddrPosStart = 1
	pktAddrPosEnd   = 2

	pktSerialPosStart = 2
	pktSerialPosEnd   = 15

	pktCmdPosStart = 15
	pktCmdPosEnd   = 17

	pktDataCntLoPosStart = 17
	pktDataCntLoPosEnd   = 19
	pktDataCntHiPosStart = 19
	pktDataCntHiPosEnd   = 21

	pktDataPosStart = 21

	minimumPktSize = 25

	masterAddress = "!"

	pingSerial         = "ASL!!!!!!!!!!"
	pingCommand        = "$0"
	enablePingCommand  = "F0"
	disablePingCommand = "N0"
	readingCommand     = "r0"
	pktSOF             = ":"
	pktEOF             = '\x04'
	preAmble           = "UU"
)

var (
	// ErrMasterPacket - error message returned when a master packet is detected
	ErrMasterPacket = fmt.Errorf("packet is from a master don't process")
	// ErrNoStartChar - error message returned when a packet doesn't have a start character
	ErrNoStartChar = fmt.Errorf("packet doesn't contain start character")
	// ErrTooShort - error message returned when a packet doesn't have a full length header
	ErrTooShort = fmt.Errorf("packet too short does not contain a full length header")
	// ErrNoCRC - error message returned when a packet doesn't have a valid CRC
	ErrNoCRC = fmt.Errorf("packet too short does not contain a CRC")
	// ErrSizeMismatch - error message returned when a packet has mismatch data count with length
	ErrSizeMismatch = fmt.Errorf("packet too short mismatch with data count specified")
	// ErrInvalidChar - error message returned when a packet contains a master character
	ErrInvalidChar = fmt.Errorf("packet serial contains ! this is only issued by masters")
)

// byteDef a slice of data must represent a reading in byte, with name n and lenght l
type byteDef struct {
	n string
	l int
}

var crcTable = [...]uint16{
	0x0000, 0xC0C1, 0xC181, 0x0140, 0xC301, 0x03C0, 0x0280, 0xC241,
	0xC601, 0x06C0, 0x0780, 0xC741, 0x0500, 0xC5C1, 0xC481, 0x0440,
	0xCC01, 0x0CC0, 0x0D80, 0xCD41, 0x0F00, 0xCFC1, 0xCE81, 0x0E40,
	0x0A00, 0xCAC1, 0xCB81, 0x0B40, 0xC901, 0x09C0, 0x0880, 0xC841,
	0xD801, 0x18C0, 0x1980, 0xD941, 0x1B00, 0xDBC1, 0xDA81, 0x1A40,
	0x1E00, 0xDEC1, 0xDF81, 0x1F40, 0xDD01, 0x1DC0, 0x1C80, 0xDC41,
	0x1400, 0xD4C1, 0xD581, 0x1540, 0xD701, 0x17C0, 0x1680, 0xD641,
	0xD201, 0x12C0, 0x1380, 0xD341, 0x1100, 0xD1C1, 0xD081, 0x1040,
	0xF001, 0x30C0, 0x3180, 0xF141, 0x3300, 0xF3C1, 0xF281, 0x3240,
	0x3600, 0xF6C1, 0xF781, 0x3740, 0xF501, 0x35C0, 0x3480, 0xF441,
	0x3C00, 0xFCC1, 0xFD81, 0x3D40, 0xFF01, 0x3FC0, 0x3E80, 0xFE41,
	0xFA01, 0x3AC0, 0x3B80, 0xFB41, 0x3900, 0xF9C1, 0xF881, 0x3840,
	0x2800, 0xE8C1, 0xE981, 0x2940, 0xEB01, 0x2BC0, 0x2A80, 0xEA41,
	0xEE01, 0x2EC0, 0x2F80, 0xEF41, 0x2D00, 0xEDC1, 0xEC81, 0x2C40,
	0xE401, 0x24C0, 0x2580, 0xE541, 0x2700, 0xE7C1, 0xE681, 0x2640,
	0x2200, 0xE2C1, 0xE381, 0x2340, 0xE101, 0x21C0, 0x2080, 0xE041,
	0xA001, 0x60C0, 0x6180, 0xA141, 0x6300, 0xA3C1, 0xA281, 0x6240,
	0x6600, 0xA6C1, 0xA781, 0x6740, 0xA501, 0x65C0, 0x6480, 0xA441,
	0x6C00, 0xACC1, 0xAD81, 0x6D40, 0xAF01, 0x6FC0, 0x6E80, 0xAE41,
	0xAA01, 0x6AC0, 0x6B80, 0xAB41, 0x6900, 0xA9C1, 0xA881, 0x6840,
	0x7800, 0xB8C1, 0xB981, 0x7940, 0xBB01, 0x7BC0, 0x7A80, 0xBA41,
	0xBE01, 0x7EC0, 0x7F80, 0xBF41, 0x7D00, 0xBDC1, 0xBC81, 0x7C40,
	0xB401, 0x74C0, 0x7580, 0xB541, 0x7700, 0xB7C1, 0xB681, 0x7640,
	0x7200, 0xB2C1, 0xB381, 0x7340, 0xB101, 0x71C0, 0x7080, 0xB041,
	0x5000, 0x90C1, 0x9181, 0x5140, 0x9301, 0x53C0, 0x5280, 0x9241,
	0x9601, 0x56C0, 0x5780, 0x9741, 0x5500, 0x95C1, 0x9481, 0x5440,
	0x9C01, 0x5CC0, 0x5D80, 0x9D41, 0x5F00, 0x9FC1, 0x9E81, 0x5E40,
	0x5A00, 0x9AC1, 0x9B81, 0x5B40, 0x9901, 0x59C0, 0x5880, 0x9841,
	0x8801, 0x48C0, 0x4980, 0x8941, 0x4B00, 0x8BC1, 0x8A81, 0x4A40,
	0x4E00, 0x8EC1, 0x8F81, 0x4F40, 0x8D01, 0x4DC0, 0x4C80, 0x8C41,
	0x4400, 0x84C1, 0x8581, 0x4540, 0x8701, 0x47C0, 0x4680, 0x8641,
	0x8201, 0x42C0, 0x4380, 0x8341, 0x4100, 0x81C1, 0x8081, 0x4040,
}

// Hextobin covnerts the raw string of data to a integer
func Hextobin(hexstring string) (convertedVar int) {
	res, _ := hex.DecodeString(hexstring)
	result := 0
	for x := range res {
		varLen := len(res) - 1
		arrayPos := varLen - x
		shift := 8 * uint8(x)
		value := int(res[arrayPos])
		result += value << shift
	}
	return result
}

// calculateCRC calculates the CRC of the recieved packet
func calculateCRC(pkt string) (CRC string) {
	var crc uint16
	var temp uint8

	for _, value := range pkt {
		temp = uint8(crc ^ uint16(value))
		crcShift := crc >> 8
		crcGet := crcTable[temp]
		crc = crcShift ^ crcGet
	}

	var h, l uint8 = uint8(crc >> 8), uint8(crc & 0xff)
	crcArray := make([]byte, 2)
	crcArray[0] = h
	crcArray[1] = l
	crcRes := hex.EncodeToString(crcArray)
	actCRC := strings.ToUpper(crcRes)
	return actCRC
}

// Packet - object containing the information rx/tx inside an ASL packet
type Packet struct {
	timestamp int64
	serial    string
	address   string
	cmd       string
	devType   string
	datacnt   string
	data      string
	crc       string
	raw       string
}

// NewTxPkt returns the pointer to a created txpacket
func NewTxPkt(address, serialNumber, cmd, payload string) *Packet {
	p := new(Packet)
	p.serial = serialNumber
	p.address = address
	p.cmd = cmd
	p.data = payload

	// Load header SOF + ADDRESS + COMMAND
	pkt := pktSOF + p.address + p.serial + p.cmd

	// Load datacount
	datacnt := len(p.data) / 2
	datacntStr := fmt.Sprintf("%04X", datacnt)
	dataCnt := datacntStr[2:4] + datacntStr[0:2]
	p.datacnt = dataCnt
	pkt += p.datacnt
	if datacnt != 0 {
		pkt += p.data
	}

	// Calculate CRC
	crc := calculateCRC(pkt)
	p.crc = crc

	// Load raw with full packet
	p.raw = pkt + crc + string(pktEOF)
	p.timestamp = time.Now().Unix()
	return p
}

// Bytes - returns the raw string in the packet as a byte array for transmission
func (p *Packet) Bytes() []byte {
	msg := preAmble + p.raw
	return []byte(msg)
}

// NewRxPkt returns the pointer to a created Packet built from the pkt string provided
func NewRxPkt(pkt string) (*Packet, error) {
	p := new(Packet)
	p.raw = pkt
	err := p.parse()
	if err != nil {
		return p, err
	}
	p.timestamp = time.Now().Unix()

	return p, nil
}

func (p *Packet) parse() error {
	raw := strings.TrimSpace(p.raw)
	raw = strings.TrimLeft(raw, "UU")
	raw = strings.TrimRight(raw, string(pktEOF))

	// Check incoming packet for the presence of the SOF character
	index := strings.LastIndex(raw, pktSOF)

	// If the packet doesn't contain a SOF the index will be set to -1
	if index == -1 {
		p.data = raw
		return ErrNoStartChar
	}

	raw = raw[index:]

	if len(raw) < pktDataPosStart {
		return ErrTooShort
	}

	if len(raw) < minimumPktSize {
		return ErrNoCRC
	}

	datacntLo := raw[pktDataCntLoPosStart:pktDataCntLoPosEnd]
	datacntHi := raw[pktDataCntHiPosStart:pktDataCntHiPosEnd]
	datacnt := datacntHi + datacntLo
	dataLength := Hextobin(datacnt)
	if len(raw) < minimumPktSize+dataLength {
		return ErrSizeMismatch
	}

	addr := raw[pktAddrPosStart:pktAddrPosEnd]
	if addr == masterAddress || addr == "\xc3" {
		// This is a packet from me
		return ErrMasterPacket
	}

	serial := raw[pktSerialPosStart:pktSerialPosEnd]

	if strings.Contains(serial, "!") {
		return ErrInvalidChar
	}

	cmd := raw[pktCmdPosStart:pktCmdPosEnd]
	data := raw[pktDataPosStart : len(raw)-4]
	rxCRC := raw[len(raw)-4:]

	rxCRCval := uint16(Hextobin(rxCRC))
	p.crc = strings.TrimSpace(rxCRC)

	calcCRC := calculateCRC(raw[:len(raw)-4])
	calcCRCval := uint16(Hextobin(calcCRC))

	if calcCRCval != rxCRCval {
		return fmt.Errorf("Packet CRC Failed - Rx: 0x%04X, Calc: 0x%04X", rxCRCval, calcCRCval)
	}

	if (cmd == readingCommand) && (dataLength == 0) {
		return ErrMasterPacket //fmt.Errorf("Packet is a reading request from a master don't process")
	}

	p.serial = serial
	p.address = addr
	p.cmd = cmd
	p.datacnt = datacnt
	p.data = data
	p.timestamp = time.Now().Unix()
	return nil
}
