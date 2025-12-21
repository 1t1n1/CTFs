import lief
import sys
import capstone

state = int(sys.argv[1], 16)

pe = lief.parse('./ntfsm.exe')
with open('./ntfsm.exe', 'rb') as f:
    code = f.read()

def print_hex_colored(string):
    hex_start = string.find(', 0x')
    if hex_start == -1:
        print(string)
        return
    print(string[:hex_start+2], end='')
    print(f'\033[92m{string[hex_start+2:]}\033[0m')

def get_uint_at_va(addr):
    offset = pe.rva_to_offset(addr - pe.imagebase)
    return int.from_bytes(code[offset:offset+4], 'little')

def get_next_states_from_disass(disass):
    delim = 'mov qword ptr [rsp + 0x58d30]'
    next_states = []
    possible_final_state_indicator = 0
    for i in disass:
        string = f'{i.mnemonic} {i.op_str}'
        if delim in string:
            next_state = string[string.find(', ')+2:]
            if next_state.startswith('0x'):
                next_states.append(int(next_state[2:], 16))
            else:
                next_states.append(int(next_state))

            if possible_final_state_indicator == 0 and 'mov rax, qword ptr [rsp + 0x58ab8]' in string:
                possible_final_state_indicator = 1
            elif possible_final_state_indicator == 1 and 'inc rax' in string:
                possible_final_state_indicator = 2
            elif possible_final_state_indicator == 2 and 'mov qword ptr [rsp + 0x58ab8], rax' in string:
                possible_final_state_indicator = 3
    # for i in disass:
    #     if i.mnemonic == 'jmp':
    #         start_registering = True
    #     string = f'{i.mnemonic} {i.op_str}'
    #     if ', 0x' not in string:
    #         continue
    #     hex_val = string[string.find(', 0x')+2:]
    #     if start_registering and hex_val.startswith('0x'):
    #         next_states.append(int(hex_val[2:], 16))
    if len(next_states) == 0 and possible_final_state_indicator == 3:
        print('FOUND FINAL STATE')
        next_states.append(0xFFFFFFFF)  # Final state
    return next_states

def get_next_states_from_addr(addr):
    offset = pe.rva_to_offset(addr - pe.imagebase)
    md = capstone.Cs(capstone.CS_ARCH_X86, capstone.CS_MODE_64)
    n_rdtsc = 0
    should_register = False
    disass = []
    for i in md.disasm(code[offset:offset+1000], addr):
        if i.mnemonic == 'rdtsc':
            n_rdtsc += 1
        if n_rdtsc >= 3:
            break
        if f'{i.mnemonic} {i.op_str}' == 'movsx eax, byte ptr [rsp + 0x30]' or \
           f'{i.mnemonic} {i.op_str}' == 'movzx eax, byte ptr [rsp + 0x30]':
            should_register = True
        if should_register:
            disass.append(i)
    return get_next_states_from_disass(disass)

def disass_at_va(addr):
    offset = pe.rva_to_offset(addr - pe.imagebase)
    md = capstone.Cs(capstone.CS_ARCH_X86, capstone.CS_MODE_64)
    n_rdtsc = 0
    should_print = False
    for i in md.disasm(code[offset:offset+1000], addr):
        if i.mnemonic == 'rdtsc':
            n_rdtsc += 1
        if n_rdtsc >= 3:
            break
        if f'{i.mnemonic} {i.op_str}' == 'movsx eax, byte ptr [rsp + 0x30]' or \
           f'{i.mnemonic} {i.op_str}' == 'movzx eax, byte ptr [rsp + 0x30]':
            should_print = True
        if should_print:
            print_hex_colored(f"0x{i.address:x}: {i.mnemonic} {i.op_str}")

def print_state_code(state):
    addr = get_uint_at_va(0x140c687b8 + state * 4)
    code_addr = 0x140000000 + addr
    print(f'State {state}: {hex(addr)} -> {hex(code_addr)}\n')
    disass_at_va(code_addr)

def get_next_states(state):
    addr = get_uint_at_va(0x140c687b8 + state * 4)
    code_addr = 0x140000000 + addr
    return get_next_states_from_addr(code_addr)

def solve(state, depth, path=[]):
    # TODO: Fix this, prob wrong
    if state == 0xFFFFFFFF:
        print(f'Found path with state -1: {[hex(s) for s in path]}')
        exit()
    if depth == 16:
        print(f'Found path with depth {depth}: {[hex(s) for s in path]}')
        return
    next_states = get_next_states(state)
    # print(f'Depth: {depth:2} State {hex(state)} -> Next states: {[hex(s) for s in next_states]}')
    for s in next_states:
        solve(s, depth+1, path + [s])

print_state_code(state)
# solve(state, 0)
