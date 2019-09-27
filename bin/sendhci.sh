#!/bin/bash

name=$0
function usage() {
        echo "usage: $name [OPTIONS]"
        echo ""
        echo "OPTIONS:"
        echo "     --host|-h ENDPOINT"
        echo "         where to send hcidump data"
        echo "         default: localhost:9001/hcidump"
        echo ""
        echo "     --device|-d HCI_DEVICE_NUM"
        echo "         number of the hci device to use"
        echo "         default: 0"
        echo ""
        echo "     --verbose|-v"
        echo "         print scan data while running"
        echo "         default: false"
        echo ""
        echo "EXAMPLE:"
        echo "  $name --host 192.168.99.101:9001/hcidump"
}

# read args
while [[ "$1" =~ ^- ]]; do case $1 in
        --host | -h)
                shift; hostaddr=$1
                ;;
        --device | -d)
                shift; device=$1
                ;;
        --verbose | -v)
                verbose=1
                ;;
        *)
                usage
                exit 1
                ;;
esac; shift; done

# report extra args
if [[ $@ ]]; then
        usage
        exit 1
fi

# start BLE scanning
if [[ $verbose ]]; then
        hcitool lescan --duplicates &
else
        hcitool lescan --duplicates 1> /dev/null &
fi

# send hcidump data to the data processor:
#   hcidump (get the data):
#     -i specifies device ID
#     -R dumps "raw" hex data
#   xargs (create a command for each piece of data):
#     -d splits the hcidump data on '>' (it still has newlines)
#     -I{} replaces '{}' with the hcidump data in the curl command
#   curl (send the data upstream):
#     -s -S is silent except for errors
#     -H sets a header - we're sending text
#     -X specifies the HTTP verb (POST)
#     -d gives the data, which is replaced with the hex dump
#     the final arg is the host address
(hcidump -i ${device:-0} -R | \
    xargs -d '>' -I{} \
        curl \
        -s -S \
        -H "Content-Type: application/text" \
        -X POST \
        -d '{}' \
        ${hostaddr:-localhost:9001/hcidump}) &

# wait for signal/shutdown
trap "trap - SIGTERM && kill -- -$$ 2> /dev/null" SIGINT SIGTERM EXIT
echo "started $(jobs -pr); use ctrl+c to stop"
echo "sending data from hci${device:-0} to ${hostaddr:-localhost:80/hcidump}"
wait
echo "stopped"
