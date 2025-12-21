import dis

def outer():
    x = 10
    def inner():
        nonlocal x
        x = 20
    return inner

dis.dis(outer, show_offsets=True)
print(dis.show_code(outer))
print(outer.__code__.co_code.hex())
for const in outer.__code__.co_consts:
    if isinstance(const, type(outer.__code__)):
        print(const.co_code.hex())
    else:
        print(f'const: {const}')


# Parent Bytecode: 
# 5e01
# 9500
# 5301
# 6d01
# 5501
# 3401
# 5302
# 1a00
# 6a08
# 6e00
# 5500
# 2400
#
# Child Bytecode:
# const: None
# const: 10
# 3e01
# 9500
# 5301
# 6d00
# 6700
# 
#   --          0       MAKE_CELL                1 (x)
# 
#    3          2       RESUME                   0
# 
#    4          4       LOAD_CONST               1 (10)
#               6       STORE_DEREF              1 (x)
# 
#    5          8       LOAD_FAST                1 (x)
#              10       BUILD_TUPLE              1
#              12       LOAD_CONST               2 (<code object inner at 0x7f6d488856f0, file "/home/morpheus/Downloads/FlareOn2025/2_-_project_chimera/dis-test.py", line 5>)
#              14       MAKE_FUNCTION
#              16       SET_FUNCTION_ATTRIBUTE   8 (closure)
#              18       STORE_FAST               0 (inner)
# 
#    8         20       LOAD_FAST                0 (inner)
#              22       RETURN_VALUE
# 
# Disassembly of <code object inner at 0x7f6d488856f0, file "/home/morpheus/Downloads/FlareOn2025/2_-_project_chimera/dis-test.py", line 5>:
#   --          0       COPY_FREE_VARS           1
# 
#    5          2       RESUME                   0
# 
#    7          4       LOAD_CONST               1 (20)
#               6       STORE_DEREF              0 (x)
#               8       RETURN_CONST             0 (None)
# Name:              outer
# Filename:          /home/morpheus/Downloads/FlareOn2025/2_-_project_chimera/dis-test.py
# Argument count:    0
# Positional-only arguments: 0
# Kw-only arguments: 0
# Number of locals:  1
# Stack size:        2
# Flags:             OPTIMIZED, NEWLOCALS
# Constants:
#    0: None
#    1: 10
#    2: <code object inner at 0x7f6d488856f0, file "/home/morpheus/Downloads/FlareOn2025/2_-_project_chimera/dis-test.py", line 5>
# Variable names:
#    0: inner
# Cell variables:
#    0: x
# None