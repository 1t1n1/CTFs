#!/bin/bash
set -eux
trap 'echo "### Connectivity check failed! ###"' ERR

if [ "${1:-}" != "--no-init" ]; then
    echo 'Stopping CAN interface......'
    /opt/microcontroller/stop_can_interface.sh

    echo 'Initializing CAN bus connection......'
    /opt/microcontroller/initialize.sh
fi

echo 'Testing CAN bus with echo service (positive case)......'
PAYLOAD=00$(dd if=/dev/urandom bs=7 count=1 status=none | xxd -p)
CAN_REQ=200#${PAYLOAD}
CAN_RSP=208#${PAYLOAD}
for i in `seq 0 4`
do
    cansend can0 ${CAN_REQ}
    sleep 0.0$[ $((RANDOM % 100)) ]
done
timeout --preserve-status 5 candump -L can0 | grep -i --max-count=1 ${CAN_RSP}

echo 'Testing CAN bus with echo service (negative case)......'
PAYLOAD=ff$(dd if=/dev/urandom bs=7 count=1 status=none | xxd -p)
CAN_REQ=200#${PAYLOAD}
CAN_RSP=208#${PAYLOAD}
for i in `seq 0 4`
do
    cansend can0 ${CAN_REQ}
    sleep 0.0$[ $((RANDOM % 100)) ]
done
timeout --preserve-status 5 candump -L can0 | grep -i --max-count=1 ${CAN_RSP} && false || true

echo '### Connectivity check succeeded. ###'
