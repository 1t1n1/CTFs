#!/bin/bash
set -euox pipefail

pinctrl set 18 op pd
sleep 1
pinctrl set 18 ip pn
sleep 1
