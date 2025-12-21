#!/bin/bash
set -euox pipefail

ip link set can0 down || true
pkill -9 slcand || true
