import sys

if len(sys.argv) != 2:
    print('Usage {0} file'.format(sys.argv[0]));
    exit()

filename = sys.argv[1]

file = open(filename)

n = 8
for line in file.readlines():
    line = line.strip()
    chunks = [line[i:i+n] for i in range(0, len(line), n)]
    newLine = ''
    for chunk in chunks:
        newLine += chunk[::-1]
    print(newLine)

