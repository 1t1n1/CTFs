#!/bin/python3

import string
import itertools

# checksum  =  s => {                                                                                                       
#   let result = '';                                                                                                        
#   let x = 0;                                                                                                              
#   try {                                                                                                                   
#     for (let x = 0; x+3 <= s.length; x+=3) {                                                                              
#       let next =                                                                                                          
#         (s.charCodeAt(x)%2)*64 +                                                                                          
#         (s.charCodeAt(x+1)%8)*8 +                                                                                         
#         s.charCodeAt(x+2)%8;                                                                                              
#       next = Math.min(Math.max(0x20, next), 0x7E);                                                                        
#       result += String.fromCharCode(next);                                                                                
#     }                                                                                                                     
#     return result;                                                                                                        
#   } catch(_) {                                                                                                            
#     throw ChecksumError(s, result, x);                                                                                    
#   }                                                                                                                       
# }
#   .?  K 7 hA  [Cdml<U}9P  @dBpM) -$A%!X5[ '% U(!_ (]c 4zp$RpUi(mv!u4!D%i%6!D'Af$Iu8HuCP>qH.*(Nex.)X&{I'$ ~Y0mDPL1 U08<2G{ ~ _:h\ys! K A( f.'0 p!s    fD] (  H  E < 9Gf.' XH,V1 P * -P
# According to the pool thing, the first 27 characters of the flag should NOT be 01Kb3W_5l__9LUzNcH3cu1dw_s_

def decrypt(char):
    possible_values = list()
    accepted_characters = string.digits + string.ascii_letters + '_@!?-'
    encrypt = lambda a, b, c : chr(min(max(32, ((a % 2) * 64 + (b % 8) * 8 + (c % 8))), 126))
    for i in accepted_characters:
        i = ord(i)
        for j in accepted_characters:
            j = ord(j)
            for k in accepted_characters:
                k = ord(k)
                if encrypt(i, j, k) == char:
                    possible_values.append(chr(i) + chr(j) + chr(k))
    return possible_values



def sort_passwords(chars_list):
    pool = 'c_3L9zKw_l1HusWN_b_U0c3d5_1'
    pool = [pool[i] for i in range(len(pool))]
    i = 1337
    for j in range(len(pool)):
        i = (i if i != 0 else 1) * 16807 % 2147483647 # Erreur possible: Mal géré le i (l'avoir modifié trop tôt, alors qu'il aurait dû être modifié à la fin de la loop)
        pool_i = pool.pop(i % len(pool))
        current_block = j // 3
        current_letter = j % 3
        k = 0
        while (k < len(chars_list[current_block])):
            current_char = chars_list[current_block][k][current_letter]
            if chars_list[current_block][k][current_letter] != pool_i:
                chars_list[current_block].pop(k)
            else:
                k += 1
    return chars_list



def get_possible_passwords(pass_list):
    possible_passwords = [i for i in itertools.permutations(pass_list, len(pass_list))]
    # for block in range(len(pass_list)):
    #     for chars in range(len(block)):
    return possible_passwords



if __name__ == "__main__":
    encrypted_password = "   .?  K 7 hA  [Cdml<U}9P  @dBpM) -$A%!X5[ '% U(!_ (]c 4zp$RpUi(mv!u4!D%i%6!D'Af$Iu8HuCP>qH.*(Nex.)X&{I'$ ~Y0mDPL1 U08<2G{ ~ _:h\ys! K A( f.'0 p!s    fD] (  H  E < 9Gf.' XH,V1 P * -P"
    passwords = list()
    for character in encrypted_password:
        print(f'Decrypting {character}...')
        decrypted_characters = decrypt(character)
        passwords.append(decrypted_characters)
    passwords = sort_passwords(passwords)
    # passwords = get_possible_passwords(passwords)
    print("".join([passwords[i][0] for i in range(len(passwords))]))
    # with open('output.txt', 'w') as f:
    #     for i in range(len(passwords)):
    #         f.write(passwords[i])