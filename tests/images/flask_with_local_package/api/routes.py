from flask import Blueprint

api = Blueprint('api', __name__)

# Registered with a url prefix /api in server.py
# Accessed via /api/hello-api
@api.route("/hello-api", methods = ['GET'])
def hello_api():
    print("Hello api!")
    return f"Hello api!"