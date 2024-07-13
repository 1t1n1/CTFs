import requests
import json

URL = "http://challenges.ulctf.ca:30202/maze"

try:
    response = requests.get(URL, timeout=10)
    if response.status_code == 200:
        print("hello from the server \U0001F600 \n")
        maze = response.text
        mazeobj = json.loads(maze)
        mazeuni = mazeobj["maze"]
        print(mazeuni)
        print(f"Pour avoir la liste rendez-vous ici -->{URL}")

    else:
        print("Failed to get response from server.")
except requests.exceptions.RequestException as e:
    print("Exception occurred:", e)
    print("Please check your network/firewall settings.")

