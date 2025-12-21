#!/bin/bash

input_file="rockyou.txt"
prefix="rockyou_split.txt"
jump=6

# Initialize output files
for ((i=0; i<jump; i++)); do
    > "${prefix}_${i}.txt"  # Empty each output file if it exists
done

# Read and distribute lines
i=0
while IFS= read -r line; do
    file_index=$((i % jump))
    echo "$line" >> "${prefix}_${file_index}.txt"
    ((i++))
done < "$input_file"