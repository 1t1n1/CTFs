# idleaf `lang:Rego` `points:40`

## FR



## EN

idleaf is an experimental Role-Based Access Control (RBAC) engine for the key-value database `fsdb`. Each key in the database can be assigned a set of users or teams, each with one of the following roles:
- `admin`: full control
- `write`: read/write access
- `read`: read only 

Users can be assigned to multiple teams, and teams can contain other teams.



## Example Input/Output

Input:
```
{
    "keys": {
      "foo": {
        "john": "read"
      }
    },
    "key": "foo",
    "action": "write",
    "user": "john"
}
```

Output:
```
{
  "allow": false,
  "errors": [
    "access_denied: `john` is not allowed to perform action `write` on key `foo`"
  ]
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
<td><code>opa eval -b src data.idleaf.result -f pretty -i tests/test_sample.in</code></td>
</tr>

<tr>
<td><code>bin/shell</code></td>
<td>Start a shell in an ephemeral container</td>
</tr>

</table>

