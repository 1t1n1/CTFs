from flask import Blueprint, request, jsonify, send_file
from Crypto.Cipher import AES
from Crypto.Random import get_random_bytes
import hashlib
from cryptography.hazmat.primitives import hashes
from cryptography.hazmat.primitives.asymmetric import padding
from cryptography.hazmat.primitives.serialization import load_pem_public_key
from cryptography.exceptions import InvalidSignature
import json
import socket
from io import BytesIO

def decrypt_data(data, key):
    iv = data[:16]
    tag = data[-16:]
    ciphertext = data[16:-16]
    cipher = AES.new(key, AES.MODE_GCM, nonce=iv)
    plaintext = cipher.decrypt_and_verify(ciphertext, tag)
    return plaintext.decode()

def generate_encryption_key():
    key = hashlib.sha256('cai-gateway'.encode() + b'BackUp05@').digest()
    return key

with open('/home/pegasus/Downloads/config.json', 'rb') as f:
    data = f.read()

key = generate_encryption_key()

print(decrypt_data(data, key))