![[Pasted image 20221116213656.png]]
![[Pasted image 20221116213850.png]]
Webserver on an odd port is a good thing to look at
![[Pasted image 20221116215049.png]]
![[Pasted image 20221116215101.png]]
![[Pasted image 20221116215126.png]]
![[Pasted image 20221116215151.png]]
![[Pasted image 20221116215412.png]]
![[Pasted image 20221116215553.png]]
dwarf2json to build a profile for volatility using the link in the precedent question
![[Pasted image 20221116215751.png]]
![[Pasted image 20221116215801.png]]
PID is 1413 because we knew that python3 was the malware 
![[Pasted image 20221116215919.png]]
![[Pasted image 20221116215951.png]]
You needed to dig a bit (understand Empire's code and packet structure)
There was also a blog post that you could have could have found that detailed empire packet structure:
![[Pasted image 20221116220136.png]]
You take the bytes underlined in the precedent picture
Skip the 0d0a (newline) and what precedes it as this is just the HTTP headers and separate the packet using the known packet structure
![[Pasted image 20221116220551.png]]
![[Pasted image 20221116220559.png]]
![[Pasted image 20221116220627.png]]
Malware usually use staging key to obfuscate communications, but at this point it probably used shared secret
![[Pasted image 20221116220833.png]]
We were supposed to bruteforce the keyspace to find the key (because we knew that it was about AES256, CBC mode and 32 byte session keys). There were clever ways. 
Je crois que le PID pouvait être utilisé pour nous aider à savoir si on avait la bonne clé
![[Pasted image 20221116221307.png]]

Dans un autre challenge, il a utilisé scapy pour jouer avec les packets en python (et dpdk)