import numpy as np

def encrypt(message, key):

    message_points = np.array([ord(char) for char in message])

    if message_points.shape[0] % key.shape[0] != 0:
        padding = key.shape[0] - (message_points.shape[0] % key.shape[0])
        message_points = np.concatenate([message_points, np.zeros(padding)])

    message_matrix = message_points.reshape((-1, key.shape[0])).T
    encoded_matrix = np.dot(key, message_matrix) % 251

    return bytearray(int(code) for code in encoded_matrix.T.flatten())
    
key = np.array([[n1, n2], [n3, n4]])
message = "FLAG:"+"HF-fakeflag"
encoded_message = encrypt(message, key)
print(encoded_message)
