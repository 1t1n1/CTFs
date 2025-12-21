import string

with open('flags_static.txt', 'r') as f:
    flags = f.readlines()

flag_template = flags[0].strip()
for i in range(72):
    current_char_template = flag_template[i]
    if current_char_template == '{' or current_char_template == '}':
        print(current_char_template, end="")
        continue
    
    # La bonne lettre n'est pas dans le flag (à la bonne place), donc si je n'arrive pas à trouver F par exemple, c'est que c'est la bonne lettre, puis passe au prochain.

    if current_char_template.isupper():
        for test_uppercase in string.ascii_uppercase:
            letter_found = False

            for flag in flags:
                flag_stripped = flag.strip()
                if flag_stripped[i] == test_uppercase:
                    letter_found = True
                    break
            if not letter_found:
                print(test_uppercase, end="")
                break
    else:
        for test_lowercase in string.ascii_lowercase:
            letter_found = False

            for flag in flags:
                flag_stripped = flag.strip()
                if flag_stripped[i] == test_lowercase:
                    letter_found = True
                    break
            if not letter_found:
                print(test_lowercase, end="")
                break

# FLAG{OISJDfoijoIwDjpOapOXKcnNKXPLAsfkMNWJjbdhYVBUyxcJKnlMePOIRFUGIshdJv}