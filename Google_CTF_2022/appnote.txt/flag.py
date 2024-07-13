#!/bin/python3

strings = ''
with open('/home/morpheus/Downloads/google_ctf_temp/appnote.txt/strings.txt', 'r') as f:
    strings = f.readlines()

flag = ''
is_flag = False
for string in strings:
    if not string.startswith('flag'):
        continue
    interesting_char = string[6] 
    if interesting_char == '{':
        is_flag = True
    if len(string) > 9 and is_flag:
        flag += interesting_char
    if interesting_char == '}':
        break
print('Flag is ' + flag)
