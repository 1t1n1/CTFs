/root/anonymized.phony
/root/flag.txt

installed: `openssh-server build-essential ltrace vim iputils-ping gdb`
gdb installé plus dernièrement

/etc/ssh/sshd_config: `UsePrivilegeSeperation no`
So perpetrators probably had root access

Last file modified is flag...
``` bash
$ find ../../ -type f -mtime -26 2>/dev/null 
../../sshd_root_dir/root/flag.txt
```

And we can se that sshd crashed just two days before, as said in the description!
``` bash
$ find ../../ -type f -mtime -28 2>/dev/null
../../sshd_root_dir/var/lib/apt/extended_states
../../sshd_root_dir/var/lib/systemd/coredump/sshd.core.93794.0.0.11.1725917676
../../sshd_root_dir/var/lib/dpkg/triggers/Lock
../../sshd_root_dir/var/lib/dpkg/triggers/Unincorp
[...]
../../sshd_root_dir/usr/share/mime/x-epoc/x-sisx-app.xml
../../sshd_root_dir/usr/share/mime/version
../../sshd_root_dir/usr/lib/x86_64-linux-gnu/liblzma.so.5.4.1
```

Cleaner shot:
``` bash
$ find ../../ -type f -mtime -28 -exec stat --format="%y %n" {} \; 2>/dev/null | sort
[...]
2024-09-09 17:21:59.000000000 -0400 ../../sshd_root_dir/var/log/apt/history.log  
2024-09-09 17:21:59.000000000 -0400 ../../sshd_root_dir/var/log/apt/term.log  
2024-09-09 17:21:59.000000000 -0400 ../../sshd_root_dir/var/log/dpkg.log  
2024-09-09 17:34:33.000000000 -0400 ../../sshd_root_dir/usr/lib/x86_64-linux-gnu/liblzma.so.5.4.1  
2024-09-09 17:34:36.000000000 -0400 ../../sshd_root_dir/var/lib/systemd/coredump/sshd.core.93794.0.0.11.1725917676  
2024-09-11 16:55:59.000000000 -0400 ../../sshd_root_dir/root/flag.txt
```
We can see that the server did stuff and then crashed sshd 13 minutes after the last file modification. Then flag file was written two days later.

Core dump shows that sshd crashed after calling liblzma (xz vulnerability?)
vulnerable xz is `5.6.0`. Installed xz is `5.4.0`... (oups)

Actually, maybe the attacker tried the xz vuln and failed, which resulted in the process crashing

Interesting strings in crash dump:
- `SSH_CONNECTION=172.17.0.1 33894 172.17.0.2 22`
- `publickey test pkalg rsa-sha2-512-cert-v01@openssh.com pkblob RSA-CERT SHA256:blGX19E9esdXjw9Qm0l7Gj8nCj7y9ZoM7nmy7n/4Ask CA RSA SHA256:bK+V4IHxEiEppo0j7oSD+TOIPf2HkY1dxy0L7H6qtAo`
- `YAAADAQABAAABgQCpGiLjoWCVr+cYuUrdboUJAKTwqx3yC3VE5DFH++aFtMUSoapZysBkJN90tniliwVqNHjIZ+NupB6gzZ+9LJeNI/OA2kAduCbgIBiP3aj7Ckh6jkKsf/gEsEvOt3oRkFmvs2T8Rvr40oedk0W6TPHiES1tJgzXEaJb+ClrO/9el+UCd2OQ0/IVAGpvwrY6/WBEW23PUqDqLpnxbFlWMdcYealH3R0cV3nvwxxz/9CgSqdzJWRfFHW6ZVV3npS2fkqd0e43cMGFVilLW6BuVUtjNQWf9KelusQX+VvgDn1j53I3mvgV2LmVBzNuhPQGiOE19b8EA5z3dfTuGwfbRRq12cgeUY7DNJHUYrwBi6pcYv1LV8y+XfeOsuzv5xlknLe2udtn5tIQpaPPhP9rfckF8ocVJ44PXj4FMSdlbhPZPiqSOm4CMHuVw/KDYP8Y/F0ROblClRIr4nuV14ZhkXVwq441YXdN/BAXqsCg+YvW88fZbdKMhPjgYPcg90MbQO8=`
- `oWCVr+cYuUrdboUJAKTwqx3yC3VE5DFH++aFtMUSoapZysBkJN90tniliwVqNHjIZ+NupB6gzZ+9LJeNI/OA2kAduCbgIBiP3aj7Ckh6jkKsf/gEsEvOt3oRkFmvs2T8Rvr40oedk0W6TPHiES1tJgzXEaJb+ClrO/9el+UCd2OQ0/IVAGpvwrY6/WBEW23PUqDqLpnxbFlWMdcYealH3R0cV3nvwxxz/9CgSqdzJWRfFHW6ZVV3npS2fkqd0e43cMGFVilLW6BuVUtjNQWf9KelusQX+VvgDn1j53I3mvgV2LmVBzNuhPQGiOE19b8EA5z3dfTuGwfbRRq12cgeUY7DNJHUYrwBi6pcYv1LV8y+XfeOsuzv5xlknLe2udtn5tIQpaPPhP9rfckF8ocVJ44PXj4FMSdlbhPZPiqSOm4CMHuVw/KDYP8Y/F0ROblClRIr4nuV14ZhkXVwq441YXdN/BAXqsCg+YvW88fZbdKMhPjgYPcg90MbQO8= root@e303d6e4c6d7`
- `HOSTNAME=e303d6e4c6d7`
- `PWD=/fmnt`

