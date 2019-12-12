/* Apache v2 license
*  Copyright (C) <2019> Intel Corporation
*
*  SPDX-License-Identifier: Apache-2.0
 */

package driver

import (
	"errors"
)

const advertSize = 46

// TempoDiscCurrent is data derived from the announcement of current data from a
// Blue Maestro Tempo Disk sensor.
type TempoDiscCurrent struct {
	MAC         [6]byte
	Name        string
	Temperature float32
}

// TempoDecodeError is the error type returned during unmarshalling if the input
// data cannot be decoded into TempoDiscCurrent.
type TempoDecodeError error

var (
	InvalidLength       = TempoDecodeError(errors.New("wrong data length"))
	InvalidPreamble     = TempoDecodeError(errors.New("wrong preamble"))
	InvalidPDUType      = TempoDecodeError(errors.New("wrong PDU type"))
	InvalidManufacturer = TempoDecodeError(errors.New("wrong manufacturer ID"))
	InvalidTemperature  = TempoDecodeError(errors.New("temperature exceeds functional range"))
)

// UnmarshalBinary decodes advertisement data from Tempo Disks.
func (tcd *TempoDiscCurrent) UnmarshalBinary(data []byte) error {
	if len(data) != advertSize {
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

	for i, b := range data[7:13] {
		tcd.MAC[5-i] = b
	}
	tcd.Name = string(data[37:45])
	tcd.Temperature = float32(int16(data[27])<<8|int16(data[28])) / 10.0

	if tcd.Temperature < -30.0 || tcd.Temperature > 75.0 {
		return InvalidTemperature
	}
	return nil
}

// isASCIIPrintable returns true if the string contains any characters outside
// the range 0x21-0x7E, i.e. the characters in the following regex class:
// [a-zA-Z0-9!-#$%&'()*+,-./:;<=>?@[\]^_`{\}~]. Note that this excludes 0x20,
// the ASCII space character.
func isASCIIPrintable(s string) bool {
	for i := range s {
		if !(s[i] >= '!' && s[i] <= '~') {
			return false
		}
	}
	return true
}

