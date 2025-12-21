# instructions

[_English_](#english) 

## Français

<img align="left" width="100" height="100" src="https://2024.csgames.org/logo.png">

Bienvenue à la compétition de debugging des CS Games 2024!

Chlorophyllai a corrompu la dernière copie du code le plus important nécessaire à la survie de notre société. Malheureusement, les dommages sont si importants que le code n'est plus fonctionnel. Celui-ci a été modifié pour introduire un nombre important de bogues. Parfois, les projects ont été complètement réécrits de manière non conventionnelle. Heureusement pour nous, il nous reste encore de nombreux cas de test qui peuvent vérifier la justesse du fonctionnement de chaque projet. Votre objectif est de restaurer les projets à un état fonctionnel en appliquant des correctifs à chaque projet pour qu'ils passent tous les cas de test.

Compte tenu de la nature critique de la situation, vous allez compétitionner contre d'autres équipes pour **corriger le plus de bogues possible, et ce, le plus rapidement possible**.

Dans le projet Gitness de votre équipe, vous aurez bientôt accès au dépôt `bugs`. Il contient une copie du code sur lequel vous devrez travailler et pousser vos correctifs.

**Toutes les 10 minutes**, tous les projets de votre dépôt seront testés pour évaluer leur viabilité. Pour chaque projet, votre équipe perdra des points en fonction du ratio de cas de test échoués multiplié par la valeur en points du projet. Par exemple, un projet valant 20 points qui passe 25% des cas de test entraînera une pénalité de 15 points pour la période d'évaluation. L'équipe avec le plus de points à la fin de la compétition gagne.

## Règles

- Vous pouvez utiliser des branches et des "pull requests" pour gérer votre travail, mais seule la branche `main` est évaluée. Assurez-vous que vos modifications sont fusionnées dans `main` avant la prochaine évaluation.
- L'accès à internet et aux outils IA/LLM est autorisé.
- Les participants au sein d'une équipe peuvent collaborer, mais les équipes ne peuvent pas collaborer entre elles ni recevoir d'aide externe.
- Bien que ce ne soit pas le but de la compétition, vous pouvez réécrire l'ensemble du projet tant que le langage original du projet est utilisé.
- Ne soumettez pas de code malveillant, de fichiers volumineux ou de modification qui pourrait affecter la disponibilité ou l'équité de la compétition.
  - Ne tentez pas d'exécuter des commandes externes, de "forkbomb", de boucles infinies, d'appels réseau externes, scans de ports, dénis de service, etc.

### Structure du dépôt

Le dépôt `bugs` contient plusieurs dossiers pour chaque projet. Dans chaque dossier, vous trouverez les fichiers et répertoires suivants pour vous aider à construire et tester le projet :

### `src/`

Ce dossier contient le code source du projet. **C'est le seul dossier qui est évalué. Les modifications apportées aux autres fichiers du dépôt ne sont pas prises en compte.**

### `tests/`

Ce dossier contient des cas de test publics pour le projet. Chaque cas de test est une paire de fichiers : `test_name.in` et `test_name.res`. Le fichier `.in` contient l'entrée du programme et le fichier `.res` contient la sortie attendue.

### `bin/build`

Si le projet nécessite d'être compilé, ce script sera utilisé pour construire le projet. La commande de construction sous-jacente est fournie dans le README du projet.

### `bin/run` 

Ce script exécute le projet sur un fichier spécifique dans le dossier `tests`. Si le projet nécessite d'être compilé, `bin/build` doit être appelé avant.

```sh
bin/run tests/test_sample.in
```

### `bin/test`

- `bin/test` construira le projet et exécutera tous les tests.
- `bin/test test_name` construira le projet et montrera la différence entre la sortie de `bin/run tests/test_name.in` et la sortie attendue dans le fichier `tests/test_name.res`.

### `ENV=local`

Par défaut, les scripts dans `bin/` exécuteront des commandes `docker`. Si vous avez les dépendances nécessaires d'un projet installées sur votre machine, vous pouvez exécuter les scripts avec `ENV=local` pour exécuter les commandes du projet localement.

```sh
ENV=local bin/build
ENV=local bin/run
ENV=local bin/test

# ou
export ENV=local
bin/build
bin/run
bin/test
```

## Cloner le dépôt

Dans le bouton "Clone" de n'importe quel dépôt sur Gitness, cliquez sur "Generate Clone Credential". Cela générera un jeton qui peut être utilisé pour cloner les dépôts.

Pour éviter de devoir entrer ce jeton pour chaque operation git, vous pouvez cloner le dépôt en utilisant le jeton comme partie de l'URL :
```sh
git clone https://token:$GITNESS_TOKEN@server.golden-bull.ts.net/git/team-$TEAM/bugs.git
```


Configurez git pour utiliser le jeton pour tous les dépôts sur le serveur :
```sh
git config --global url."https://token:$GITNESS_TOKEN@server.golden-bull.ts.net".insteadOf https://server.golden-bull.ts.net
```

Ou configurez le gestionnaire de credentials de git pour stocker le jeton après la première utilisation :
```sh
git config --global credential.helper store
```

## English

<img align="left" width="100" height="100" src="https://2024.csgames.org/logo.png">

Welcome to the CS Games 2024 Debugging Competition!

Chlorophyllai has corrupted the last copy of the most important codebase necessary for the survival of our society. Unfortunately, the damages are so extensive that the code is no longer functional. The code was modified to introduce a significant number of bugs. Sometimes, completely rewritten in unconventional ways. Lucky for us, we still have numerous test cases that can verify the correctness of each projects. Your goal is to restore the projects to a working state by applying fixes to each project that will make them pass all the test cases.

Given the critical nature of the situation, you will be competing against other teams to **fix as many bugs as possible, as quickly as possible**.

In your team's Gitness project, you will soon be given access to the `bugs` repository. It contains a copy of the codebase that will be working on and push your fixes to.

**Every 10 minutes**, all the projects in your repository will be tested to assess their viability. For every project, your team will lose points based on the ratio of failing test cases multiplied by the project's point value. For instance, a project worth 20 points that passes 25% of the test cases will result in a 15 points penalty for that time period. The team with the most points at the end of the competition wins.

## Rules

- You may use branches and pull requests to manage your work, but only the `main` branch is evaluated. Ensure that your changes are merged into `main` before the next evaluation.
- Access to internet and LLM/AI tools is allowed.
- Participants within a team can collaborate, but teams cannot collaborate with each other nor receive external help.
- While not the competition's goal, you can rewrite the entire project as long as the project's original language is used.
- Don't submit malicious code, large files, or any change that could affect the fairness of the competition.
  - Do not attempt to execute external commands, fork-bombs, infinite loops, external network calls, port scan, denial of service, etc.
  
## Repository structure

The `bugs` repository contains multiple folders for each projects. In each folder, you will find the following files and directories to help you build and test the project. 

### `src/`

This folder contains the source code of the project. **This is the only folder that is evaluated. Changes to other files are not considered.**

### `tests/`

This folder contains public test cases for the project. Each test case is a pair of files: `test_name.in` and `test_name.res`. The `.in` file contains the input to the program and the `.res` file contains the expected output.

### `bin/build`

If the project requires a build step, this script will be used to build the project. The project's README will provide the underlying build command. 

### `bin/run`

This script runs the project on a specific file in the `tests` folder. If the project requires a build step, `bin/build` must be called before.

``` sh
bin/run tests/test_sample.in
```

### `bin/test`

- `bin/test` will build the project and run all tests.
- `bin/test test_name` will build the project and show the difference between the output of `bin/run tests/test_name.in` and the expected output in the file `tests/test_name.res`.

### `ENV=local`

By default, the scripts in `bin/` will run `docker` commands. If you have the necessary dependencies of a project installed on your machine, you can run the scripts using the environment variable `ENV=local` to run the project's commands locally.

``` sh
ENV=local bin/build
ENV=local bin/run
ENV=local bin/test

# or 
export ENV=local
bin/build
bin/run
bin/test
```

## Cloning the repository

In the clone button of any repository on Gitness, click "Generate Clone Credential". This will generate a token that can be used to clone Gitness repositories.

To avoid password prompts when working with the repository, either clone the repository using the token as part of the URL:
``` sh
git clone https://token:$GITNESS_TOKEN@server.golden-bull.ts.net/git/team-$TEAM/bugs.git
```

Configure git to use the token for all repositories on the server:
``` sh
git config --global url."https://token:$GITNESS_TOKEN@server.golden-bull.ts.net".insteadOf https://server.golden-bull.ts.net
```

Or configure git credential helper to store the token after the first use:
``` sh
git config --global credential.helper store
```

