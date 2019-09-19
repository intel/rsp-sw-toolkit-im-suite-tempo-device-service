# Tempo Disk Device Service

This service hosts an HTTP endpoint to which Blue Maestro 
Tempo Disk advertisement data can be sent. It's decoded and
forwarded to EdgeX, registering new sensors as needed.

## How To Use
Build the service `make build`; you'll need a Go compiler.

You can run the service locally or in a Docker container;
`make image` will build a Docker image and `make run` will
run that image, connected to the local network, listening
on port 9001. Using just `make` will build the service and
image, then run it on 9001.

By default, the service serves on port 80, but it accepts 
`-p <port>` to run on a different port. 

The service lists for messages at `/hcidump`. Forward HCI 
dump data to this endpoint and the service will process it. 

Use the `sendhci.sh` script to send data from a local Bluetooth
device. By default, it's configured to use `hci0` and forward
the data to `localhost:9001`. 

## TODO
Currently the service receives and processes HCI Dump messages,
but doesn't actually do anything but log them. The actual EdgeX
steps come next, and will involve sending the data and registering
new sensors.
