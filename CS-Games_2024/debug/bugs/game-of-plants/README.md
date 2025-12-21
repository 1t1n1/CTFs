# game-of-plants `lang:JavaScript` `points:20`

## FR
Game-of-plants est utilisé pour modéliser la croissance des plantes en utilisant les règles du [Jeu de la vie de Conway](https://fr.wikipedia.org/wiki/Jeu_de_la_vie) après un nombre donné d'itérations. 

## EN
Game-of-plants is used to model the plant growth using the [Conway's Game of Life](https://en.wikipedia.org/wiki/Conway%27s_Game_of_Life) rules after a given number of iterations.


## Example Input/Output

Input:
```
4
........
........
..111...
....1...
........
```

Output:
```
........
...11...
..1..1..
...11...
........
```

## Scripts

<table>
<tr>
<th>Script</th>
<th>Description</th>
</tr>

<tr>
<td><code>bin/build</code></td>
<td>Build the project's container</td>
</tr>

<tr>
<td><code>bin/test</code></td>
<td>Run all test cases</td>
</tr>

<tr>
<td><code>bin/test test_sample</code></td>
<td>Run a single test case</td>
</tr>

<tr>
<td><code>ENV=local bin/build</code></td>
<td><code># N/A</code></td>
</tr>

<tr>
<td><code>ENV=local bin/run tests/test_sample.in</code></td>
<td><code>node src/main.js tests/test_sample.in</code></td>
</tr>

<tr>
<td><code>bin/shell</code></td>
<td>Start a shell in an ephemeral container</td>
</tr>

</table>

