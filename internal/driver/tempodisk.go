/* Apache v2 license
*  Copyright (C) <2019> Intel Corporation
*
*  SPDX-License-Identifier: Apache-2.0
 */

package driver

import (
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
)

const advertSize = 46

// TempoDiscCurrent is data derived from the announcement of current data from a
// Blue Maestro Tempo Disk sensor.
type TempoDiscCurrent struct {
	MAC         BLAddr
	Name        string
	Temperature float32
}

type BLAddr [6]byte

func (m BLAddr) String() string {
	return fmt.Sprintf("%02X:%02X:%02X:%02X:%02X:%02X",
		m[0], m[1], m[2], m[3], m[4], m[5])
}

// TempoDecodeError is the error type returned during unmarshalling if the input
// data cannot be decoded into TempoDiscCurrent.
type TempoDecodeError error

var (
	InvalidLength       = TempoDecodeError(errors.New("not enough data"))
	InvalidPreamble     = TempoDecodeError(errors.New("wrong preamble"))
	InvalidPDUType      = TempoDecodeError(errors.New("wrong PDU type"))
	InvalidManufacturer = TempoDecodeError(errors.New("wrong manufacturer ID"))
	InvalidTemperature  = TempoDecodeError(errors.New("temperature exceeds functional range"))

	nonStd = regexp.MustCompile(`^04.*?0001(?P<mac>.{12}).*?(?P<payload>11ff3301.{28}).*?(?P<name>0909.{16})`)
)

// parseNonStandard uses a regex to guess which portions of the message match the
// expected format, then reconstructs it to match that guess and parses the data
// that way, if it can.
func parseNonStandard(data []byte) (TempoDiscCurrent, error) {
	// inefficient, but effective
	asStr := hex.EncodeToString(data)
	result := nonStd.FindStringSubmatchIndex(asStr)
	if len(result) < 8 || result[2] == -1 || result[4] == -1 || result[6] == -1 {
		return TempoDiscCurrent{}, errors.New("unable to match data")
	}

	newData := make([]byte, advertSize)
	newData[0] = 0x04
	copy(newData[7:13], data[result[2]/2:result[3]/2])  // mac
	copy(newData[17:35], data[result[4]/2:result[5]/2]) // payload
	copy(newData[35:45], data[result[6]/2:result[7]/2]) // name

	tcd := new(TempoDiscCurrent)
	return *tcd, tcd.UnmarshalBinary(newData)
}

// UnmarshalBinary decodes advertisement data from Tempo Disks.
func (tcd *TempoDiscCurrent) UnmarshalBinary(data []byte) error {
	if len(data) < advertSize {
		return InvalidLength
	}
	if data[0] != 0x04 {
		return InvalidPreamble
	}
	if data[17] != 0x11 {
		return InvalidLength
	}
	if data[18] != 0xFF {
		return InvalidPDUType
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
