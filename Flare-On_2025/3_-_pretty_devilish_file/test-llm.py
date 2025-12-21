import zlib

# Your hex-escaped data as a raw byte string
compressed_hex = (
    r'\x6d\x0d\x0a\xb1\x01\x78\xb8\x6c\x1c\xdd\xc4\x9f\x58\x59\x25\x00\x96\xfc\xda\xa9\xc6\xaf\x92\xdc\xfa\xa8\x3f\x0d\xc6\x3d\xfe\xe8\xdc\xd3\x03\x6d\x8c\xb9\xc1\xa7\xb5\x8d\x95\x1b\x9c\x0c\x46\xb3\x2a\xcc\x84\x54\x2f\x97\xe8\x84\x95\x6c\x57\x38\xda\xf4\xa5\xa4\x2f\x66\x1b\xc8\x41\x96\xf6\xe2\xb4\x7f\x8f\xa1\xa1\xce\x9f\xaf\xb5\xf0\x0d\x0e\x04\xfe\xcd\xd0\x18\xa8\xa4\x7a\x85\x33\xad\xff\x85\x33\x7a\xcc\x75\xd8\x7e\xb3\xf6\xb5\xe6\x6a\x35\x7c\x2e\x45\xc3\x1e\xe4\x6d\x9f\x0e\xf5\x18\x44\xc3\x57\x19\xf1\xbd\xc6\x2a\x67\x85\xce\x14\x84\x86\x11\xbc\x02\x9d\xd2\xed\xd7\x04\xec'
)

# Convert hex escapes to bytes (Python automatically interprets \xHH)
compressed_data = bytes.fromhex(compressed_hex.replace(r'\x', '').replace('\\', ''))
# Alternative: Use codecs.decode if the above doesn't work in your env
# import codecs
# compressed_data = codecs.decode(compressed_hex, 'unicode_escape').encode('latin1')

# Method 1: Simple decompress with raw mode
try:
    decompressed_data = zlib.decompress(compressed_data, wbits=-15)
    print("Decompressed successfully!")
    print(decompressed_data.decode('latin1', errors='replace'))  # PDF is usually Latin-1
except zlib.error as e:
    print(f"Decompression failed: {e}")

# Method 2: Streaming decompressor (better for large streams)
decompressor = zlib.decompressobj(wbits=-zlib.MAX_WBITS)
decompressed_data = decompressor.decompress(compressed_data)
decompressed_data += decompressor.flush()  # Ensure all data is read
print(decompressed_data.decode('latin1', errors='replace'))