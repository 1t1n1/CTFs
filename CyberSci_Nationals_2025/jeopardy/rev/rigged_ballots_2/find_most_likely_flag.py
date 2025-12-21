import os
from collections import Counter

def letters_in_string(s):
    # Count letters including braces, case-insensitive
    return Counter(c.lower() for c in s if c.isalpha() or c in '{}')

def contains_all_letters(line, target_letters):
    line_letters = letters_in_string(line)
    for letter, count in target_letters.items():
        if line_letters[letter] < count:
            return False
    return True

def find_lines_with_flag_letters():
    target = "cybersci{}"
    target_letters = letters_in_string(target)

    for shift in range(101):
        filename = f"Recovered_Flag_{shift}.txt"
        if not os.path.isfile(filename):
            continue

        with open(filename, 'r') as f:
            lines = f.readlines()

        for i, line in enumerate(lines, 1):
            line = line.strip()
            if contains_all_letters(line, target_letters):
                print(f"File: {filename}, Line {i}: {line}")

if __name__ == "__main__":
    find_lines_with_flag_letters()
