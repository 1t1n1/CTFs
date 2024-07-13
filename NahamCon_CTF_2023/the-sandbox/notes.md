Had to install qemu-aarch64-static

Steps to setup the environment:

```sh
$ sudo apt install qemu-user-static libc6-arm64-cross

$ sudo ln -s /usr/aarch64-linux-gnu/lib/ld-linux-aarch64.so.1 /lib/ld-linux-aarch64.so.1

$ sudo ln -s /usr/aarch64-linux-gnu/lib/libc.so.6 /lib/libc.so.6

$ sudo update-binfmts --install aarch64 /usr/bin/qemu-aarch64-static --magic '\x7fELF\x02\x01\x01\x00\x00\x00\x00\x00\x00\x00\x00\x00\x02\x00\xb7\x00' --mask '\xff\xff\xff\xff\xff\xff\xff\x00\xff\xff\xff\xff\xff\xff\xff\xff\xfe\xff\xff\xff'

$ chmod +x <path-to-aarch64-binary>
```

You can now execute the arm64 binary under x86 host (should work for other infrastructures too)
```sh
$ qemu-aarch64-static <path-to-aarch64-binary>
```

qemu is used to emulate a whole operating system, but qemu-user is used to emulate a simple program on a host of a different architecture than the program's target 
architecture. `static` simply means that it is used with staticaly linked binaries (verify these informations before publishing)

libc6-arm64-cross downloads arm64 libraries under /usr/aarch64-linux-gnu/lib/

The second and third commands link /lib/<the library> to where it is actually located because the binary will usually look under /lib/ for the libraries they need.

The fourth command updates qemu database to add arm64 architecture.

The last commands simply execute the program.

However, you can also use gdb with multiple architectures!

Add this setup command:
```sh
$ sudo apt install gdb-multiarch
```

And use these commands instead to execute the binary:
```sh
# On a terminal
$ qemu-aarch64-static -g <port> <path-to-aarch64-binary>

# On another terminal
(gdb) gef-remote --qemu-user --qemu-binary <path-to-aarch64-binary> localhost <port>        # See https://hugsy.github.io/gef/commands/gef-remote/ and `help gef-remote` in gef
```

Note : REMOVE /lib/libc.so.6 WHEN I'M DONE! I don't want problems





To break at the main function: `b *0x55000012C0` (it's image base + 0x12C0)