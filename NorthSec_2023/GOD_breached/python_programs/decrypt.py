#!/usr/bin/env python

# IMPORTANT NOTE: This doesn't work
# The intended solution is to call the program with a fake server

key = b'Oogum_Boogum'
with open('new/unknown2', 'rb') as input_f:
    data = ''
    i = 0
    while True:
        next_byte = input_f.read(1)
        # print(f'next_byte = {next_byte}')
        if next_byte == b'':
            break
        data += chr(int.from_bytes(next_byte, 'big') ^ key[(i + 3) % len(key)])
        i += 1
    with open('output.bin', 'w') as output_fd:
        output_fd.write(data)