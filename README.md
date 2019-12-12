# Tempo Disk Device Service

This service hosts an HTTP endpoint which takes in `hcidump` data, finds Blue 
Maestro Tempo Disc advertisements, decodes them, and sends temperature readings
to EdgeX, registering new sensors as needed.

The current service is specific to Blue Maestro Tempo Disc sensors, but puts
forth some framework which could be extended to a generic BLE device service.

## Building
Build the service `make build`; it's written for Go 1.12 with module support.
The output is `cmd/tempo-device-service`. It can run locally, though you may 
need to modify the `GO` variable to match your OS/architecture if that's your
intent.

## Running in Docker
Assuming you're running Docker on `localhost`:
- Run EdgeX dependencies (`edgex-consul`, `core-data`, `core-metadata`, 
  `support-logging`). 
- Edit the [service's configuration](cmd/res/docker/configuration.toml) so that
  the `Host`s and `Port`s for `Registry` and EdgeX `Clients` will be reachable.
- Use `make` to build and run the service in a Docker container.
- Verify connectivity by `GET`ting `localhost:9001/` - should respond `200 OK`.
- POST hex data to `localhost:9001/hcidump`

## Options
The `configuration.toml` includes a `Driver` configuration section with the
single value `ListenAddress`. This address specifies the HTTP address that the 
service will listen for incoming `hcidump` data. In the above example, it's
set to `9001`.

## Usage
The service hosts two HTTP endpoints:
- `/` is simply a healthcheck endpoint. If the service is up and reachable, this
 should respond with `200 OK` and write a log message.
- `/hcidump` accepts hex encoded strings which it attempts to decode as Tempo
  Disc data; non-matching data is simply ignored.
  
The `/hcidump` endpoint expects data in the same output format as `hcidump -R`, 
but it doesn't know what the source of the actual data is, so you can easily 
test the service with `curl` or Postman or similar. The data should be a single 
line from `hcidump -R`, sans the `>` character. For example:

    curl -s -S -X POST \
      -H 'Content-Type: application/text' \
      -d  '04 3E 2B 02 01 00 01 8C EA 79 03 EE C1 1F 02 01 06 11 FF 33 '` 
         `'01 0D 64 00 3C 32 3D 00 E0 00 00 00 00 01 00 09 09 43 31 45 '`
         `'45 30 33 37 39 C5' \
     192.168.99.101:9001/hcidump

The endpoint ignore `'\n'`, `'\r'`, `'\t'`, and `' '`, and accepts both upper 
and lower case hex characters. This allows it to accept `hcidump` data directly,
but it'll also accept data without formatting:

    curl -s -S -X POST \
      -H 'Content-Type: application/text' \
      -d '043e2b020100018cea7903eec11f02010611ff33010d64003c323d00e000000000010009094331454530333739c5' \
     192.168.99.101:9001/hcidump

When all is working, you'll see logs with `msg`s like:

    Sent new reading: {MAC:C1:EE:03:79:EA:8C Name:C1EE0379 Temperature:22.4}
    
### Data Format
The data format matches advertisements from Blue Maestro Tempo Disc sensors.
In the current code version, messages are considered valid if this is true:
 - the message is 92 hex characters (ignoring `[\n\r\t ]`)
 - byte 0 is `0x04`
 - bytes 17-20 are `0x11FF3301`
 - converted temperature is in [-30, 70]

In this case, the name, MAC address, and temperature are decoded and sent as a
reading to EdgeX. The name used for the message defaults to the device's reported
name (bytes 37-44); however, if those bytes are outside of the ASCII range, the
service instead uses the first 4 bytes of the MAC, converted to upper case hex
(doing so matches the default names the disc uses). It additionally logs a 
warning with the device's reported and assigned name.

#### More Specific Data Format
In the table below, `Check` shows if the value must `Match` exactly, 
is `Extract`ed as is, or ignored `-`:

|Bytes|Check|Value|Meaning|
|:---:|:---:|---:|:---|
|0|Match|04|BLE preamble|
|1-4| - |3e 2b 02 01|Access address|
|5| - |00|BLE 2 bit ADV_IND, 2 bit RFU, 2 bit Tx/Rx|
|6| - |01|2 bit RFU, 6 bit total payload length|
|7-12|Extract; used if Name is invalid|7f 94 4c 33 a1 ce| LE MAC address|
|13| - |1f|Message length|
|14| - |02|Sub-payload is 2 bytes, including type|
|15| - |01|Sub-payload Type is "Flags"|
|16| - |06|5 bit flags: Sim LE,BR/EDR Host, Sim LE,BR/EDR Controller, No BR/EDR, LE gen, LE Lim|
|17|Match|11|Sub-payload is 17 bytes, including type|
|18|Match|ff|Sub-payload Type is "Manufacturer Specific Data"|
|19-20|Match|33 01|Blue Maestro Manufacturer's ID|
|21| - |0d|Tempo Disc version number|
|22| - |64|Battery percentage|
|23-24| - |00 3c|Current log interval in seconds|
|25-26| - |28 de|Stored log count|
|27-28|Extract|01 22|Current temperature, 10ths of a degree|
|29-30| - |00 00|Current humidity (if supported)|
|31-32| - |00 00|Current dew point (if supported)|
|33| - |01|Mode|
|34| - |00|Alarm breach count (if alarms set)|
|35| - |09|Sub-payload is 9 bytes, including type|
|36| - |09|Sub-payload Type is "Complete Local Name"|
|37-44| Extract; used if 8 ASCII bytes |43 45 41 31 33 33 34 43|ASCII name (CEA1334C)|
|45| - |c2|Checksum (CRC32)|

### Sending BLE data
The [`sendhci.sh` script](bin/sendhci.sh) reads the `hcidump` raw data and `curl`s
it to a host, presumably the one running the tempo-device-service. By default, 
it uses `hcitool lescan --duplicates` to starts scanning for BLE advertisements,
but on some systems, this gives `Input/Output Error`, in which case you can tell
the script to skip the scan command, and instead activate the scan with something
else like `bluetoothctl` (instructions for this are below).

If you have a computer with a working bluetooth adapter, you can use the script 
(note that you'll likely need to use `sudo`). It uses sensible defaults, but
supports a small handful of options:

    usage: ./bin/sendhci.sh [OPTIONS]
    OPTIONS:
         --host|-h ENDPOINT
             Specify where to send hcidump data.
             default: localhost:80/hcidump

         --device|-d HCI_DEVICE_NUM
             Specify the number of the hci device to use
             default: 0

         --verbose|-v
             Print commands and scan data while running.
             default: false

         --no-lescan
             Only do hcidump, skipping the lescan.
             this is useful if you're starting scanning elsewhere,
             e.g., via when using bluetoothctl:
                 > sudo bluetoothctl
                 > menu scan
                 > duplicate-data on
                 > back
                 > scan on
             default: false

         --whitelist|-w
             Pass the '--whitelist' option to lescan (if using).
             Doing so instructs the bluetooth device to only report
             data that matches a device on the bluetooth whitelist.
             default: false

          --help
             Show this message and exit.


Sending data in this way requires a working bluetooth stack and device. On linux, 
most or all of the software is likely already installed. If so, the script should
work without issue. If it doesn't, here are some tips you can use to troubleshoot.

Processing messages can go wrong in a few main steps:
- the sensors aren't powered/on/beaconing
- BLE messages aren't picked up by the receiver
- messages aren't making it over the network to the service
- the services doesn't recognize the message or can't extract the temperature 

#### Get Bluetooth messages from the device
If you don't have bluetooth tools installed, you can get them on ubuntu with:

    apt-get install -y bluetooth bluez bluez-hcidump rfkill
    
If it's not running, you can start bluetooth with:

    systemctl start bluetooth.service
    
You can at list your bluetooth devices like this:

    rfkill list bluetooth --output ID,SOFT,HARD 
    
If you don't see your bluetooth device, you'll need to do some Googling to debug.

Assuming the device is listed, it also needs to be free from soft/hard blocks.
A soft block is essentially a note to the system to disable the device. You can
remove a soft block with `rfkill unblock bluetooth`. The script automatically
tries to handle this case. A hard block is a physical switch that prevents the 
device from running. If you see `blocked` in the `HARD` column, you'll need to 
find and change the physical switch setting.

If the device is up and running, you can try to scan for BLE messages with:

    hcitool lescan --duplicates

If this gives a message like `Input/Output error`, it may be that its unable to
get access to the hci device, probably because some other service is making use
of it. While `hcitool` is a convenient way of scripting interactions with the
bluetooth device, for debugging you may find it easier to try interacting with 
it via the interactive `bluetoothctl` CLI. The script's help message outputs the
commands to do so, but they're repeated here with some comments:

```bash
# Enter the bluetoothctl CLI.
> sudo bluetoothctl

# Enter the "scan" submenu.
> menu scan

# Enable "duplicate" BLE messages, which are diabled by default.
# Without this, a device's beacon is only reported once until the discovery of 
# that device times out.
> duplicate-data on

# Move back up to the main menu.
> back

# Start scanning for BLE messages.
> scan on
```
 
This will list the address of all BLE devices advertising within range. If you
don't see any, then either scanning isn't working or there aren't any devices
nearby. You can verify that your tempo disc is beaconing by installing their app 
to your smartphone, which will show any nearby discovered devices and their names.
Refer to their documentation for more information.

If you don't see any scan results, with neither `hcitool lescan` nor `bluetoothctl`,
you can try resetting the device via `bluetoothctl`. Enter `bluetoothctl`, then 
try cycling (or just powering) the bluetooth device via `power on`/`power off` 
commands. 

#### Make sure messages can make it over the network
The `curl` command sends messages to the target host specified with `-h <hostaddr>`.
By default, it's set to `localhost:9001/hcidump` since `9001` is the default 
port for the `tempo-device-service`. If you're running it on a different host or
have changed the port number, you'll need to use the `-h` flag to adjust it. 

If it's running in a Docker container, you'll need to make sure the container's 
port is exposed (this is done by default in the `docker-compose.yml`). You can 
verify it via `docker service ls`. You should see the `tempo-device-service` is 
running (replicas are 1/1) and that the port is exposed (there should be a line 
indicating which port(s) from the container are forwarded to which on the host). 
If the service isn't running, check the logs and follow typical Docker debugging.
If the ports aren't exposed, fix it in the compose and redeploy.

Check network connectivity to the container via `curl http://<host>:<port>/` 
(i.e., send a GET request to the service). It should respond with `200 OK`. If
not, make sure the service is running on the host and port you expect, then make
sure general network connectivity is working between the hosts. Obviously, this
is simpler if both are running on the same host.

#### Make sure the messages are correct 
This service was written for a specific device, so the messages are only processed 
correctly if they fit the format explained above. Follow the instructions in the 
`Usage` section to send fake data to the service via `curl` or another tool of
choice. You should see the service logs update with information about processing
the message. If not, then either network connectivity is still a problem (see
above) or the message format is invalid.

If that's working as expected, but you're still not getting messages from the
tempo devices, then it's likely they're not sending messages or they're not the
expected format. Follow the instructions in the section above about Bluetooth to
verify the device is visible from your phone and then view the raw message dumps
by running the `sendhci.sh` script with the `-v` flag. If the device is visible,
you can find messages its sending by searching for those matching the format
described in an earlier section. Subpayload headers with `FF 33 01` represent
Blue Maestro manufacturer data, and at least some of the messages should have
a message matching the rest of that format (the devices do send some other data,
but the format is different). If the message format is different from that, then
either your device doesn't match or the manufacturer has unexpectedly changed 
their API without warning.

