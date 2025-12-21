# serplante `lang:Python` `points:20`

## FR 

Serplante est utilisé pour analyser une carte 2D afin de détecter la présence d'un serpent génétiquement modifié.

Les serpents sont formés par une chaîne de caractères `#` où la tête et la queue peuvent être identifiées sans ambiguïté. Les serpents n'aiment pas former de boucles, se chevaucher eux-mêmes ou toucher d'autres serpents. Ils peuvent seulement se déplacer vers le haut, le bas, la gauche et la droite.

## EN 

Serplante is used to analyze a 2D map to detect the presence of a genetically modified snake.

Snakes are formed by a chain of `#` characters where the head and tail can be identified without ambiguity. Snakes don't like forming loops, overlap each other or touch other snakes. Their movements is limited to up, down, left and right.



## Example Input/Output

Input:
```

 ###########  <= snake
 #         ##

     ##
#### #######
#  #
#######   ##
          ##
 ####
```

Output:
```
groups: 5
snakes: 1
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
<td><code>python3 src/serplante.py tests/test_sample.in</code></td>
</tr>

<tr>
<td><code>bin/shell</code></td>
<td>Start a shell in an ephemeral container</td>
</tr>

</table>

