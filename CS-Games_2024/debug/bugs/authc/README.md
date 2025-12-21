# authc `lang:C/C++` `points:20`

## FR 

authc implémente un mécanisme d'authentification par nom d'utilisateur/mot de passe avec prise en charge des mots de passe à usage unique basés sur le temps (TOTP) en tant que deuxième facteur d'authentification pour les utilisateurs administratifs.

## EN

authc implements a username/password authentication mechanism with support for time-based one-time passwords (TOTP) as a second authentication factor for administrative users.


## Example Input/Output

Input:
```
login
guest
wrongpassword
```

Output:
```
Enter username:
Enter password:
ERR: invalid username or password
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
<td><code>bin/build

g++ src/main.cpp -o authc</code></td>
</tr>

<tr>
<td><code>ENV=local bin/run tests/test_sample.in</code></td>
<td><code>./authc < tests/test_sample.in</code></td>
</tr>

<tr>
<td><code>bin/shell</code></td>
<td>Start a shell in an ephemeral container</td>
</tr>

</table>

