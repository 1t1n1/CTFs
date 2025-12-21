#!/usr/bin/env python3

def decrypt_line(line, shift):
    """Decrypt a line with the given shift"""
    result = ""
    for char in line:
        result += chr(ord(char) - shift)
    return result

def contains_flag_pattern(text):
    """Check if text contains the cybersci flag pattern"""
    return "cybersci{" in text.lower() and "}" in text

def contains_latino_name_pattern(text):
    """Check if text contains patterns typical of Latino names"""
    text_lower = text.lower()
    
    # Common Latino name patterns
    latino_patterns = [
        # Common endings
        "ez", "es", "ón", "án", "ín", "ás", "ís", "ús",
        # Common name parts
        "maria", "josé", "jose", "carlos", "ana", "luis", "juan", "pedro", 
        "miguel", "antonio", "francisco", "manuel", "jesus", "alejandro",
        "fernando", "ricardo", "roberto", "eduardo", "daniel", "jorge",
        "sofia", "isabella", "camila", "valentina", "lucia", "elena",
        # Common surnames
        "garcia", "rodriguez", "martinez", "hernandez", "lopez", "gonzalez",
        "perez", "sanchez", "ramirez", "torres", "flores", "rivera",
        "gomez", "diaz", "reyes", "morales", "jimenez", "alvarez"
    ]
    
    # Check for patterns
    for pattern in latino_patterns:
        if pattern in text_lower:
            return True
    
    # Check for common endings
    words = text_lower.split()
    for word in words:
        if len(word) > 3:
            if (word.endswith('ez') or word.endswith('es') or word.endswith('on') or 
                word.endswith('an') or word.endswith('in') or word.endswith('as') or
                word.endswith('is') or word.endswith('us')):
                return True
    
    return False

def is_readable_text(text):
    """Check if text looks like readable names/words"""
    # Count alphabetic characters
    alpha_count = sum(1 for c in text if c.isalpha())
    total_count = len(text)
    
    if total_count == 0:
        return False
    
    # Should be mostly alphabetic with some spaces/punctuation
    alpha_ratio = alpha_count / total_count
    
    # Check for reasonable character distribution
    has_vowels = any(v in text.lower() for v in 'aeiou')
    has_consonants = any(c in text.lower() for c in 'bcdfghjklmnpqrstvwxyz')
    
    return alpha_ratio > 0.7 and has_vowels and has_consonants

# Extract unique lines from ballot data
unique_lines = [
    "9lvzi{i}9ooiiz9z9}llfzv}ozi}lzfz9lzlf",
    "ZZsieuzi9seiuiZerr{e99r}rzieizrsrreui", 
    "p`zr^e9zkrn9Gpnmoe^nzms}esn{oe99zGzkn",
    "ziie9ez`9}rnr{ezllnereo}e`oz}loo`z9o}",
    "zkezuullees9hqe9okzezls}ze{hzzqzplzop",
    "s^9ezz9nzzu}eun9}tle{lo}s^s{ozz{e9tet",
    "oZzzin^uzosu9iilzzs9z9p}nnzz^9pinz{lz",
    "N\\}LLht|sJg|elLigQszfi{}yItgnxxrnnio{",
    "lclio9{liiz9llezt9ozlze}oezctoc9zltie",
    "zoo`e9{o``s9zzrme9esomr}ermoeeo9sze`r",
    "iqzrzzz9r{znie``oeznzll}zvvre9qeeoizv",
    "This_is_a_0_This_is_a_0_This_is_a_0__"
]

print("VAL VERDE ELECTION DECRYPTION")
print("Looking for: cybersci{...} flag and Latino names")
print("=" * 60)

flag_candidates = []
name_candidates = []

for line_num, line in enumerate(unique_lines):
    print(f"\nLine {line_num + 1}: {line[:40]}...")
    
    for shift in range(1, 101):
        try:
            decrypted = decrypt_line(line, shift)
            
            # Check for flag pattern
            if contains_flag_pattern(decrypted):
                print(f"  *** FLAG FOUND! Shift {shift}")
                print(f"      {decrypted}")
                flag_candidates.append({
                    'line': line_num + 1,
                    'shift': shift,
                    'original': line,
                    'decrypted': decrypted,
                    'type': 'flag'
                })
            
            # Check for Latino names
            elif contains_latino_name_pattern(decrypted) and is_readable_text(decrypted):
                print(f"  Latino name pattern! Shift {shift}")
                print(f"      {decrypted}")
                name_candidates.append({
                    'line': line_num + 1,
                    'shift': shift,
                    'original': line,
                    'decrypted': decrypted,
                    'type': 'name'
                })
            
            # Check for any readable text that might be names
            elif is_readable_text(decrypted) and len(decrypted.strip()) > 10:
                # Count how name-like it is
                words = decrypted.lower().split()
                name_score = 0
                for word in words:
                    if len(word) > 2:
                        # Common name characteristics
                        if word[0].isupper() or any(char in word for char in 'aeiou'):
                            name_score += 1
                
                if name_score >= 2:  # At least 2 name-like words
                    print(f"  Possible names, Shift {shift}: {decrypted}")
                    
        except ValueError:
            continue

print("\n" + "=" * 60)
print("SUMMARY OF FINDINGS")
print("=" * 60)

if flag_candidates:
    print("\nFLAG CANDIDATES:")
    for candidate in flag_candidates:
        print(f"  Line {candidate['line']}, Shift {candidate['shift']}: {candidate['decrypted']}")

if name_candidates:
    print(f"\nNAME CANDIDATES ({len(name_candidates)} found):")
    for candidate in name_candidates:
        print(f"  Line {candidate['line']}, Shift {candidate['shift']}: {candidate['decrypted']}")

# Look for the most common shift value (should be consistent across the file)
if flag_candidates or name_candidates:
    all_candidates = flag_candidates + name_candidates
    shifts = [c['shift'] for c in all_candidates]
    from collections import Counter
    shift_counts = Counter(shifts)
    
    print(f"\nSHIFT FREQUENCY ANALYSIS:")
    for shift, count in shift_counts.most_common():
        print(f"  Shift {shift}: {count} occurrences")
    
    if shift_counts:
        most_common_shift = shift_counts.most_common(1)[0][0]
        print(f"\nMost likely shift value: {most_common_shift}")
        
        print(f"\nDECRYPTING ALL LINES WITH SHIFT {most_common_shift}:")
        print("-" * 50)
        for line_num, line in enumerate(unique_lines):
            try:
                decrypted = decrypt_line(line, most_common_shift)
                print(f"Line {line_num + 1}: {decrypted}")
            except:
                print(f"Line {line_num + 1}: [Decryption error]")

else:
    print("\nNo clear candidates found. Let's try a broader search...")
    
    # Try some manual test cases based on common Latino names
    test_shifts = [13, 17, 20, 25, 32, 42]
    print("\nTesting specific shifts on promising lines:")
    
    for shift in test_shifts:
        print(f"\nShift {shift}:")
        for i, line in enumerate(unique_lines[:5]):  # Test first 5 lines
            try:
                decrypted = decrypt_line(line, shift)
                if is_readable_text(decrypted):
                    print(f"  Line {i+1}: {decrypted}")
            except:
                continue

print("\n" + "=" * 60)
print("NEXT STEPS:")
print("1. Identify the correct shift value from the results above")
print("2. Apply that shift to ALL lines in Ballots.txt")
print("3. Look for the cybersci{...} flag and Latino names")
print("4. Reconstruct the original character order (undo the shuffle)")
print("=" * 60)