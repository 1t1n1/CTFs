for i in range(128):
    x = i
    t = 0
    while True:
        # print(x)
        t = t * 2 + x % 2
        x = int(x / 2)
        if x == 0:
            break
    if t == 67:
        print(i)
        break
    