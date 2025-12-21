#!/bin/bash
set -euox pipefail

cd $(dirname $0)

stty -F /dev/serial0 raw parenb -parodd -cstopb cs8 115200 time 5 min 0 \
     line 0 -brkint -icrnl -imaxbel -opost -onlcr -isig -icanon -iexten \
     -echo -echok -echoctl -echoke

pkill -9 stm32flash || true
stm32flash -b 0 -w firmware.bin -v -g 0 /dev/serial0
