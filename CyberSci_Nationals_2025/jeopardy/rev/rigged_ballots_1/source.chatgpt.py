from random import randint
import base64

def xor_ballots(str1, str2):
    if len(str1) != len(str2):
        str1 = str1 + str2[len(str1):]
    return ' '.join(
        hex(ord(a) ^ ord(b)) for a, b in zip(str1, str2)
    )

filename = 'Hail our leader and savior Esperanza.txt'

with open(filename, 'r') as file:
    FLAG = file.readlines()

xor_code = filename[:-4]

fixed_values = []

for name in FLAG:
    uncode = xor_ballots(name.strip(), xor_code)
    encode = base64.b64encode(uncode.encode('ascii'))
    encode = encode.decode('utf-8')
    fixed_values.append(encode.replace('=', ''))

with open('Flag.txt', 'w') as fp:
    for en_ballot in fixed_values:
        fp.write('%s\n' % en_ballot)
