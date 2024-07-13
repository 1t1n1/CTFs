#!/usr/bin/python3

# This is the value of the variable after its last assignation
zf3dede9f2b604c66b0343ed1d25408da = '554c4354462d6d34637230355f63346e5f63306e7434316e5f73336372337435'

# I then convert every pack of two alphanumeric character to the corresponding ascii character (assuming that the alphanumeric couple is in hexadecimal representation) and obtain the flag!
answer = "".join([chr(int(zf3dede9f2b604c66b0343ed1d25408da[i:i+2], 16)) for i in range(0, len(zf3dede9f2b604c66b0343ed1d25408da), 2)])
print(answer)
