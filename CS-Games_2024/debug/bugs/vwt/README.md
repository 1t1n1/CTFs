# vwt `lang:JavaScript` `points:30`

## FR
Les vegetal web tokens (VWT) sont comme les JSON web tokens (JWT), mais avec quelques différences, ils utilisent un encodage en base 32 sans rembourrage au lieu de l'encodage en base 64.

## EN
Vegetal web tokens (VWT) are like JSON web tokens (JWT), but with some differences, they use base 32 encoding without padding instead of base 64 encoding.


## Example Input/Output

Input:
```
{
  "method": "verify",
  "token": "PMRGC3DHEI5CESCTGI2TMIRMEJ2HS4BCHIRFMV2UEJ6Q.PMRHG5LCEI5CE5LTMVZCE7I.ADJNU2SIE5HGUWKSWBIHXWTN775AR6ZPKHLVRCU6BYZAXUP3ADRA",
  "key": "secret"
}
```

Output:
```
{
  "sub": "user"
}
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

