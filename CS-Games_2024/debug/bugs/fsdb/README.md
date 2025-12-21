# fsdb `lang:Bash` `points:20`

FR: fsdb est une base de données clé-valeur très performante écrite en bash. Elle utilise un protocole texte simple pour gérer la base de données.

EN: fsdb is highly performant key-value database written in bash. It uses a simple text protocol to manage the database. 


## Example Input/Output

Input:
```
set parent/key bar
get parent/key
```

Output:
```
1: OK
2: 3
bar
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
<td><code>bash src/fsdb.sh tests/test_sample.in</code></td>
</tr>

<tr>
<td><code>bin/shell</code></td>
<td>Start a shell in an ephemeral container</td>
</tr>

</table>

