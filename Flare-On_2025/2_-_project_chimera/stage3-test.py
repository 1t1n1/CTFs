import marshal
import dis
import types
from uncompyle6.main import decompile

with open('stage3.marshal', 'rb') as f:
    marshalled_code = f.read()

code = marshal.loads(marshalled_code)
print(code.co_code.hex())
for const in code.co_consts:
    if isinstance(const, types.CodeType):
        print(const.co_code.hex())
    else:
        print(f'const: {const}')
# bytecode = dis.Bytecode(code)
# for instr in bytecode:
#     print(instr.opname, instr.argrepr)


# code = marshal.loads(marshalled_code)
# print(code.co_code.hex())
# print(code.co_names)
# print(code.co_varnames)
# print(code.co_consts)
# print(code.co_filename)
# print(code.co_name)
# print(code.co_firstlineno)
# print(code.co_lines)
# print(code.co_stacksize)
# print(code.co_flags)
# print(code.co_argcount)
# print(code.co_posonlyargcount)
# print(code.co_kwonlyargcount)
# print(code.co_nlocals)
# print(code.co_freevars)
# print(code.co_cellvars)
# 
# for const in code.co_consts:
#     if isinstance(const, types.CodeType):
#         print('sdcanlksadnckjla')
#         print("\nInspecting code object:", const.co_name)
#         print("co_code:", const.co_code.hex())
#         print("co_names:", const.co_names)
#         print("co_varnames:", const.co_varnames)
#         print("co_consts:", const.co_consts)
#         print("co_filename:", const.co_filename)
#         print("co_firstlineno:", const.co_firstlineno)
#         print("co_stacksize:", const.co_stacksize)
#         print("co_flags:", const.co_flags)
#         print("co_argcount:", const.co_argcount)
#         print("co_posonlyargcount:", const.co_posonlyargcount)
#         print("co_kwonlyargcount:", const.co_kwonlyargcount)
#         print("co_nlocals:", const.co_nlocals)
#         print("co_freevars:", const.co_freevars)
#         print("co_cellvars:", const.co_cellvars)

# import sys
# decompile(code, (3, 8), sys.stdout)



# for a, b, c, d in code.co_positions():
#     print(a, b, c, d)

# The code object is corrupted. I must read the marshal by myself.