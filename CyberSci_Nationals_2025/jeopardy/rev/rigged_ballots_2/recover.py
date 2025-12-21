order = [5, 0, 12, 1, 8, 3, 11, 9, 6, 10, 2, 7, 4, 13, 14, 15, 16, 17, 18, 19,
         20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36]

inverse_order = [0] * len(order)
for i, o in enumerate(order):
    inverse_order[o] = i

def undo_reorder(s):
    prefix_len = len(order)
    prefix = s[:prefix_len]
    suffix = s[prefix_len:]

    reordered = [''] * prefix_len
    for i, c in enumerate(prefix):
        reordered[inverse_order[i]] = c
    return ''.join(reordered) + suffix

def undo_shift(s, shift):
    result = []
    for c in s:
        code = ord(c)
        if 32 <= code <= 126:
            new_code = code - shift
            while new_code < 32:
                new_code += 95
            result.append(chr(new_code))
        else:
            result.append(c)
    return ''.join(result)

def recover_flag_line(line, shift):
    undone = undo_reorder(line)
    undone = undo_shift(undone, shift)
    # Ensure the line length is exactly 37 characters by padding or trimming
    if len(undone) < 37:
        undone = undone.ljust(37)
    else:
        undone = undone[:37]
    return undone

def main():
    with open('Ballots.txt', 'r') as f:
        ballots = [line.rstrip('\n') for line in f.readlines()]

    for shift in range(101):
        recovered_flag = [recover_flag_line(line, shift) for line in ballots]

        with open(f'Recovered_Flag_{shift}.txt', 'w') as f:
            for line in recovered_flag:
                f.write(line + '\n')

        print(f"Recovered Flag saved in Recovered_Flag_{shift}.txt")

if __name__ == "__main__":
    main()
