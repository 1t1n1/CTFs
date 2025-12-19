import marshal

with open('stage3.marshal', 'rb') as f:
    marshalled_code = f.read()

code_object = marshal.loads(marshalled_code)
exec(code_object)
