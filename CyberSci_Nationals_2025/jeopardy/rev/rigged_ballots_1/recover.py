import base64

def xor_ballots(str1, str2):
    # XOR two strings character-wise, returning the original string
    if len(str1) != len(str2):
        str1 = str1 + str2[len(str1):]
    return ''.join(
        chr(ord(a) ^ ord(b)) for a, b in zip(str1, str2)
    )

filename = 'Hail our leader and savior Esperanza.txt'
xor_code = filename[:-4]  # Same XOR key used before

def add_base64_padding(s):
    # Base64 strings need to be a multiple of 4 in length
    return s + '=' * (-len(s) % 4)

with open('Flag.txt', 'r') as f:
    encoded_lines = f.read().splitlines()

original_lines = []

for encoded in encoded_lines:
    # Add padding back
    padded = add_base64_padding(encoded)
    # Base64 decode to get XORed hex string (like '0x1a 0x3f ...')
    decoded_bytes = base64.b64decode(padded)
    decoded_str = decoded_bytes.decode('ascii')
    
    # The decoded string is a space-separated hex string, e.g. '0x1a 0x3f 0x...'
    hex_values = decoded_str.split()
    # Convert each hex string to the original character
    xor_str = ''.join(chr(int(h, 16)) for h in hex_values)
    
    # XOR with xor_code again to get original
    original = xor_ballots(xor_str, xor_code)
    original_lines.append(original)

# Save or print recovered lines
with open('Recovered_Hail_our_leader_and_savior_Esperanza.txt', 'w') as f:
    for line in original_lines:
        f.write(line + '\n')

print("Recovered file written to Recovered_Hail_our_leader_and_savior_Esperanza.txt")
