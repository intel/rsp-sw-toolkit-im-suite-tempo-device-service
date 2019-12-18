/* Apache v2 license
*  Copyright (C) <2019> Intel Corporation
*
*  SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"github.com/edgexfoundry/device-sdk-go/pkg/startup"
	"github.com/intel/rsp-sw-toolkit-im-suite-tempo-device-service/internal/driver"
)

const (
	serviceName string = "tempo-device-service"
)

// Version is replaced by -ldflags "-X <module-name>/main.Version=1.0.0"
var Version = "1.0.0"

func main() {
	startup.Bootstrap(serviceName, Version, driver.NewProtocolDriver())
}
