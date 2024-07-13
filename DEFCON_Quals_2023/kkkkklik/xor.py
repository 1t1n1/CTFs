key = 3

with open('kkkkklik-data.bin', 'rb') as f:
    xored_data = ''
    data = f.read(1)
    while data != b'':
        print(f'data = {data}')
        xored_data += chr(3 ^ int.from_bytes(data))
        data = f.read(1)

    with open('output.bin', 'wb') as w:
        w.write(xored_data.encode())

