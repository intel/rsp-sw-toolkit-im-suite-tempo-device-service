package main

import (
	"bytes"
	"encoding/hex"
	"github.impcloud.net/RSP-Inventory-Suite/expect"
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
	tcd := new(TempoDiskCurrent)
	w.ShouldSucceed(tcd.UnmarshalBinary(data[:n]))
	w.ShouldBeEqual(tcd.MAC, "C1:EE:03:79:EA:8C")
	w.ShouldBeEqual(tcd.Temperature, float32(22.4))
}

func TestDecodeTempoData(t *testing.T) {
	w := expect.WrapT(t)
	data := w.ShouldHaveResult(hex.DecodeString(
		"043E2B020100018CEA7903EEC11F020106" +
			"11FF33010D64003C32F5" +
			"010200000000010009094331454530333739B7")).([]byte)
	tcd := new(TempoDiskCurrent)
	w.ShouldSucceed(tcd.UnmarshalBinary(data))
	w.ShouldBeEqual(tcd.Temperature, float32(25.8))
}
