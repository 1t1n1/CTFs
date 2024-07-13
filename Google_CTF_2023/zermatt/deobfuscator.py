#!/usr/bin/env python
import re

# TODO : Enlever les `break` a la fin des derniers `if_blocks`
# TODO : Gerer les cas ou t'as deux `while` qui se suivent

class Deobfuscator:
    def __init__(self, file_content):
        self.file_content = file_content
        self.regex_assignation = re.compile(r'([a-zA-Z0-9]+) = (.*)\n') # TODO : Voir si second groupe est bon
        self.tested_variable_regex_right = re.compile(r'if\s?\(\s?([a-zA-Z][a-zA-Z0-9]*)\s?==\s?([0-9]+)\s?\)')
        self.tested_variable_regex_left = re.compile(r'if\s?\(\s?([0-9]+)\s?==\s?([a-zA-Z][a-zA-Z0-9]*)\s?\)')
        self.env = {}

    def extract_if_blocks(self, instruction_block, offset):
        if_blocks = list()
        first_line_number = offset
        # The following lines assumes that there is no instructions between two if of the same level
        # Add the instructions that precede the if blocks
        while True:
            try:
                if_index = instruction_block[offset].index('if (')
                break
            except ValueError:
                if_blocks.append(instruction_block[offset]) # The line is a basic instruction, so just add it
                offset += 1

        preceding_whitespaces = instruction_block[offset][:if_index]
        start_if_regex = re.compile(preceding_whitespaces + r'if .*\n')
        end_if_regex = re.compile(preceding_whitespaces + r'end\n')

        if_block_start_index = current_index = offset
        # Add the actual if blocks
        while True:
            current_index += 1
            current_line = instruction_block[current_index]

            if 'elseif' in current_line:
                # TODO : Handle this in the functions that will receive it
                print("Can't handle elseif")
                return instruction_block

            if end_if_regex.match(current_line):
                if_blocks.append(instruction_block[if_block_start_index: current_index + 1])
                if start_if_regex.match(instruction_block[current_index + 1]) is None:
                    break
                if_block_start_index = current_index + 1

        # Add the last lines that follow the if block
        while True:
            current_index += 1
            current_line = instruction_block[current_index]
            if preceding_whitespaces in current_line:
                if_blocks.append(current_line)
            else:
                break

        return if_blocks

    def if_block_sorting_key(self, if_block):
        right_test_match = self.tested_variable_regex_right.search(if_block[0])
        left_test_match = self.tested_variable_regex_left.search(if_block[0])
        if right_test_match:
            return int(right_test_match.group(2))
        elif left_test_match:
            return int(left_test_match.group(1))

    def sort_if_blocks(self, if_blocks):
        first_blocks = []
        middle_blocks = []
        last_blocks = []
        first_list_encountered = False
        for i, if_block in enumerate(if_blocks):
            if not isinstance(if_block, list):
                if not first_list_encountered:
                    first_blocks.append(if_block)
                else:
                    last_blocks.append(if_block)
            if isinstance(if_block, list):
                first_list_encountered = True
                middle_blocks.append(if_block)

        middle_blocks.sort(key=self.if_block_sorting_key)
        sorted_blocks = first_blocks
        sorted_blocks.extend(middle_blocks)
        sorted_blocks.extend(last_blocks)
        return sorted_blocks

    def handle_while_true(self, instruction_block, offset):
        unobfuscated_lines = []
        if_blocks = self.extract_if_blocks(instruction_block, offset + 1)
        sorted_if_blocks = self.sort_if_blocks(if_blocks)

        amount_of_lines_handled = 1 # First handled line is the while true
        for if_block in sorted_if_blocks:
            if len(if_block) == 1:
                # That's not an actual if_block, it's a simple instruction
                unobfuscated_lines.append(if_block[0])
                amount_of_lines_handled += 1
            else:
                amount_of_lines_handled += 1 # We skipped the if line
                # Iterate over the block, removing the first `if`` and the last `end`
                for i, line in enumerate(if_block[1:len(if_block) - 1]):
                    if 'while true do' in line:
                        new_block, length = self.handle_while_true(if_block, i + 1) # + 1 because the loop skipped the `if`
                        unobfuscated_lines.extend(new_block)
                        amount_of_lines_handled += length
                    else:
                        unobfuscated_lines.append(line)
                        amount_of_lines_handled += 1
                amount_of_lines_handled += 1 # We skipped the if's end
        amount_of_lines_handled += 1 # Last handled line is the while's end
        return unobfuscated_lines, amount_of_lines_handled

    def run(self):
        unobfuscated = list()
        i = 0
        while True:
            if i >= len(self.file_content):
                break

            line = self.file_content[i]
            assignation_match = self.regex_assignation.search(line)
            if 'while true do' in line:
                new_block, length = self.handle_while_true(self.file_content, i)
                unobfuscated.extend(new_block)
                i += length
            elif assignation_match:
                if 'xor' in line:
                    operand = line[line.index('xor'):]
                else:
                    operand = assignation_match.group(2)
                self.env[assignation_match.group(1)] = operand
                i += 1
                # Do if I have problems
                # try:
                #     self.env[assignation_match.group(1)] = eval(operand)
                # except (NameError, SyntaxError):
                #     self.env[assignation_match.group(1)] = operand
            else:
                unobfuscated.append(line)
                i += 1
        
        with open('zermatt-deobfuscated.lua', 'w') as w:
            for line in unobfuscated:
                w.write(line)

if __name__ == '__main__':
    with open('/home/kali/Documents/infosec_notes/CTFs/Google_CTF_2023/zermatt/zermatt-formatted.lua', 'r') as fd:
        file_content = fd.readlines()
    deobfuscator = Deobfuscator(file_content)
    deobfuscator.run()
