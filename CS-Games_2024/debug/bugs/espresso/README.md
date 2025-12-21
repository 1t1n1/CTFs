# espresso `lang:Java` `points:30`

# FR

Vous êtes bien occupé à corriger des bugs? Nous allons commander des boissons avec l'application Espresso. Une fois votre commande envoyée, l'application vous indiquera le coût total de votre commande. Choisissez ce que vous voulez, et nous nous occuperons du reste. Voici le menu:

- Boissons:
  - `coffee`: 1,00 $
  - `tea`: 0,90 $
- Tailles:
  - `small`: pas de coût supplémentaire
  - `medium`: +1,00 $
  - `large`: +2,00 $
- Suppléments:
  - `cream`: +0,30 $
  - `milk`: +0,20 $
  - `sugar`: +0,10 $
  - `lemon`: +0,10 $

# EN

Busy fixing bugs? We're making drink orders using the Espresso app. Once your order is sent, the app will tell you the total cost of your order. Pick anything you want, and we'll take care of the rest. Here is the menu:

- Drinks:
  - `coffee`: $1.00
  - `tea`: $0.90
- Sizes:
  - `small`: no additional cost
  - `medium`: +$1.00
  - `large`: +$2.00
- Addons:
  - `cream`: +$0.30
  - `milk`: +$0.20
  - `sugar`: +$0.10
  - `lemon`: +$0.10


## Example Input/Output

Input:
```
Order for: Bob
- 1 large coffee

Order for: Alice
- 2 medium tea
```

Output:
```
Order for: Bob
Total: $3.00

Order for: Alice
Total: $3.80
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

javac src/*.java</code></td>
</tr>

<tr>
<td><code>ENV=local bin/run tests/test_sample.in</code></td>
<td><code>java src.Main tests/test_sample.in</code></td>
</tr>

<tr>
<td><code>bin/shell</code></td>
<td>Start a shell in an ephemeral container</td>
</tr>

</table>

