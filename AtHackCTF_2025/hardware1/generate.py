with open('template.nfc', 'r') as f:
    template = f.read()

inputs = [0x00000001, 0x01030307, 0x00001337, 0x13370000, 0x00000539 , 0x05390000, 0x00010203, 0x01020304, 0x01010101, 0x04040404, 0x09090909, 0xdeadbeef, 0xcafebabe, 0x6c656574, 0x726f6f74, 0x524f4f54, 0x52303037, 0x424f5353, 0x4c454554, 0x524f4745, 0x526f6765, 0x726f6765, 0x524f4752, 0x526f6772, 0x726f6772, 0x30313233, 0x31323334, 0x31313131, 0x34343434, 0x39393939, 0x41444d49, 0x41646d69, 0x61646d69, 0x5355444f, 0x5375646f, 0x7375646f, 0x646f6173, 0x446f6173, 0x444f4153]

def bcc(hexa):
    bcc = 0
    for h in hexa:
        bcc ^= int(h, 16)
    hexa = hex(bcc)[2:]
    return '0' * (2 - len(hexa)) + hexa

i = 11
for hexa in inputs:
    hexa = hex(hexa)[2:]
    hexa = '0' * (8 - len(hexa)) + hexa
    hexa = [hexa[i:i+2] for i in range(0, len(hexa), 2)]
    out = template.replace('--------', " ".join(hexa))
    out = out.replace('++++++++++++++', " ".join(hexa) + ' ' + bcc(hexa))
    with open(f'test{i}.nfc', 'w') as f:
        f.write(out)
    i += 1
