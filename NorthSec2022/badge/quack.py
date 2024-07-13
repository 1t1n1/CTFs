from pwn import *

QUACK_DATA = [b'\x1b[?1049hquackquackquickquack\x1b[H\x1b[2J\x1b[3Jquickquickquack\x1b[H\x1b[2J\x1b[3Jquickquack\x1b[H\x1b[2J\x1b[3Jquackquickquackquick\x1b[H\x1b[2J\x1b[3Jquackquickquack\x1b[H\x1b[2J\x1b[3Jquickquick\x1b[H\x1b[2J\x1b[3Jquackquack\x1b[H\x1b[2J\x1b[3Jquickquack\x1b[H\x1b[2J\x1b[3Jquackquickquick\x1b[H\x1b[2J\x1b[3Jquickquickquack\x1b[H\x1b[2J\x1b[3Jquackquickquackquick\x1b[H\x1b[2J\x1b[3Jquackquickquack\x1b[H\x1b[2J\x1b[3Jquack\x1b[H\x1b[2J\x1b[3Jquackquackquack\x1b[H\x1b[2J\x1b[3Jquackquackquick\x1b[H\x1b[2J\x1b[3Jquick\x1b[H\x1b[2J\x1b[3Jquack\x1b[H\x1b[2J\x1b[3Jquackquickquickquick\x1b[H\x1b[2J\x1b[3Jquickquackquick\x1b[H\x1b[2J\x1b[3Jquick\x1b[H\x1b[2J\x1b[3Jquickquack\x1b[H\x1b[2J\x1b[3Jquackquickquick\x1b[H\x1b[2J\x1b[3Jquickquickquack\x1b[H\x1b[2J\x1b[3Jquickquickquick\x1b[H\x1b[2J\x1b[3Jquick\x1b[H\x1b[2J\x1b[3Jquack\x1b[H\x1b[2J\x1b[3Jquickquickquickquick\x1b[H\x1b[2J\x1b[3Jquickquick\x1b[H\x1b[2J\x1b[3Jquickquickquick\x1b[H\x1b[2J\x1b[3Jquickquickquackquick\x1b[H\x1b[2J\x1b[3Jquickquackquickquick\x1b[H\x1b[2J\x1b[3Jquickquack\x1b[H\x1b[2J\x1b[3Jquackquackquick\x1b[H\x1b[2J\x1b[3Jquackquickquickquick\x1b[H\x1b[2J\x1b[3Jquackquick\x1b[H\x1b[2J\x1b[3Jquickquickquickquack\x1b[H\x1b[2J\x1b[3Jquackquackquickquick\x1b[H\x1b[2J\x1b[3Jquackquickquack\x1b[H\x1b[2J\x1b[3Jquickquack\x1b[H\x1b[2J\x1b[3Jquickquackquack\x1b[H\x1b[2J\x1b[3Jquackquack\x1b[H\x1b[2J\x1b[3Jquickquickquick\x1b[H\x1b[2J\x1b[3Jquackquickquackquack\x1b[H\x1b[2J\x1b[3Jquickquickquick\x1b[H\x1b[2J\x1b[3J\r', b'\x1b[?1049l\r']

def get_quack_data(tube):
    print('Talk to the duck so we can retrieve the precious information')
    return tube.recvlines(2)

def filter_quack(string):
    index = 0
    while True:
        next_quack = string.find(b'quack', index)
        next_quick = string.find(b'quick', index)
        if next_quack < next_quick:
            string = string[:index] + b' ' + string[next_quack:]
            index += 1
        else:
            string = string[:index] + b' ' + string[next_quick:]
            index += 1

        while True:
            if not (string[index:index+5] == b'quack' or string[index:index+5] == b'quick'):
                break
            index += 5

        if index >= len(string) - 1:
            break
    return string.strip().decode()

def quack_to_bin(string):
    index = 0
    bin_data = ''
    while True:
        if string[index:index+1] == '':
            break
        elif string[index:index+1] == ' ':
            bin_data += ' '
            index += 1
        elif string[index:index+5] == 'quack':
            bin_data += '-'
            index += 5
        elif string[index:index+5] == 'quick':
            bin_data += '.'
            index += 5
    return bin_data

def get_quack_flag(tube):
    ansi_escapes = [b'\x1b[?1049h', b'\x1b[H\x1b[2J\x1b[3J', b'\x1b[?1049l']
    if not QUACK_DATA or len(QUACK_DATA) == 0:
        quack_data = get_quack_data(tube)
    else:
        quack_data = QUACK_DATA
    quack_data = filter_quack(quack_data[0])
    quack_bin = quack_to_bin(quack_data)
    return quack_bin

if __name__== '__main__':
    tube = tubes.serialtube.serialtube(port='/dev/ttyUSB0')
    print(get_quack_flag(tube))