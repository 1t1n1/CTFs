#!/usr/bin/env python

## Run with python -m flask run or sudo python3 ./server.py (admin rights are
## required for port 80)
from flask import *

app = Flask(__name__)

# http://storage.nsec/?k=Oogum_Boogum 

# http://storage.nsec/?k=Oogum_Boogum&n when requesting the command
@app.route("/")
def decrypt():
    if (request.args.get('n') == '' and request.args.get('k') == 'Oogum_Boogum'):
        return Response("dec")
        # return Response("dns")
    elif(request.args.get('w') == '' and request.args.get('k') == 'Oogum_Boogum'):
        return Response('\x64\x6c\x02\x00')
    elif(request.args.get('d') == '' and request.args.get('k') == 'Oogum_Boogum'):
        return Response('C:\\users\\nic-lovin\\desktop\\enc.enc|C:\\users\\nic-lovin\\desktop\\dec.dec|')
    else:
        return Response("Unkown command")

if __name__ == '__main__':
    app.run(debug=False, port=80)