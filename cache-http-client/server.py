from flask import Flask, request, jsonify
import memcache, os

app = Flask(__name__)

# Retrieve port from environment variable or use default 8080
PORT = int(os.environ.get('PORT', 8080))
# Retrieve memcached server address from environment variable or use default 'memcached:11211'
MEMCACHED_SERVER = os.environ.get('CACHE_SERVER', 'memcached:11211')
cache = memcache.Client([MEMCACHED_SERVER], debug=0)

@app.route('/', methods=['GET'])
def get_value():
    key = request.args.get('key')
    if key:
        cached_value = cache.get(key)
        if cached_value is not None:
            return jsonify({key: cached_value})
        else:
            return jsonify({"error": f"Key '{key}' not found"}), 404
    else:
        return jsonify({"error": "Please provide a 'key' parameter in the query"}), 400

@app.route('/', methods=['POST'])
def set_value():
    data = request.get_json()
    if 'key' not in data or 'value' not in data:
        return jsonify({"error": "Please provide both 'key' and 'value' in the request body"}), 400
    key = data['key']
    value = data['value']
    cache.set(key, value)
    return jsonify({"success": True})

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=PORT)
