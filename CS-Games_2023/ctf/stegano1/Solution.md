The solution was that the image's height was shorter than the actual image could display. We thus needed to modify the image's height in the header. I did it with xxd
1. `xxd <original_image> > image.xxd`
2. Modify the right location (0x00000000a3) (This is the location that corresponds the most to the image's height following the 0xFFC0 beacon)
3. `xxd -r image.xxd new_image.xxd` 