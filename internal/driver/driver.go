/* Apache v2 license
*  Copyright (C) <2019> Intel Corporation
*
*  SPDX-License-Identifier: Apache-2.0
 */

package driver

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"github.com/edgexfoundry/device-sdk-go"
	deviceModels "github.com/edgexfoundry/device-sdk-go/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	coreModels "github.com/edgexfoundry/go-mod-core-contracts/models"
	"io"
	"net/http"
	"strings"
	"time"
)

type Driver struct {
	Logger  logger.LoggingClient
	AsyncCh chan<- *deviceModels.AsyncValues
	done    chan interface{}
	server  *http.Server
}

// NewProtocolDriver returns the package-level driver instance.
func NewProtocolDriver() deviceModels.ProtocolDriver {
	return new(Driver)
}

const (
	ConfigListenAddr = "ListenAddress"
)

// Initialize the driver.
func (driver *Driver) Initialize(lc logger.LoggingClient, asyncCh chan<- *deviceModels.AsyncValues) error {
	driver.Logger = lc
	driver.AsyncCh = asyncCh
	driver.done = make(chan interface{})

	var listenAddr string
	if err := GetDriverConfig().Get(ConfigListenAddr, &listenAddr); err != nil {
		return err
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(writer http.ResponseWriter, r *http.Request) {
		driver.Logger.Info(fmt.Sprintf("health check from %q", r.UserAgent()))
	})
	mux.Handle("/hcidump", hciHandler{driver: driver})
	driver.server = &http.Server{Addr: ":" + listenAddr, Handler: mux}

	go driver.runUntilCancelled()
	return nil
}

// runUntilCancelled will block forever until done is signaled to shutdown.
func (driver *Driver) runUntilCancelled() {
	driver.Logger.Info(fmt.Sprintf("Starting server on port %s.", driver.server.Addr))
	go func() {
		driver.Logger.Info("Server stopped: %v", driver.server.ListenAndServe())
	}()

	<-driver.done
	driver.Logger.Info("Stopping server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := driver.server.Shutdown(ctx); err != nil {
		driver.Logger.Error("Server shutdown failed: %v", err)
	}
}

// Stop instructs the protocol-specific DS code to shutdown gracefully, or
// if the force parameter is 'true', immediately. The driver is responsible
// for closing any in-use channels, including the channel used to send async
// readings (if supported).
func (driver *Driver) Stop(force bool) error {
	close(driver.done)
	close(driver.AsyncCh)
	return nil
}

// HandleReadCommands ignore all requests.
func (driver *Driver) HandleReadCommands(_ string, _ map[string]coreModels.ProtocolProperties,
	_ []deviceModels.CommandRequest) ([]*deviceModels.CommandValue, error) {
	return nil, nil
}

// HandleWriteCommands ignores all requests.
func (driver *Driver) HandleWriteCommands(_ string, _ map[string]coreModels.ProtocolProperties,
	_ []deviceModels.CommandRequest, params []*deviceModels.CommandValue) error {
	return nil
}

type hciHandler struct {
	driver *Driver
}

func (hh hciHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	defer Drain(request.Body)

	buff := bytes.Buffer{}
	if err := IgnoreEOF(io.CopyN(&buff, request.Body, 200)); err != nil {
		hh.driver.Logger.Error(fmt.Sprintf("unknown error reading request body: %+v", err))
		writer.WriteHeader(500)
	}

	decoder := hex.NewDecoder(NewSpaceSkipReader(&buff))
	data := make([]byte, 100)
	n, err := decoder.Read(data)
	if err != nil {
		// not exactly an "error", but we can't process it.
		hh.driver.Logger.Info(fmt.Sprintf("data contains non-hex data: %v", err))
		writer.WriteHeader(400)
	}

	tcd := TempoDiscCurrent{}
	if err := tcd.UnmarshalBinary(data[:n]); err != nil {
		return
	}

	if _, notFound := device.RunningService().GetDeviceByName(tcd.Name); notFound != nil {
		if err := hh.driver.registerTempoDisc(tcd); err != nil {
			hh.driver.Logger.Error(fmt.Sprintf("Failed to register %q: %+v", tcd.MAC, err))
			return
		}
	}

	if err := hh.driver.sendTemperature(tcd); err != nil {
		hh.driver.Logger.Error(fmt.Sprintf("Failed to create new Temperature: %+v", err))
		return
	}
}

func (driver *Driver) checkName(tcd *TempoDiscCurrent) {
	if len(tcd.Name) != 8 || !isASCIIPrintable(tcd.Name) {
		oldName := []byte(tcd.Name)
		tcd.Name = strings.ToUpper(hex.EncodeToString(tcd.MAC[:4]))
		driver.Logger.Warn(fmt.Sprintf(
			"Device name isn't 8 bytes of ASCII printable characters (bytes: %#02X);" +
				"defaulting to using the upper-case hex of its first six MAC bytes: %s",
			oldName, tcd.Name))
	}
}

func (driver *Driver) sendTemperature(tcd TempoDiscCurrent) error {
	origin := time.Now().UnixNano() / int64(time.Millisecond)
	value, err := deviceModels.NewFloat32Value("Temperature", origin, tcd.Temperature)
	if err != nil {
		return err
	}

	driver.AsyncCh <- &deviceModels.AsyncValues{
		DeviceName:    tcd.Name,
		CommandValues: []*deviceModels.CommandValue{value},
	}
	driver.Logger.Info(fmt.Sprintf("Sent new reading: %+v", tcd))

	return nil
}

func (driver *Driver) registerTempoDisc(tcd TempoDiscCurrent) (err error) {
	_, err = device.RunningService().AddDevice(coreModels.Device{
		Name:           tcd.Name,
		AdminState:     coreModels.Unlocked,
		OperatingState: coreModels.Enabled,
		Profile:        coreModels.DeviceProfile{Name: "Tempo-Disc"},
		Protocols: map[string]coreModels.ProtocolProperties{
			"": {},
		},
	})
	driver.Logger.Info(fmt.Sprintf("Registered new tempo disc: %q", tcd.MAC))
	return
}
