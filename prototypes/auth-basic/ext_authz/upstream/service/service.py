from flask import Flask, request
import pickle

app = Flask(__name__)


@app.route('/service')
def hello():
    return 'upstream.service: Hello ' + request.headers.get('x-current-user') + ' from behind Envoy!'

@app.route('/')
def root():
    return 'upstream.service: Hello from behind Envoy!\n----\n' + str(request.headers) + '\n----\n'

if __name__ == "__main__":
    app.run(host='0.0.0.0', port=8080)
