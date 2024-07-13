Quand on envoie une query et que la console js est ouverte, on voit un avertissement à propos de l'encodage windows-1252. Ceci est indiqué puisque les informations envoyés seront réencodés en ISO-8859-1. La particularité intéressante entre ces deux encodages est que le premier contient les mêmes caractères que le second, mais il en contient aussi d'autres, notamment des quotes spéciales (ce qui nous intéresse grandement!). Je dois investiguer d'avantage là dessus...

![[Windows-1252.png]]
![[ISO-8859-1.png]]


Possible command injection with param('username') in the if condition.

Conclusion:
Il fallait exploiter la fonction 
```perl
DBI->quote($value, $data_type);
```
qui permettait un second paramètre optionel. Celui-ci permettait d'éviter de quote si le data_type ne le nécessite pas (comme pour une valeur numérique, par exemple). Ainsi, lorsqu'on passait un tableau dans la requête POST dans le champ username ou password, il était possible d'éviter le quote de notre $value et ainsi faire une injection

Ce que j'ai bien fait/dois améliorer:
J'ai très bien fait de ne pas lâcher le morceau. J'étais très près de la solution et j'avais même déjà trouvé et analysé les pages web que la page contenant la solution référait. Mon problème était avec mon interprétation de comment est-ce que les valeurs que l'on passe aux champs username et passwords sont ensuite imbriquées dans la requête SQL. Je croyais que mon _input_ serait _quoté_, mais ce n'était pas le cas. Aussi, pour le $data_type, je comprenais ce que je devais envoyer, mais la documentation était mauvaise alors je ne l'ai pas trouvé. J'ai dû me rabattre à la solution.