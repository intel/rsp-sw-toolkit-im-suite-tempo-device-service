package driver

import (
	"errors"
	"fmt"
)

// TempoDiskCurrent is data derived from the announcement of current data from a
// Blue Maestro Tempo Disk sensor.
type TempoDiskCurrent struct {
	MAC         string
	Temperature float32
}

// leMACToString converts a LE-MAC address to a pretty-printed string.
func leMACToString(macLE []byte) string {
	return fmt.Sprintf("%02X:%02X:%02X:%02X:%02X:%02X",
		macLE[5], macLE[4], macLE[3], macLE[2], macLE[1], macLE[0])
}

// TempoDecodeError is the error type returned during Unmarshalling if the input
// data cannot be decoded into TempoDiskCurrent.
type TempoDecodeError error

var (
	InvalidLength       = TempoDecodeError(errors.New("wrong data length"))
	InvalidPreamble     = TempoDecodeError(errors.New("wrong preamble"))
	InvalidPDUType      = TempoDecodeError(errors.New("wrong PDU type"))
	InvalidManufacturer = TempoDecodeError(errors.New("wrong manufacturer ID"))
)

// UnmarshalBinary decodes advertisement data from Tempo Disks.
func (tcd *TempoDiskCurrent) UnmarshalBinary(data []byte) error {
	if len(data) != 46 {
		return InvalidLength
	}
	if data[0] != 0x04 {
		return InvalidPreamble
	}
	if data[18] != 0xFF {
		return InvalidPDUType
	}
	if data[17] != 0x11 {
		return InvalidLength
	}
	if data[19] != 0x33 || data[20] != 01 {
		return InvalidManufacturer
	}

	tcd.MAC = leMACToString(data[7:16])
	tcd.Temperature = float32(int16(data[27])<<8|int16(data[28])) / 10.0
	return nil
}
