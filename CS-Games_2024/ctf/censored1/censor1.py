from random import random

def censor1(flag):
    out = ""
    key = []
    for c in flag:
        if c.isalpha():
            rand = 0
            while not rand:
                rand = int(random() * 26)
            start = ord('A') + (ord(c) & 32) # Start at 'A' or 'a' depending on case
            num = ord(c)
            num -= start   # Ici on a le numero de la lettre
            num += rand
            num %= 26
            num += start   # On remet la lettre a sa case
            out += chr(num)
                           # Donc le case est conservé
            key.append(rand)
        else:
            out += c
    print(f"key: {key}")
    return out
flag = "SWTJ{VEKFUjxbiuOkHqeMemUCAouBVIIYLfyjXRZIkzjaQGZSjmzHLhgBtMTDWONAWrzqSw}"

censor1(flag)