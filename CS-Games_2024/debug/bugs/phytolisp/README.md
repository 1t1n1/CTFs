# phytolisp `lang:PHP` `points:40`

# FR

Phytolisp est un interpréteur d'un dialecte Lisp écrit en PHP.

# EN

Phytolisp is a Lisp dialect interpreter written in PHP.


## Example Input/Output

Input:
```

```

Output:
```
helloworld
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
<td><code>php src/index.php tests/test_sample.in</code></td>
</tr>

<tr>
<td><code>bin/shell</code></td>
<td>Start a shell in an ephemeral container</td>
</tr>

</table>

