# ircman

Even though this old binary seems harmless, it seems that it contains some crucial data in it. 

Flag: HTB{1RCb0t5_4r3_fun!} 

Solution: 
1. Find the password (the letters are ROT13'ed)
2. Start a server at the port (`nv -lvnp 8000`)
3. Start the client (`./ircman`)
4. Send `PRIVMSG #room :@pass M1N1M4L!`
5. Send `PRIVMSG #room :@flag`
