#!/bin/bash

HOSTADDR="localhost:9001"
DEVICE="0"

hcitool lescan --duplicates 1> /dev/null &
(hcidump -i $DEVICE -R | \
    xargs -d '>' -I{} \
        curl \
        -s -S \
        -H "Content-Type: application/text" \
        -X POST \
        -d '{}' \
        $HOSTADDR) &

trap "trap - SIGTERM && kill -- -$$ 2> /dev/null" SIGINT SIGTERM EXIT
echo "started $(jobs -pr)"
echo "sending data from hci$DEVICE to $HOSTADDR"
wait
echo "stopped"
