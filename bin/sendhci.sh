#!/bin/bash

name=$0
function usage() {
        echo "usage: $name [OPTIONS]"
        echo "OPTIONS:"
        echo "     --host|-h ENDPOINT"
        echo "         where to send hcidump data"
        echo "         default: localhost:80/hcidump"
        echo ""
        echo "     --device|-d HCI_DEVICE_NUM"
        echo "         number of the hci device to use"
        echo "         default: 0"
        echo ""
        echo "     --whitelist|-w"
        echo "         pass --whitelist option to lescan"
        echo "         default: false"
        echo ""
        echo "     --verbose|-v"
        echo "         print scan data while running"
        echo "         default: false"
}

function ensure_arg() {
        if [[ $# -ne 2 ]]; then
                echo "missing arg for $1"
                usage; exit 1
        fi
}

hostaddr=localhost:9001/hcidump
device=0

# read options
while [[ "$1" =~ ^- ]]; do case $1 in
        --host | -h)
                ensure_arg $1 $2; shift; hostaddr=$1
                ;;
        --device | -d)
                ensure_arg $1 $2; shift; device=$1
                ;;
        --whitelist | -w)
                whitelist=1
                ;;
        --verbose | -v)
                verbose=1
                ;;
        *)
                usage; exit 1
                ;;
esac; shift; done

# consider it an error if there were more options supplied
if [[ $# -ne 0 ]]; then
        echo 'too many args'
        usage; exit 1
fi

# make sure the device is an unblocked Bluetooth device
function check_device() {
    if [[ "$1" != "$device" ]]; then
            return
    fi

    if [[ "$2" != "bluetooth" ]]; then
            echo "device $1 is $2, not bluetooth"
            exit 1;
    fi

    # check softblocks
    if [[ "$3" != "unblocked" ]]; then
            echo "removing softblock from $1..."
            (set -x && rfkill unblock bluetooth)
            if [[ $(rfkill list $device --noheading --output SOFT) != "unblocked" ]]; then
                    echo "device $1 is '$3', but 'rfkill unblock bluetooth' failed"
                    exit 1;
            fi
    fi

    # check hardblocks
    if [[ "$4" != "unblocked" ]]; then
            echo "device $1 is $4"
            exit 1;
    fi
}

dstatus=$(rfkill list $device --noheading --output TYPE,SOFT,HARD)
if [[ $? -ne 0 ]]; then
    echo "cannot find device $device";
    exit 1;
fi
check_device $dstatus
dname=$(rfkill list $device --noheading --output DEVICE)

scan_opts='--duplicates'
if [[ $whitelist ]]; then
    scan_opts="$scan_opts --whitelist"
fi

# start scanning
if [[ $verbose ]]; then
        xargs_opt='-t'
        (set -x && hcitool -i $dname lescan $scan_opts) &
else
        xargs_opt=''
        hcitool -i $dname lescan $scan_opts 1> /dev/null &
fi

# grab HCI data and send it to the processor
(hcidump -i $dname -R | \
    xargs $xargs_opt -d '>' -I{} \
        curl \
        -s -S \
        -H "Content-Type: application/text" \
        -X POST \
        -d '{}' \
        ${hostaddr:-localhost:80/hcidump}) &

# trap ctrl+c; kill the processes
trap "trap - SIGTERM && kill -- -$$ 2> /dev/null" SIGINT SIGTERM EXIT
echo "started $(jobs -pr | tr '\n' ' '); use SIGTERM (usually ctrl+c) to stop"
echo "sending data from $dname to ${hostaddr:-localhost:80/hcidump}"
wait
echo "stopped"
