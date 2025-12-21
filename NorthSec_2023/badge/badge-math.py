from pwn import *

tube = tubes.serialtube.serialtube(port='/dev/ttyUSB0')
# print(tube.recvuntil(b'\x1b[0;32mnsec> \x1b[0m'))
print('sending math command')
tube.sendline('math'.encode())
print(f'tube.recvline() = {tube.recvline()}')
question = tube.recvline()
# print(question)

tube.sendline('math'.encode())
raw_question = tube.recvline()
index = raw_question.find(b'=')
if index == -1:
    raw_question = tube.recvuntil(b"= ")
    index = raw_question.find(b'=')
question = raw_question[:index].decode()
print(f'question = {question}')
print(f'tube.recvline() = {tube.recvline()}')
answer = str(eval(question)).encode()
raw_possible_question = tube.recvuntil(b"= ")
possible_question = raw_possible_question[:raw_possible_question.find(b'=')].decode()
possible_answer = str(eval(possible_question)).encode()
if possible_answer != answer:
    answer = possible_answer
print(f'answer = {answer}')
tube.sendline(answer)
print(f'tube.recvline() = {tube.recvline()}')

## From here, everything should be fine

new_raw_question = tube.recvuntil(b"= ")
new_question = new_raw_question[:new_raw_question.find(b"=")].decode()
print(f'new_question = {new_question}')
tube.sendline(str(eval(new_question)).encode())
print(f'tube.recvline() = {tube.recvline()}')


new_raw_question = tube.recvuntil(b"= ")
new_question = new_raw_question[:new_raw_question.find(b"=")].decode()
print(f'new_question = {new_question}')
tube.sendline(str(eval(new_question)).encode())
print(f'tube.recvline() = {tube.recvline()}')


while True:
    new_raw_question = tube.recvuntil([b'= ', b'\n'])
    if new_raw_question.find(b'=') == -1:
        print(f'new_raw_question = {new_raw_question}')
        break
    # new_raw_question = tube.recvuntil(b"= ")
    new_question = new_raw_question[:new_raw_question.find(b"=")].decode()
    print(f'new_question = {new_question}')
    new_answer = str(eval(new_question)).encode()
    print(f'new_answer = {new_answer}')
    tube.sendline(new_answer)
    print(f'tube.recvline() = {tube.recvline()}')

while True:
    print(tube.recvline())