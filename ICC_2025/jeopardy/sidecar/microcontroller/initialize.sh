#!/bin/bash
set -euox pipefail

cd $(dirname $0)

./stop_can_interface.sh

./reset.sh
./write_firmware.sh
./start_can_interface.sh
