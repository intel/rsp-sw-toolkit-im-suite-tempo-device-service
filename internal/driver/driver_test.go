package driver

import (
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	expect "github.com/intel/rsp-sw-toolkit-im-suite-expect"
	"testing"
)

func TestDecodeTempoData_Name(t *testing.T) {
	w := expect.WrapT(t)
	driver := Driver{Logger:logger.NewClient("", false, "", "DEBUG")}

	data := tempoData(w)
	tcd := new(TempoDiscCurrent)
	w.ShouldSucceed(tcd.UnmarshalBinary(data))
	driver.checkName(tcd)
	w.ShouldBeEqual(tcd.MAC, [6]byte{0xC1, 0xEE, 0x03, 0x79, 0xEA, 0x8C})
	w.ShouldBeEqual(tcd.Name, "C1EE0379")
	w.ShouldBeEqual(tcd.Temperature, float32(25.8))

	td := TempoDiscCurrent{
		MAC: [6]byte{0xDE, 0xAD, 0xBE, 0xEF, 0xDC, 0xBA},
		Name: "_hello!_",
		Temperature: 25,
	}
	data = w.ShouldHaveResult(td.MarshalBinary()).([]byte)
	w.ShouldSucceed(tcd.UnmarshalBinary(data))
	driver.checkName(tcd)
	w.ShouldBeEqual(tcd.MAC, [6]byte{0xDE, 0xAD, 0xBE, 0xEF, 0xDC, 0xBA})
	w.ShouldBeEqual(tcd.Name, "_hello!_")
	w.ShouldBeEqual(tcd.Temperature, float32(25))

	td = TempoDiscCurrent{
		MAC: [6]byte{0xDE, 0xAD, 0xBE, 0xEF, 0xDC, 0xBA},
		Name: "C1\000\0000379",
		Temperature: 25,
	}
	data = w.ShouldHaveResult(td.MarshalBinary()).([]byte)
	w.ShouldSucceed(tcd.UnmarshalBinary(data))
	driver.checkName(tcd)
	w.ShouldBeEqual(tcd.MAC, [6]byte{0xDE, 0xAD, 0xBE, 0xEF, 0xDC, 0xBA})
	w.ShouldBeEqual(tcd.Name, "DEADBEEF")
	w.ShouldBeEqual(tcd.Temperature, float32(25))
}
