from flask import Flask, jsonify, request
from services.recommendation_service import get_recommendations

app = Flask(__name__)

@app.route('/predict', methods=['GET'])
def get_recommendations_endpoint():
    user_id = request.args.get('user_id')
    recommendations = get_recommendations(user_id)
    return jsonify({'songs': recommendations})

if __name__ == '__main__':
    app.run(debug=True)
