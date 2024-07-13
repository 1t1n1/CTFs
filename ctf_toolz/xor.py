#!/usr/bin/env python3

import argparse
from pwn import xor

def xor_string(input_string, xor_key):
    # Ensure the XOR key is as long as the input string
    key_len = len(xor_key)
    input_len = len(input_string)
    if key_len < input_len:
        xor_key *= (input_len // key_len) + 1
    xor_key = xor_key[:input_len]

    # Perform the XOR operation
    result = xor(input_string, xor_key)
    return result

if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="XOR a string with a key using pwntool's xor.")
    parser.add_argument("input_string", type=str, help="The input string to XOR.")
    parser.add_argument("xor_key", type=str, help="The XOR key (can be a string or bytes).")

    args = parser.parse_args()

    # Check if the xor_key starts with '0x' and convert it to bytes if needed
    if args.xor_key.startswith("0x"):
        xor_key = bytes.fromhex(args.xor_key[2:])
    else:
        xor_key = args.xor_key.encode()

    result = xor_string(args.input_string.encode(), xor_key)

    print("Input String:", args.input_string)
    print("XOR Key:", args.xor_key)
    print("Result (ascii):", result.decode())

    print("Result (hex):", end=' ')
    for byte in result:
        print(hex(byte)[2:], end=' ')
