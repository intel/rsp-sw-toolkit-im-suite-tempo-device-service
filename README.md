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

### Example Output
Here's some sample output. The logs are the device's MAC, its 
reported temperature, and the raw advertisement data:
```bash
> docker run --rm -it --net host tempo-device-service -p 9001
2019/09/19 19:24:22 Serving on HTTP port: 9001
2019/09/19 19:24:22 {MAC:CE:A1:33:4C:94:7F Temperature:24.5} 043E2B020100017F944C33A1CE1F02010611FF33010D64003C330600F500000000010009094345413133333443C5
2019/09/19 19:24:28 {MAC:C1:EE:03:79:EA:8C Temperature:24.9} 043E2B020100018CEA7903EEC11F02010611FF33010D64003C330400F900000000010009094331454530333739B8
2019/09/19 19:24:32 {MAC:C1:EE:03:79:EA:8C Temperature:24.9} 043E2B020100018CEA7903EEC11F02010611FF33010D64003C330400F900000000010009094331454530333739BD
2019/09/19 19:24:38 {MAC:CE:A1:33:4C:94:7F Temperature:24.5} 043E2B020100017F944C33A1CE1F02010611FF33010D64003C330700F500000000010009094345413133333443C5
2019/09/19 19:24:40 {MAC:CE:A1:33:4C:94:7F Temperature:24.5} 043E2B020100017F944C33A1CE1F02010611FF33010D64003C330700F500000000010009094345413133333443BF
2019/09/19 19:24:46 {MAC:C1:EE:03:79:EA:8C Temperature:24.9} 043E2B020100018CEA7903EEC11F02010611FF33010D64003C330400F900000000010009094331454530333739BA
2019/09/19 19:24:46 {MAC:CE:A1:33:4C:94:7F Temperature:24.5} 043E2B020100017F944C33A1CE1F02010611FF33010D64003C330700F500000000010009094345413133333443C5
2019/09/19 19:24:48 {MAC:C1:EE:03:79:EA:8C Temperature:24.9} 043E2B020100018CEA7903EEC11F02010611FF33010D64003C330400F900000000010009094331454530333739B7
2019/09/19 19:24:50 {MAC:C1:EE:03:79:EA:8C Temperature:24.9} 043E2B020100018CEA7903EEC11F02010611FF33010D64003C330400F900000000010009094331454530333739C5
```

## TODO
Currently the service receives and processes HCI Dump messages,
but doesn't actually do anything but log them. The actual EdgeX
steps come next, and will involve sending the data and registering
new sensors.

It may be worth decoding some of the other advertisement data. 
They broadcast their battery level, log info, and some aggregate
stats. If the disk supports it, it also sends out humidity and
dew point (if it's not supported, its still sent, but always as 0).

There are some commands they support for changing logging/reporting
intervals and temperature units. At the very least, it may be worth 
checking the settings and making sure that its reporting in Celsius
and taking readings at a sufficiently high rate, but the steps to
get to that point involve connecting to a BLE UART service, sending 
commands, and reading responses -- much more involved than just 
sniffing the advertisements.

