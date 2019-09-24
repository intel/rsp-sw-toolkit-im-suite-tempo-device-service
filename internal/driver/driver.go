/*
 * INTEL CONFIDENTIAL
 * Copyright (2019) Intel Corporation.
 *
 * The source code contained or described herein and all documents related to the source code ("Material")
 * are owned by Intel Corporation or its suppliers or licensors. Title to the Material remains with
 * Intel Corporation or its suppliers and licensors. The Material may contain trade secrets and proprietary
 * and confidential information of Intel Corporation and its suppliers and licensors, and is protected by
 * worldwide copyright and trade secret laws and treaty provisions. No part of the Material may be used,
 * copied, reproduced, modified, published, uploaded, posted, transmitted, distributed, or disclosed in
 * any way without Intel/'s prior express written permission.
 * No license under any patent, copyright, trade secret or other intellectual property right is granted
 * to or conferred upon you by disclosure or delivery of the Materials, either expressly, by implication,
 * inducement, estoppel or otherwise. Any license under such intellectual property rights must be express
 * and approved by Intel in writing.
 * Unless otherwise agreed by Intel in writing, you may not remove or alter this notice or any other
 * notice embedded in Materials by Intel or Intel's suppliers or licensors in any way.
 */

package driver

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	sdk "github.com/edgexfoundry/device-sdk-go"
	sdkModel "github.com/edgexfoundry/device-sdk-go/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	edgexModels "github.com/edgexfoundry/go-mod-core-contracts/models"
	"io"
	"log"
	"net/http"
	"time"
)

type Driver struct {
	Logger  logger.LoggingClient
	AsyncCh chan<- *sdkModel.AsyncValues
	done    chan interface{}
	server  *http.Server
}

// NewProtocolDriver returns the package-level driver instance.
func NewProtocolDriver() sdkModel.ProtocolDriver {
	return new(Driver)
}

// Initialize the driver.
func (driver *Driver) Initialize(lc logger.LoggingClient, asyncCh chan<- *sdkModel.AsyncValues) error {
	driver.Logger = lc
	driver.AsyncCh = asyncCh
	driver.done = make(chan interface{})

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(writer http.ResponseWriter, r *http.Request) {
		driver.Logger.Info("health check from %q", r.UserAgent())
	})
	mux.Handle("/hcidump", hciHandler{driver: driver})
	driver.server = &http.Server{Addr: ":49983", Handler: mux} // TODO: make port driver config

	go driver.runUntilCancelled()
	return nil
}

// runUntilCancelled will block forever until done is signaled to shutdown.
func (driver *Driver) runUntilCancelled() {
	driver.Logger.Info("Starting server.")
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
func (driver *Driver) HandleReadCommands(deviceName string, protocols map[string]models.ProtocolProperties, reqs []sdkModel.CommandRequest) ([]*sdkModel.CommandValue, error) {
	return nil, nil
}

// HandleWriteCommands ignores all requests.
func (driver *Driver) HandleWriteCommands(deviceName string, protocols map[string]models.ProtocolProperties, reqs []sdkModel.CommandRequest, params []*sdkModel.CommandValue) error {
	return nil
}

type hciHandler struct {
	driver *Driver
}

func (hh hciHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	defer Drain(request.Body)

	buff := bytes.Buffer{}
	if err := IgnoreEOF(io.CopyN(&buff, request.Body, 200)); err != nil {
		log.Printf("error: %+v\n", err)
		writer.WriteHeader(500)
	}

	decoder := hex.NewDecoder(NewSpaceSkipReader(&buff))
	data := make([]byte, 100) // TODO: benchmark using a buffer pool
	n, err := decoder.Read(data)
	if err != nil {
		log.Printf("error: %+v\n", err)
		writer.WriteHeader(500)
	}

	tcd := TempoDiskCurrent{}
	if err := tcd.UnmarshalBinary(data[:n]); err != nil {
		return
	}

	if _, notFound := sdk.RunningService().GetDeviceByName(tcd.MAC); notFound != nil {
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

func (driver *Driver) sendTemperature(tcd TempoDiskCurrent) error {
	origin := time.Now().UnixNano() / int64(time.Millisecond)
	value, err := sdkModel.NewFloat32Value("Temperature", origin, tcd.Temperature)
	if err != nil {
		return err
	}

	driver.AsyncCh <- &sdkModel.AsyncValues{
		DeviceName:    tcd.MAC,
		CommandValues: []*sdkModel.CommandValue{value},
	}
	driver.Logger.Info(fmt.Sprintf("Sent new reading: %+v", tcd))

	return nil
}

func (driver *Driver) registerTempoDisc(tcd TempoDiskCurrent) (err error) {
	_, err = sdk.RunningService().AddDevice(edgexModels.Device{
		Name:           tcd.MAC,
		AdminState:     edgexModels.Unlocked,
		OperatingState: edgexModels.Enabled,
		Profile:        edgexModels.DeviceProfile{Name: "Tempo-Disc"},
		Protocols: map[string]edgexModels.ProtocolProperties{
			"": {},
		},
	})
	driver.Logger.Info(fmt.Sprintf("Registered new tempo disc: %q", tcd.MAC))
	return
}
