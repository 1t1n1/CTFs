# 
#  CyberSci Nationals 2024
#
#  Dmitriy Beryoza (0xd13a@gmail.com)
#
#  Jeopardy competition
#
#  Projector RE challenge solver
#

from z3 import *
import struct
import hashlib

def brute_md5(hash):
	i = 0
	while i <= 0xffffff:
		if hashlib.md5(struct.pack("<I",i)).digest() == hash:
			return i
		i += 1
	return 0


print("Solving Projector 1")

positions = [22, 27, 16, 25, 5, 24, 10, 23, 19, 17, 3, 9, 0, 1, 15, 13, 12, 14, 11, 20, 18, 6, 4, 
  7, 21, 26, 8, 2, 28]

key = [0] * len(positions)

xor_factor = [0x22, 0x26, 0x81, 0xcb, 0x10, 0xf4, 0xd0, 0x0e, 0xef, 0x5a, 0xdb, 0x37, 0x1b, 0x3d, 0x83, 0xf6, 
0x2b, 0x6e, 0xf6, 0x2c, 0x5a, 0x52, 0x0c, 0x75, 0xf9, 0x9b, 0x36, 0x11, 0x54];
  
key_encoded = [0x6c, 0x15, 0xd4, 0xfb, 0x43, 0xd9, 0x86, 0x41, 0xa1, 0x6f, 0x97, 0x1a, 0x5a, 0x74, 0xb4, 0xae, 
0x7a, 0x43, 0xa1, 0x78, 0x13, 0x64, 0x5a, 0x58, 0xc9, 0xac, 0x7f, 0x43, 0x15];

result = ""
for i in range(len(positions)):
	key[positions[i]] = chr(key_encoded[positions[i]] ^ xor_factor[positions[i]])

result = "".join(key)

assert(result == "N3U0S-VON5L-AI7XQ-WTI6V-07IRA")
print(result)


print("\nSolving Projector 2")

val1 = brute_md5(b"\x35\xd6\x58\x39\x75\x22\xb3\xb2\xb0\xc7\x87\xad\x3a\x24\xaf\x77")
val2 = brute_md5(b"\x40\x41\xf2\xc7\x52\xe7\x9a\xe5\x27\x4a\x1b\xb3\x0e\xa0\x93\xd0")
val3 = brute_md5(b"\x83\x18\x68\xa1\xe2\xb8\x5b\x20\x6b\x89\xd5\x2b\x7a\xa9\x53\xf3")
result = "{}".format(struct.unpack('d', struct.pack('Q', val1 << 40 | val2 << 16 | val3))[0])
assert(result == "45.4244367483658")
print(result)
val4 = brute_md5(b"\xc9\x7b\x6a\x3b\xbd\x04\x3a\x5d\x1d\x50\x9a\x2a\xb4\xc0\xc4\xc5")
val5 = brute_md5(b"\xd6\x94\x06\x93\x58\xad\x43\xeb\xaf\x9e\xcb\x88\xe4\xeb\x39\xe2")
val6 = brute_md5(b"\xd7\xfd\x27\x45\x1d\x06\xbf\xa2\xa1\x5b\x30\x72\x2a\x18\xfa\x15")
result = "{}".format(struct.unpack('d', struct.pack('Q', val4 << 40 | val5 << 16 | val6))[0])
assert(result == "-75.69152330448755")
print(result)


print("\nSolving Projector 3")

solver = Solver()

for i in range(35):
	globals()['c%i' % i] = BitVec('c%i' % i, 32)

	solver.add(And(globals()['c%i' % i] >= 32, globals()['c%i' % i] <= 126))

