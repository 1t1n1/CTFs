#!/usr/bin/env python
import pyshark

written_data = []

capture = pyshark.FileCapture('tower_logs.pcapng', display_filter='modbus.func_code == 16')
for packet in capture:
    written_data.append(packet.modbus.reference_num)

for data in written_data:
    print(chr(int(data)), end='')