In the last function before crash (in `/lib/x86_64-linux-gnu/liblzma.so.5`), there is a string: `RSA_public_decrypt`

Seems like it tried to `call 0x0` (`call rax` and last registered `RIP` is `0x0`)

Ctrl+f for `RSA_public_decrypt` at `https://arstechnica.com/security/2024/04/what-we-know-about-the-xz-utils-backdoor-that-almost-infected-the-world/` really makes me think that the attacker tried the xz backdoor but crashed the process entirely. **What is certain now is that it has to do with xz's backdoor.**

Problem is, the description states "Now criminals are trying to sell me my own data!!!", so the vuln must have worked...

Mayyyybe run with valgrind, because that's what the microsoft dev used to find the vuln (it caused errors)

> The backdoor is composed of many parts introduced over multiple commits:
> 	- Using IFUNCs in the build process, which will be used to hijack the symbol resolve functions by the malware

Maybe the flag is in the command that the attacker intended to execute (must find vulnerable version of sshd to decrypt command)
$rsi: points to seemingly encrypted data
$rdi: size of seemingly encrypted data

En même temps y'a juste l'auteur de la vuln qui peut encrypter des commandes avec sa clé privée...

call RAX shouldn't have crashed... rax should be correctly set...

Possible que la clé ssh publique est simplement loadé dans ssh en mémoire vive pendant son exécution en loadant les clé dans .ssh

Pour continuer:
- Fichier dans `~/tmp/sshd`
- `gdb -c sshd.core.93794.0.0.11.1725917676 sshd` puis `bt` ou `i r` et `telescope $rsi -l 64` pour voir le payload (de taille `0x200`)
- hxxps://arstechnica.com/security/2024/04/what-we-know-about-the-xz-utils-backdoor-that-almost-infected-the-world/
- hxxps://www.akamai.com/blog/security-research/critical-linux-backdoor-xz-utils-discovered-what-to-know
- hxxps://cs4157.github.io/www/2024-1/lect/21-xz-utils.pdf
- hxxps://medium.com/@knownsec404team/analysis-of-the-xz-utils-backdoor-code-d2d5316ac43f
- Pistes possibles:
	- Rouler le sshd vulnérable et m'arrange pour quil exécute ce qu'il était après faire, mais sans crasher pour récupérer ensuite la commande déchiffrée (je crois que c'est possible que la commande contienne le flag), mais en même temps c'est impossible qu'il puisse avoir encodé son propre message vu que c'est juste l'auteur de la backdoor qui peut le faire...
	- Ça aurait pas dû crasher... Même avec un sshd pas vulnérable ca aurait pas crasher, ça aurait amené à la bonne version de `RSA_public_decrypt`...
	- Relire notes
