with open('Ballots_2.txt', 'r') as file:
    ballots = file.readlines()

dec_names = ''
for shift in range(101):
    dec_names += f'###### Shift: {shift} ######\n'
    for enc_name in ballots:
        dec_name = ''
        enc_name = enc_name.strip()
        for i in range(len(enc_name)):
            if ord(enc_name[i]) - shift >= 32:
                dec_name += chr(ord(enc_name[i]) - shift)
            else:
                dec_name += enc_name[i]
        dec_names += dec_name + '\n'
    dec_names += '\n\n'

with open(f'Recovered_Flag.txt', 'w') as fp:
    fp.write(dec_names)
