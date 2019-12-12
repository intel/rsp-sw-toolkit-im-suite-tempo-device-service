/* Apache v2 license
*  Copyright (C) <2019> Intel Corporation
*
*  SPDX-License-Identifier: Apache-2.0
 */

package driver

import (
	"bytes"
	"encoding/hex"
	"github.com/intel/rsp-sw-toolkit-im-suite-expect"
	"testing"
)

func TestHCIDumpToTempoData(t *testing.T) {
	w := expect.WrapT(t)

	input := []byte("04 3E 2B 02 01 00 01 8C EA 79 03 EE C1 1F 02 01 06 11 FF 33" +
		"01 0D 64 00 3C 32 3D 00 E0 00 00 00 00 01 00 09 09 43 31 45" +
		"45 30 33 37 39 C5")
	decoder := hex.NewDecoder(NewSpaceSkipReader(bytes.NewReader(input)))
	data := make([]byte, 100)
	n := w.ShouldHaveResult(decoder.Read(data)).(int)
	tcd := new(TempoDiscCurrent)
	w.ShouldSucceed(tcd.UnmarshalBinary(data[:n]))
	w.ShouldBeEqual(tcd.MAC, [6]byte{0xC1, 0xEE, 0x03, 0x79, 0xEA, 0x8C})
	w.ShouldBeEqual(tcd.Name, "C1EE0379")
	w.ShouldBeEqual(tcd.Temperature, float32(22.4))
}

func TestDecodeTempoData(t *testing.T) {
	w := expect.WrapT(t)
	tcd := new(TempoDiscCurrent)
	w.ShouldSucceed(tcd.UnmarshalBinary(tempoData(w)))
	w.ShouldBeEqual(tcd.MAC, [6]byte{0xC1, 0xEE, 0x03, 0x79, 0xEA, 0x8C})
	w.ShouldBeEqual(tcd.Name, "C1EE0379")
	w.ShouldBeEqual(tcd.Temperature, float32(25.8))
}

func TestDecodeTempoData_invalidTemperature(t *testing.T) {
	w := expect.WrapT(t)
	td := TempoDiscCurrent{
		MAC: [6]byte{0xC1, 0xEE, 0x03, 0x79, 0xEA, 0x8C},
		Name: "C1EE0379",
		Temperature: 100,
	}
	data := w.ShouldHaveResult(td.MarshalBinary()).([]byte)
	tcd := new(TempoDiscCurrent)
	w.ShouldFail(tcd.UnmarshalBinary(data))

	td = TempoDiscCurrent{
		MAC: [6]byte{0xC1, 0xEE, 0x03, 0x79, 0xEA, 0x8C},
		Name: "C1EE0379",
		Temperature: -31,
	}
	data = w.ShouldHaveResult(td.MarshalBinary()).([]byte)
	tcd = new(TempoDiscCurrent)
	w.ShouldFail(tcd.UnmarshalBinary(data))
}

func (td *TempoDiscCurrent) MarshalBinary() ([]byte, error) {
	leMAC := []byte{td.MAC[5], td.MAC[4], td.MAC[3], td.MAC[2], td.MAC[1], td.MAC[0]}
	tb := int16(td.Temperature * 10)
	if len(td.Name) < 8 {
		td.Name = td.Name + "\000\000\000\000\000\000\000\000"[len(td.Name):]
	}
	return hex.DecodeString(
		"04" + "3E2B0201" + "00" + "01" + // BLE control data
			hex.EncodeToString(leMAC) + // little-endian MAC address
			"1F" + "02" + "01" + "06" + "11" + // payload metadata
			"FF3301" + // manufacturer data flag/ID
			"0D" + "64" + "003C" + "32F5" + // version, battery %, log interval, log count
			hex.EncodeToString([]byte{uint8(tb >> 8), uint8(tb)}) + // temperature, in tenths of a degree
			"0000" + "0000" + "01" + "00" + // humidity, dew point, mode, alarm breaches
			"09" + "09" + // payload metadata
			hex.EncodeToString([]byte(td.Name[:8])) + // ASCII name
			"B7")
}

func tempoData(w *expect.TWrapper) []byte {
	td := TempoDiscCurrent{
		MAC: [6]byte{0xC1, 0xEE, 0x03, 0x79, 0xEA, 0x8C},
		Name: "C1EE0379",
		Temperature: 25.8,
	}

	return w.ShouldHaveResult(td.MarshalBinary()).([]byte)
}
