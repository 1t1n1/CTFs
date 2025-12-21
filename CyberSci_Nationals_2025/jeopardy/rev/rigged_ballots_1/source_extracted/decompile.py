import marshal, dis

with open('source.pyc', 'rb') as f:
    f.read(16)  # skip .pyc header
    code = marshal.load(f)

print(dis.dis(code))
