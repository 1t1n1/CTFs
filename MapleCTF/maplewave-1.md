À la ligne 55 de start_encoding (selon la décompilation de ghidra), on observe un pointeur de fonction qui est en fait utilisé pour écrire les données capturées par pulseaudio dans un fichier. La variable qui se trouve dans les crochets correspond au codec utilisé. Ainsi, si on utilise le codec 0, on va chercher 
```c
&ptr_de_fonction[0] = write_data_to_file
```

À la ligne 55 de start_encoding (selon la décompilation de ghidra), on observe une variable qui possède la même adresse que l'instruction machine
```asm
addr write_data_to_file  // fonction pour compression de niveau 0
addr FUN_001014f0   // fonction pour compression de niveau 1
addr FUN_00101e30   // fonction pour compression de niveau 2
```
Ainsi, lorsqu'on fait l'instruction suivante, on obtient la fonction pour la compression de niveau 0:
```c
&ptr_to_write_data_to_file[0]
```
Mais il est possible de changer l'index pour obtenir les autres fonctions. En fait, l'index correspoond exactement à la valeur que l'on fourni au programme avec l'option -c \<niveau de compression\>


---
J'ai appelé une variable file_descriptor_write_pointer. Pour comprendre ce qu'elle veut dire, voir /usr/include/x86_64-linux-gnu/bits/types/struct_FILE.h à la ligne

---

Nice discovery: When you find something like this:
```c
*(bytes *)(file_descriptor + 0x24)
```
You can right-click it and select Auto Create Structure, which will convert every occurences of this syntax to 
```c
file_descriptor->field16_0x24
```
which is much easier to understand when you know for a fact that you're dealing with a struct!
You can then rename the field to something like this:
```c
file_descriptor->write_base
```

---
uVar1 = (uint)\*ptr_input_data_buffer - (uint)file_descriptor->write_base;
: uVar1 = l'intensité audio à inscrire - la dernière intensité placée à write_base
où 
l'intensité audio à inscrire est un seul octet