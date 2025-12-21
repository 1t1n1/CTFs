#!/bin/bash
set -euox pipefail

modprobe can-raw
modprobe can
modprobe slcan
slcand -s5 -S115200 /dev/serial0 can0
ip link set can0 up
