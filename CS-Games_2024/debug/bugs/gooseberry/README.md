# gooseberry `lang:Go` `points:30`

## Description 
Gooseberry is a computation software to help AI agents perform mathematical operations on a nested data structure.

The data structure is represented in JSON using two fields:
- `type`: one of `string`, `int`, or `list`
- `value`: the value of the data structure

Once Gooseberry received the data structure, it expects the following messages to contain the operation to perform on the data structure.

An operations is defined as a list where the first element is the operation name and the following elements are its arguments. For example, to add 42 to all integers in the data structure, the `add` operation data structure would look like the following:

```json
{
  "type": "list",
  "value": [
    { "type": "string", "value": "add" },
    { "type": "int", "value": 42 }
  ]
}
```


## Description
Gooseberry est un logiciel pour aider les agents d'IA à effectuer des opérations mathématiques sur une structure de données imbriquée. 

La structure de données est représentée en JSON à l'aide de deux champs :

- `type` : dont la valeur est l'une des suivantes : `string`, `int` ou `list`
- `value` : la valeur de la structure de données

Une fois que Gooseberry a reçu la structure de données, les messages suivants doivent contenir l'opération à effectuer sur la structure de données. 

Une opération est définie comme une liste où le premier élément est le nom de l'opération et les éléments suivants sont ses arguments. Par exemple, pour ajouter 42 à tous les entiers de la structure de données, la structure de données de l'opération `add` ressemblerait à ceci :

```json
{
  "type": "list",
  "value": [
    { "type": "string", "value": "add" },
    { "type": "int", "value": 42 }
  ]
}
```


## Example Input/Output

Input:
```
{
  "type": "list",
  "value": [
    {
      "type": "int",
      "value": 20
    },
    {
      "type": "list",
      "value": [
        {
          "type": "int",
          "value": 50
        }
      ]
    }
  ]
}

{
  "type": "list",
  "value": [
    {
      "type": "string",
      "value": "add"
    },
    {
      "type": "int",
      "value": 10
    }
  ]
}
```

Output:
```
{
  "type": "list",
  "value": [
    {
      "type": "int",
      "value": 30
    },
    {
      "type": "list",
      "value": [
        {
          "type": "int",
          "value": 60
        }
      ]
    }
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
<td><code>bin/build

go build -o gooseberry src/*.go</code></td>
</tr>

<tr>
<td><code>ENV=local bin/run tests/test_sample.in</code></td>
<td><code>./gooseberry tests/test_sample.in</code></td>
</tr>

<tr>
<td><code>bin/shell</code></td>
<td>Start a shell in an ephemeral container</td>
</tr>

</table>

