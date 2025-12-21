from pwn import *

tube = tubes.serialtube.serialtube(port='/dev/ttyUSB0')
# print(tube.recvuntil(b'\x1b[0;32mnsec> \x1b[0m'))

pin = 0
while True:
    tube.sendline(f'debug enable {pin}'.encode())
    tube.recvline()
    tube.recvline()
    result = tube.recvline()
    print(f'result = {result}, pin = {pin}')
    if b'Debug mode not enabled' not in result and pin != 0:
        print('Stopping.')
        break
    # tube.recvline()
    tube.recvuntil(b'> ')
    pin += 1


# Pin is 1868