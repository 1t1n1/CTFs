from bytecodeparse import BytecodeParser

def extract_bytecode(pyc_file_path):
    with open(pyc_file_path, 'rb') as f:
        header_size = 16 if sys.version_info >= (3, 7) else 12  # .pyc header size varies by Python version
        f.read(header_size)  # skip the header
        parse = BytecodeParser(f.read())
        while stmt := parser.decompile_stmt():
            print(stmt)

extract_bytecode('example.pyc')

