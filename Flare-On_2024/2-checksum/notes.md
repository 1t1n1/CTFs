Patcher le binaire à 04A7927 et 04A78CA pour skip au checksum quand maths compris (maybe en profiter pour patch diffing)

https://en.wikipedia.org/wiki/ChaCha20-Poly1305

Nonce size is 24

.data tres large (encrypted flag)

aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa

aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa

rbx : i
rdx : key (remplacer avec base64 decoded)
rcx : key_size
rax : encrypted (remplacer avec user defined location)

mov rcx, 0x40; mov rbx, 0; alloc; mov rdx, $result; alloc; mov rax, $result;
memset 0x0001FDE9570000, 45; memset 0x0001FDE9570001, 46; memset 0x0001FDE9570002, 47; memset 0x0001FDE9570003, 48; memset 0x0001FDE9570004, 49; memset 0x0001FDE9570005, 4a; memset 0x0001FDE9570006, 4b; memset 0x0001FDE9570007, 4c; memset 0x0001FDE9570008, 4d; memset 0x0001FDE9570009, 4e; memset 0x0001FDE957000a, 4f; memset 0x0001FDE957000b, 50; memset 0x0001FDE957000c, 51; memset 0x0001FDE957000d, 52; memset 0x0001FDE957000e, 53; memset 0x0001FDE957000f, 54; memset 0x0001FDE9570010, 55; memset 0x0001FDE9570011, 56; memset 0x0001FDE9570012, 57; memset 0x0001FDE9570013, 58; memset 0x0001FDE9570014, 59; memset 0x0001FDE9570015, 5a; memset 0x0001FDE9570016, 5b; memset 0x0001FDE9570017, 5c; memset 0x0001FDE9570018, 5d; memset 0x0001FDE9570019, 5e; memset 0x0001FDE957001a, 5f; memset 0x0001FDE957001b, 60; 

45 46 47 48 49 4a 4b 4c 4d 4e 4f 50 51 52 53 54 55 56 57 58 59 5a

memset 0x0001FDE9570000, 45

Expected key: 7fd7dd1d0e959f74c133c13abb740b9faa61ab06bd0ecd177645e93b1e3825dd

(Using reverse XOR in the z_decrypt_flag). I modified the function to automatically decrypt the checksum.
1. Set execution at `0x00000000004A780A`
2. Place breakpoints to catch errors. Most important: `0x00000000004A783B`
3. Run `mov rcx, 0x40; mov rbx, 0; alloc; mov rdx, $result; alloc; mov rax, $result;` in x64dbg command
4. Run `memset 0x000001C87C120000, 0x71; memset 0x000001C87C120001, 0x0a; memset 0x000001C87C120002, 0x05; memset 0x000001C87C120003, 0x45; memset 0x000001C87C120004, 0x01; memset 0x000001C87C120005, 0x2b; memset 0x000001C87C120006, 0x5f; memset 0x000001C87C120007, 0x56; memset 0x000001C87C120008, 0x00; memset 0x000001C87C120009, 0x57; memset 0x000001C87C12000a, 0x0d; memset 0x000001C87C12000b, 0x73; memset 0x000001C87C12000c, 0x55; memset 0x000001C87C12000d, 0x07; memset 0x000001C87C12000e, 0x45; memset 0x000001C87C12000f, 0x51; memset 0x000001C87C120010, 0x2c; memset 0x000001C87C120011, 0x5f; memset 0x000001C87C120012, 0x01; memset 0x000001C87C120013, 0x03; memset 0x000001C87C120014, 0x51; memset 0x000001C87C120015, 0x05; memset 0x000001C87C120016, 0x75; memset 0x000001C87C120017, 0x0d; memset 0x000001C87C120018, 0x03; memset 0x000001C87C120019, 0x10; memset 0x000001C87C12001a, 0x52; memset 0x000001C87C12001b, 0x7b; memset 0x000001C87C12001c, 0x5e; memset 0x000001C87C12001d, 0x50; memset 0x000001C87C12001e, 0x09; memset 0x000001C87C12001f, 0x54; memset 0x000001C87C120020, 0x55; memset 0x000001C87C120021, 0x27; memset 0x000001C87C120022, 0x5a; memset 0x000001C87C120023, 0x50; memset 0x000001C87C120024, 0x13; memset 0x000001C87C120025, 0x07; memset 0x000001C87C120026, 0x7f; memset 0x000001C87C120027, 0x58; memset 0x000001C87C120028, 0x50; memset 0x000001C87C120029, 0x54; memset 0x000001C87C12002a, 0x02; memset 0x000001C87C12002b, 0x51; memset 0x000001C87C12002c, 0x25; memset 0x000001C87C12002d, 0x08; memset 0x000001C87C12002e, 0x50; memset 0x000001C87C12002f, 0x45; memset 0x000001C87C120030, 0x52; memset 0x000001C87C120031, 0x79; memset 0x000001C87C120032, 0x5a; memset 0x000001C87C120033, 0x07; memset 0x000001C87C120034, 0x55; memset 0x000001C87C120035, 0x0b; memset 0x000001C87C120036, 0x07; memset 0x000001C87C120037, 0x24; memset 0x000001C87C120038, 0x5d; memset 0x000001C87C120039, 0x04; memset 0x000001C87C12003a, 0x41; memset 0x000001C87C12003b, 0x5d; memset 0x000001C87C12003c, 0x7d; memset 0x000001C87C12003d, 0x5b; memset 0x000001C87C12003e, 0x56; memset 0x000001C87C12003f, 0x54` where the memory is replaced by the value of rdx.
5. Make sure to follow $rax in dump to catch resulting string.
6. Run and collect string from dump
The value set in memory at 4. is the decoded base64.
**TODO: See if 4 could have been a single command**
