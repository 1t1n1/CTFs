import requests
import json

#données à envoyer
data = {
    "path": [(0, 0), (0, 1), (0, 2), (1, 2), (2, 2), (2, 1), (2, 0), (1, 0), (1, 1)]
}

# headers pour spécifier le type de contenu
headers = {
    "Content-Type": "application/json"
}

# envoyer la requête POST à l'endpoint
response = requests.post("http://challenges.ulctf.ca:30202/maze ", data=json.dumps(data), headers=headers)
print("Code d'état de la réponse : ", response.status_code)

# imprimer la réponse
print(response.text)
