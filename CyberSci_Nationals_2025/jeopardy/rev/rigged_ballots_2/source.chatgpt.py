from random import randint, shuffle

shift = randint(0, 100)
filename = 'Flag.txt'

with open(filename, 'r') as file:
    FLAG = file.readlines()

order = []
for i in range(37):
    order.append(i)

shuffle(order)
print(order)

new_flag = []

for name in FLAG:
    name = name.strip()
    orig_end = int(len(name))
    # Extend the string by doubling until length >= 37
    while len(name) < 37:
        name = name + name ## La faille est peut etre ici...
    
    # Take a slice and wrap with '{' and '}'
    name = name[:orig_end] + '{' + name[orig_end:orig_end+1] + '}'
    
    # Extend to length 36 characters plus '}'
    name = name[:36] + '}'

    new_name = ''
    for letter in name:
        new_letter = ord(letter) + shift ## ROT
        if new_letter < 126:
            new_letter = new_letter
        # else branch seems to skip modification, so no else action needed
        new_name += chr(int(new_letter))
    print(new_name)
    new_flag.append(new_name[:-1])  # Drop last char (likely '}')

print(new_flag)

reordered_flag = []

for name in new_flag:
    reordered_name = ''
    i = 0
    for letter in name:
        index = order[i]
        reordered_name += name[index]
        i += 1
    reordered_flag.append(reordered_name)

print(reordered_flag)

with open('Ballots.txt', 'w') as fp:
    for ballot in reordered_flag:
        fp.write('%s\n' % ballot)
