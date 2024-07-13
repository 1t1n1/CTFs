Niveau pour la compression 0

On décompile le programme pour voir comment il fonctionne

r12 = c (ou c:)

main():
```c
while (eax != -1) {
	if (eax != 'c') {
		goto label_0
	}
}
```

Utiliser pacat ou paplay pour rouler le playback d'un fichier par pulseaudio:
```bash
$ pacat flag0.maplewave --rate 3000 --channels 3
# Donne un son correct. Je crois que l'audio répète deux fois la même chose

```

Note: On remarque qu'au final, le programme lit l'entrée audio avec pa_simple_read() et écrit dans le fichier de output ce que la fonction retourne dans son buffer.

J'ai donné une taille plus grande que réelle pour la taille du fichier et de la section data, mais c'est peut être mieux de donner une taille plus petite que réel


0x001017a4: L'adresse de l'appel intéressant dans start_encoding. C'est l'endroit où on cherche à se tenir pour analyser le stack

0x0000000000101160: Adresse du main (à partir de 0x00100000)
0x0000555555555160: Adresse du main (dans gdb)

0x00005555555557a4: L'adresse de l'appel intéressant (dans gdb)

---
##### C Calling Convention
Selon The Ghidra Book - The Definitive Guide (2020):
"_For 64-bit x86 binaries, `cdecl` varies by operating system; on Linux, up to six arguments are placed in registers RDI, RSI, RDX, RCX, R8, and R9, in that order, and any additional arguments spill onto the stack._"
Ainsi, lorsqu'on utilise le debugger de Ghidra, et qu'on met un breakpoint à l'adresse 0x00005555555557a4 (`(gdb) *0x00005555555557a4`) on peut observer ceci (Note: Expliquer comment j'ai trouvé cette adresse):

Reg | Value                   |  Official Name    |  Variable Name
RDI = 0                         = server
RSI = 555555557019  = name                 = &application_name
RDX = 2                        = dir
RCX = 0                        = dev
R8 = 555555557019    = stream_name   = &application_name
R9 = 5555555590d8   = ss                      = &sample_type

Note: On peut aussi simplement double cliquer sur sample_type

Quand on regarde ce qui se trouve à l'adresse de sample_type, on voit qu'il y a des informations. En se fiant aux offsets qui seraient donnés par https://www.freedesktop.org/software/pulseaudio/doxygen/sample_8h_source.html pour la struct pa_sample_spec, on déduit que le rate est de 16000 (0x3E80) et qu'un seul channel est utilisé. Je vais d'abord tester avec un BitsPerSample de 16, puis de 8 et si ça ne fonctionne toujours pas je peux continuer à jouer avec celui-ci.

Le header de fichier (avec un BitsPerSample de 8) (voir https://docs.fileformat.com/audio/wav/):
52 49 46 46 1e f7 00 00 57 41 56 45 66 6d 74 20 10 00 00 00 01 00 01 00 80 3e 00 00 80 3e 00 00 01 00 08 00 64 61 74 61 1e f7 00 00

Lorsqu'on modifie le fichier pour lui donner ce header, on obtient un fichier qui s'exécute avec VLC et on comprend très clairement ce qu'il dit

Flag:
maple{easy post code modulation 5716}