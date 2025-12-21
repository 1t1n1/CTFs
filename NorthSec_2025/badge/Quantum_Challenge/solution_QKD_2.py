init = [i for i in '10010010011101100110111010101110000110000000000010110000100110']

def invert(n):
    prev = init[n]
    if prev == '0':
        init[n] = '1'
    else:
        init[n] = '0'

def print_init():
    print("".join(init))

# ITERATION 1
# pass 1
invert(14)
invert(25)
invert(46)
print_init()

# ITERATION 2
# pass 1
invert(20)
print_init()

# pass 2
invert(18)
print_init()

# ITERATION 3
# pass 1
invert(5)
invert(6)
print_init()