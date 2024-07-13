#!/usr/bin/env python3

reserved_values = list()
with open("dump.txt", 'r') as f:
    for line in f.readlines():
        start = line.find("Reserved: ")
        if start != -1:
            value = line[start + 10:-1]
            if not value.startswith("0x") and value != '0':
                reserved_values.append(value)

print(reserved_values)
with open("dump-2.txt", 'w') as f:
    for value in reserved_values:
        f.write(value + '\n')