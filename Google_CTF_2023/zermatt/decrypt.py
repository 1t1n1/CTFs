#!/usr/bin/env python
str1, str2 = '168 122","134 84 208 67'.split('","')
str1 = str1.split(' ')
str2 = str2.split(' ')
result = ''
for i in range(len(str1)):
    result += chr(int(str1[i]) ^ int(str2[i % len(str2)]))
print(result)