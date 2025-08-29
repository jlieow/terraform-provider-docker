from flask import Flask

import os
from api.routes import api

# import logging
# log = logging.getLogger('werkzeug')
# log.setLevel(logging.DEBUG)

app = Flask(__name__)

app.register_blueprint(api, url_prefix="/api")

@app.route("/", methods = ['GET', 'POST'])
def hello_world():
    return f"Built by custom terraform provider!"

if __name__ == "__main__":
    app.run(debug=True, host="0.0.0.0", port=int(os.environ.get("PORT", 8080)))