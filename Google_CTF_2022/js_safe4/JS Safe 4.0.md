Voir la partie \<script\>

Voir la partie avec les quatre TODOs

Je crois que x est la fonction qui calcule le mot de passe. Je doit trouver son implémentation. Je crois qu'elle se trouve en haut de l'endroit où elle est invoquée

Function (la création d'une fonction à partir de son constructeur) semble avoir [des problèmes de sécurité](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Function/Function)

Voir pour un autre problème de sécurité [ici](https://developer.mozilla.org/en-US/docs/Web/API/setTimeout)

TODO : Enlever Chromium à la fin du CTF

Je conclu que je dois passer le checksum et le pool à la fin

Ce mot de passe permet de traverser la première vérification (envoyer dans vscode avant pour enlever les newline): CTF{20p00p00q0uF0o_00p00p3iC00p0n_00p1mX1hi00p00p1sC1pC1tD1ue1uD0wD1re1we0oi1jX00p00p1hX1tD1hR1nX1ie0mi00p0ue0tD1hi0te0li1kX0ve1sC00p0l_0te00p1re0mX0li1k_00p0mX1se1tC00p0vD1oR1nX0tD1jR1nX1re1mi0mX1ue1vF0li1ve0vD0li1pD0te1mi0te0vF0li1pD0l_1hi1tF0tD1ai1ve0oX1aX1ve1pC1jX0wF1ni1aX0uF0mR0mX1iF1te1oX0uF0mi1kX0tF1wC1ai0l_0tD00p1gG1ki0nX1ue1pD1jX1iD0ni00p1re0nX0oX0wD0nR1h_1wC00p1gG00p1k_0oR1mX1sD1oi1vC0li00p1iC00p1hi0mX00p1tF0uF0l_0nX00p1nX0li1vC00p00p00p00p1tF1pD1se00p0mX00p00p1aX00p00p1pe00p0wD00p0oi1h_1tF0uF0l_00p1kX1aX0uD1rF0ni00p1jX00p0mR00p0ue1jX}

Ceci devrait bypasser la deuxième vérification, mais ne fonctionne pas: 95cw_UbW3H_slc31Lz0_uN1__Kd

C'est juste la partie avant la vérification avec le checksum qui doit être inchangée