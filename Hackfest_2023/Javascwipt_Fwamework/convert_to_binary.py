#!/usr/bin/env python

with open('ledgitjs.js', 'r') as f:
    file_data = f.read()

binary = ''
counter = 0
for char in file_data:
    counter += 1
    if char == '\t':
        binary += '0'
    elif char == ' ':
        binary += '1'
    else:
        print(f'Error! Encountered {char}. Counter = {counter}')
        exit()

with open('ledgitjs.bin.2', 'w') as f:
    f.write(binary)