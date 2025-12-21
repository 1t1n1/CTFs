import marshal

a = compile('b = 1 + a', '<string>', 'exec')
marshalled = marshal.dumps(a)
with open("marshalled_code.bin", "wb") as f:
    f.write(marshalled)