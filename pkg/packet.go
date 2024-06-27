package sbtp

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"unsafe"
)

const PacketVersion = "SBTPv1"
const ByteNull uint8 = 0x00
const ByteEOT uint8 = 0x04
const ByteCountHeader = 16

/*
SBTPPacket is a container for sending and receiving a []byte payload over a [net.Conn]

	| Packet Version | NULL   | Content-Length | NULL   | Payload              | EOT    |
	| 6 bytes        | 1 byte | 8 bytes		   | 1 byte | Content-Length bytes | 1 byte |
*/
type SBTPPacket struct {
	sourceAddr net.Addr
	payload    []byte
}

// expectByte is a convenience method for asserting that the next byte from r is expectedByte
func expectByte(r io.Reader, expectedByte uint8) error {
	byteDest := make([]byte, 1)
	bytesRead, err := r.Read(byteDest)

	if err != nil {
		return err
	}

	if bytesRead != len(byteDest) {
		return fmt.Errorf("expected to read %d bytes, read %d", len(byteDest), bytesRead)
	}

	if byteDest[0] != expectedByte {
		return fmt.Errorf("expected to read 0x%02x, got 0x%02x", expectedByte, byteDest[0])
	}

	return nil
}

// NewSBTPPacket returns a ready-to-use SBTPPacket with an empty payload
func NewSBTPPacket(sourceAddr net.Addr) *SBTPPacket {
	return &SBTPPacket{
		sourceAddr: sourceAddr,
		payload:    make([]byte, 0),
	}
}

// GetPayload returns the current payload of the SBTPPacket
func (p *SBTPPacket) GetPayload() []byte {
	return p.payload
}

// SetPayload sets the payload of the SBTPPacket to payload
func (p *SBTPPacket) SetPayload(payload []byte) {
	p.payload = payload
}

// ReadFrom implements the [io.ReaderFrom] interface, reading an SBTPPacket from r and populating the payload. This
// makes it easy to read an SBTPPacket off of a [net.Conn].
func (p *SBTPPacket) ReadFrom(r io.Reader) (int64, error) {
	var contentLength uint64

	packetVersion := make([]byte, len(PacketVersion))

	bytesRead, err := r.Read(packetVersion)
	if err != nil {
		return int64(bytesRead), err
	}

	if bytesRead != len(packetVersion) {
		return int64(bytesRead), fmt.Errorf("expected %d bytes for packet version, got %d", len(packetVersion), bytesRead)
	}

	if string(packetVersion) != PacketVersion {
		return int64(bytesRead), errors.New("invalid packet version")
	}

	err = expectByte(r, ByteNull)
	if err != nil {
		return int64(bytesRead), errors.New("packet version missing null terminator")
	}
	bytesRead += 1

	err = binary.Read(r, binary.BigEndian, &contentLength)
	if err != nil {
		return int64(bytesRead), err
	}
	bytesRead += int(unsafe.Sizeof(contentLength))

	if contentLength > math.MaxUint64 || contentLength < 0 {
		return int64(bytesRead), fmt.Errorf("invalid packet length %d", contentLength)
	}

	err = expectByte(r, ByteNull)
	if err != nil {
		return int64(bytesRead), errors.New("content-length missing null terminator")
	}

	p.payload = make([]byte, contentLength)
	limitReader := io.LimitReader(r, int64(contentLength))
	payloadBytesRead, err := limitReader.Read(p.payload)
	if err != nil {
		return int64(bytesRead + payloadBytesRead), err
	}

	totalBytesRead := int64(bytesRead + payloadBytesRead)

	if uint64(payloadBytesRead) != contentLength {
		return totalBytesRead, fmt.Errorf("content-length mismatch: expected %d, got %d", contentLength, bytesRead)
	}

	err = expectByte(r, ByteEOT)
	if err != nil {
		return totalBytesRead, errors.New("payload missing eot terminator")
	}
	totalBytesRead += 1

	return totalBytesRead, nil
}

// WriteTo implements the [io.WriterTo] interface, writing the SBTPPacket to w. This makes it easy to send an SBTPPacket
// over a [net.Conn].
func (p *SBTPPacket) WriteTo(w io.Writer) (int64, error) {
	var bytesWritten int64

	payloadLenInt := len(p.payload)
	contentLength := uint64(payloadLenInt)

	if contentLength > math.MaxUint64 {
		return bytesWritten, fmt.Errorf("payload too large, max %x bytes", uint64(math.MaxUint64))
	}

	header := bytes.NewBuffer(make([]byte, 0, ByteCountHeader))

	header.Write([]byte(PacketVersion))
	header.WriteByte(ByteNull)

	err := binary.Write(header, binary.BigEndian, contentLength)
	if err != nil {
		return bytesWritten, err
	}

	header.WriteByte(ByteNull)

	bytesWrittenHeader, err := header.WriteTo(w)
	if err != nil {
		return bytesWrittenHeader, fmt.Errorf("error writing request header: %s", err)
	}
	bytesWritten = bytesWrittenHeader

	if bytesWritten != ByteCountHeader {
		return bytesWritten, fmt.Errorf("expected to write %d header bytes, wrote %d", ByteCountHeader, bytesWrittenHeader)
	}

	bytesWrittenPayload, err := w.Write(p.payload)
	bytesWritten += int64(bytesWrittenPayload)
	if err != nil {
		return bytesWritten, fmt.Errorf("error writing request payload: %s", err)
	}

	if uint64(bytesWrittenPayload) != contentLength {
		return bytesWritten, fmt.Errorf("expected to write %d payload bytes, wrote %d", contentLength, bytesWritten)
	}

	bytesWrittenEOT, err := w.Write([]byte{ByteEOT})
	bytesWritten += int64(bytesWrittenEOT)
	if err != nil {
		return bytesWritten, fmt.Errorf("error writing eot byte: %s", err)
	}

	if bytesWrittenEOT != 1 {
		return bytesWritten, fmt.Errorf("expected to write 1 eot byte, wrote %d", bytesWritten)
	}

	return bytesWritten, nil
}

// SourceAddr return the address where the SBTPPacket originated
func (p *SBTPPacket) SourceAddr() net.Addr {
	return p.sourceAddr
}

// ToString is a debug method that writes the SBTPPacket header, base64-encoded payload, and EOT byte footer
func (p *SBTPPacket) ToString() string {
	asString := fmt.Sprintf("| %s | 0x%02x | 0x%02x | 0x%02x |\n", PacketVersion, ByteNull, len(p.payload), ByteNull)

	payloadWriter := bytes.NewBufferString("")
	b64encoder := base64.NewEncoder(base64.StdEncoding, payloadWriter)
	_, err := b64encoder.Write(p.payload)
	if err != nil {
		return asString
	}
	payload := payloadWriter.String()
	payloadSplit := ""

	counter := 0
	for _, c := range payload {
		payloadSplit += string(c)
		if counter > 80 {
			counter = 0
			payloadSplit += "\n"
		}
		counter += 1
	}

	asString += payloadSplit + "\n"
	asString += fmt.Sprintf("| 0x%02x |", ByteEOT)

	return asString
}
