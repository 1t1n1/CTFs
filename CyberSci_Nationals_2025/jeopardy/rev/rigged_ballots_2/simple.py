def subtract_0x19_from_string(input_string):
    result = ''
    for char in input_string:
        new_char_code = ord(char) - 0x19  # Subtract 25
        # Ensure the new character code is valid
        if new_char_code < 0:
            new_char_code = ord(char)
        result += chr(new_char_code)
    return result

# Example usage
original = """
9lvzi{i}9ooiiz9z9}llfzv}ozi}lzfz9lzlf
iqzrzzz9r{znie``oeznzll}zvvre9qeeoizv
lclio9{liiz9llezt9ozlze}oezctoc9zltie
N\}LLht|sJg|elLigQszfi{}yItgnxxrnnio{
oZzzin^uzosu9iilzzs9z9p}nnzz^9pinz{lz
p`zr^e9zkrn9Gpnmoe^nzms}esn{oe99zGzkn
p`zr^e9zkrn9Gpnmoe^nzms}esn{oe99zGzkn
s^9ezz9nzzu}eun9}tle{lo}s^s{ozz{e9tet
This_is_a_0_This_is_a_0_This_is_a_0__
ziie9ez`9}rnr{ezllnereo}e`oz}loo`z9o}
zkezuullees9hqe9okzezls}ze{hzzqzplzop
zoo`e9{o``s9zzrme9esomr}ermoeeo9sze`r
ZZsieuzi9seiuiZerr{e99r}rzieizrsrreui
"""
transformed = subtract_0x19_from_string(original)
print("Transformed:", transformed)
