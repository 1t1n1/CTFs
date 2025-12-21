#!/usr/bin/python3
from pwn import *

REMOTE = False
ip = '127.0.0.1'
port = 80
filepath = './changeme'
breakpoints = ['main']

gdbscript = '''
c
'''
gdbscript = "\n".join(['b *' + bp for bp in breakpoints]) + gdbscript

elf = context.binary = ELF(filepath, checksec=False)
context.terminal = ['tmux', 'splitw', '-h', '-F' '#{pane_pid}', '-P']
if REMOTE is True:
    p = remote(ip, port)
else:
    p = gdb.debug(filepath, gdbscript=gdbscript)
    ### OR if you want to start the process alone and attach gdb after
    # p = process(filepath)
    # gdb.attach(p)