solver.add(globals()['c0'] ^ globals()['c1'] == 0x29)
solver.add(globals()['c1'] ^ globals()['c2'] == 0x0)
solver.add(globals()['c2'] ^ globals()['c3'] == 0x5)
solver.add(globals()['c3'] ^ globals()['c4'] == 0x6)
solver.add(globals()['c4'] ^ globals()['c5'] == 0x1b)
solver.add(globals()['c5'] ^ globals()['c6'] == 0x54)
solver.add(globals()['c6'] ^ globals()['c7'] == 0x61)
solver.add(globals()['c7'] ^ globals()['c8'] == 0x2d)
solver.add(globals()['c8'] ^ globals()['c9'] == 0x8)
solver.add(globals()['c9'] ^ globals()['c10'] == 0x1)
solver.add(globals()['c10'] ^ globals()['c11'] == 0x17)
solver.add(globals()['c11'] ^ globals()['c12'] == 0x1)
solver.add(globals()['c12'] ^ globals()['c13'] == 0x1c)
solver.add(globals()['c13'] ^ globals()['c14'] == 0x1)
solver.add(globals()['c14'] ^ globals()['c15'] == 0x42)
solver.add(globals()['c15'] ^ globals()['c16'] == 0xc)
solver.add(globals()['c16'] ^ globals()['c17'] == 0x70)
solver.add(globals()['c17'] ^ globals()['c18'] == 0x22)
solver.add(globals()['c18'] ^ globals()['c19'] == 0x17)
solver.add(globals()['c19'] ^ globals()['c20'] == 0x16)
solver.add(globals()['c20'] ^ globals()['c21'] == 0x1a)
solver.add(globals()['c21'] ^ globals()['c22'] == 0xd)
solver.add(globals()['c22'] ^ globals()['c23'] == 0x1)
solver.add(globals()['c23'] ^ globals()['c24'] == 0xb)
solver.add(globals()['c24'] ^ globals()['c25'] == 0x1a)
solver.add(globals()['c25'] ^ globals()['c26'] == 0x58)
solver.add(globals()['c26'] ^ globals()['c27'] == 0xc)
solver.add(globals()['c27'] ^ globals()['c28'] == 0x63)
solver.add(globals()['c28'] ^ globals()['c29'] == 0x6e)
solver.add(globals()['c29'] ^ globals()['c30'] == 0x7e)
solver.add(globals()['c30'] ^ globals()['c31'] == 0x3a)
solver.add(globals()['c31'] ^ globals()['c32'] == 0xd)
solver.add(globals()['c32'] ^ globals()['c33'] == 0x1)
solver.add(globals()['c33'] ^ globals()['c34'] == 0x16)
solver.add(globals()['c0'] ^ globals()['c34'] == 0x36)

solver.add(globals()['c0'] + globals()['c1'] == 0xb1)
solver.add(globals()['c1'] + globals()['c2'] == 0xd8)
solver.add(globals()['c2'] + globals()['c3'] == 0xd5)
solver.add(globals()['c3'] + globals()['c4'] == 0xd8)
solver.add(globals()['c4'] + globals()['c5'] == 0xe3)
solver.add(globals()['c5'] + globals()['c6'] == 0x94)
solver.add(globals()['c6'] + globals()['c7'] == 0x61)
solver.add(globals()['c7'] + globals()['c8'] == 0xad)
solver.add(globals()['c8'] + globals()['c9'] == 0xd0)
solver.add(globals()['c9'] + globals()['c10'] == 0xc9)
solver.add(globals()['c10'] + globals()['c11'] == 0xd7)
solver.add(globals()['c11'] + globals()['c12'] == 0xe5)
solver.add(globals()['c12'] + globals()['c13'] == 0xe2)
solver.add(globals()['c13'] + globals()['c14'] == 0xdd)
solver.add(globals()['c14'] + globals()['c15'] == 0x9a)
solver.add(globals()['c15'] + globals()['c16'] == 0x4c)
solver.add(globals()['c16'] + globals()['c17'] == 0x70)
solver.add(globals()['c17'] + globals()['c18'] == 0xc2)
solver.add(globals()['c18'] + globals()['c19'] == 0xd7)
solver.add(globals()['c19'] + globals()['c20'] == 0xd8)
solver.add(globals()['c20'] + globals()['c21'] == 0xdc)
solver.add(globals()['c21'] + globals()['c22'] == 0xcd)
solver.add(globals()['c22'] + globals()['c23'] == 0xc9)
solver.add(globals()['c23'] + globals()['c24'] == 0xd3)
solver.add(globals()['c24'] + globals()['c25'] == 0xe2)
solver.add(globals()['c25'] + globals()['c26'] == 0xa0)
solver.add(globals()['c26'] + globals()['c27'] == 0x4c)
solver.add(globals()['c27'] + globals()['c28'] == 0x63)
solver.add(globals()['c28'] + globals()['c29'] == 0x70)
solver.add(globals()['c29'] + globals()['c30'] == 0x80)
solver.add(globals()['c30'] + globals()['c31'] == 0xbc)
solver.add(globals()['c31'] + globals()['c32'] == 0xcd)
solver.add(globals()['c32'] + globals()['c33'] == 0xc9)
solver.add(globals()['c33'] + globals()['c34'] == 0xd8)
solver.add(globals()['c0'] + globals()['c34'] == 0xb8)


result = str(solver.check())
assert(result == "sat")

modl = solver.model()

res = ""
for i in range(35):
	obj = globals()['c%i' % i]
	res += chr(modl[obj].as_long())
assert(res == "Elliot Alderson, President, C-Sides")
print(res)
