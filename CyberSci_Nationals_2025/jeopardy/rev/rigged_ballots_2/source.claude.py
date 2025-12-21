from random import randint, shuffle

shift = randint(0, 100)

filename = 'Flag.txt'
with open(filename, 'r') as file:
    FLAG = file.readlines()

order = []
for i in range(0, 37):
    order.append(i)

shuffle(order)
print(order)

new_flag = []
for name in FLAG:
    new_name = ''
    name = name.strip()
    orig_end = int(len(name))
    while len(name) < 37:
        # The names are added in a loop, but then cut later on. This means 
        # that the letters that appear more often than the others are at the
        # beginning of the word
        name = name + name 
    
    name = name[:orig_end] + '{' + name[orig_end + 1:]
    name = name[:36] + '}'
    
    for letter in name:
        new_letter = ord(letter) 
        if new_letter + shift < 126: # Donc dans les ballots on peut pas savoir si le char a été shifté ou pas...
            new_letter = new_letter + shift
        
        new_name = new_name + chr(int(new_letter))
    print(new_name)
    new_flag.append(new_name[:-1])

print(new_flag)

reordered_flag = []
for name in new_flag:
    reordered_name = ''
    i = 0
    for letter in name:
        index = order[i]
        reordered_name = reordered_name + name[index]
        i = i + 1
    reordered_flag.append(reordered_name)

print(reordered_flag)
with open('Ballots.txt', 'w') as fp:
    for ballot in reordered_flag:
        fp.write('%s\n' % ballot)