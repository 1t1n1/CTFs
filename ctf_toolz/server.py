#!/usr/bin/env python3

import argparse
import os
import subprocess
from logging.handlers import RotatingFileHandler
import logging
from flask import Flask, request, jsonify, send_file

app = Flask(__name__)

@app.route('/file', methods=['GET'])
def serve_file():
    # /file?file=filename
    file_name = request.args.get('file', 'file')
    # Prevent path traversal
    if not file_name or '/' in file_name:
        return "Invalid file request", 400
    file_path = os.path.join(os.path.dirname(__file__), file_name)
    if os.path.exists(file_path):
        return send_file(file_path)
    else:
        return "File not found", 404

@app.route('/upload', methods=['POST'])
def upload():
    # data is in the POST data field
    data = request.get_data()
    filename = request.form.get('filename')
    data = data[data.find(b'&') + 1:]
    if filename is not None:
        if not os.path.exists('uploaded_files'):
            os.makedirs('uploaded_files')
        with open(os.path.join('uploaded_files', filename), 'wb') as f:
            f.write(data)
    else:
        print(data.decode())
    return "OK"

@app.route('/text', methods=['GET'])
def serve_text():
    return "This is a template text."

@app.route('/json', methods=['GET'])
def serve_json():
    return jsonify({"message": "This is a template JSON."})

@app.route('/xml', methods=['GET'])
def serve_xml():
    return "<root>This is a template XML.</root>"

@app.route('/html', methods=['GET'])
def serve_html():
    return "<html><body>This is a template HTML.</body></html>"

@app.route('/javascript', methods=['GET'])
def serve_javascript():
    return "console.log('This is a template JavaScript.');"

@app.route('/python', methods=['GET'])
def serve_python():
    return "print('This is a template Python.')"

@app.route('/php', methods=['GET'])
def serve_php():
    return "<?php echo 'This is a template PHP.'; ?>"

@app.route('/exec', methods=['GET'])
def execute_command():
    # /exec?cmd=the_command
    # Define a whitelist of allowed commands
    allowed_commands = []

    cmd = request.args.get('cmd', '')
    if cmd in allowed_commands:
        try:
            output = subprocess.check_output(cmd, shell=True, stderr=subprocess.STDOUT, text=True)
            app.logger.info(f"Command executed: {cmd}")
            app.logger.info(f"Output:\n{output}")
            return output
        except subprocess.CalledProcessError as e:
            return str(e.output), 500
    else:
        return "Not allowed", 403

@app.route('/print_request', methods=['GET'])
def print_request():
    request_data = {
        "Method": request.method,
        "Path": request.path,
        "Headers": dict(request.headers),
        "Query Params": request.args.to_dict(),
    }
    app.logger.info(f"Received Request: {request_data}")
    return "Request printed in logs."

if __name__ == '__main__':
    parser = argparse.ArgumentParser(description='A simple Flask server to answer request. Useful in CTFs')
    parser.add_argument('--ip', type=str, default='127.0.0.1', help='IP address to bind to')
    parser.add_argument('--port', type=int, default=1337, help='Port to listen on')
    args = parser.parse_args()
    
    # Create a logger
    logger = logging.getLogger(__name__)
    logger.setLevel(logging.INFO)
    
    # Create a file handler for the log file
    log_file_handler = RotatingFileHandler('server.log', maxBytes=1024 * 1024, backupCount=10)
    log_file_handler.setLevel(logging.INFO)
    
    # Create a console handler
    console_handler = logging.StreamHandler()
    console_handler.setLevel(logging.INFO)
    
    # Create a formatter and add it to the handlers
    formatter = logging.Formatter('%(asctime)s - %(name)s - %(levelname)s - %(message)s')
    log_file_handler.setFormatter(formatter)
    console_handler.setFormatter(formatter)
    
    # Add the handlers to the logger
    logger.addHandler(log_file_handler)
    logger.addHandler(console_handler)
    
    app.logger = logger
    
    app.run(host=args.ip, port=args.port)

