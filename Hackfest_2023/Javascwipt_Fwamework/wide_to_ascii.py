#!/usr/bin/env python

with open('decompressed.hex', 'r') as f:
    file_data = f.read()

counter = 0
hex_ascii_data = ''
for i in range(0, len(file_data), 4):
    hex_ascii_data += file_data[i+2:i+4]

with open('decompressed_ascii.hex', 'w') as f:
    f.write(hex_ascii_data)