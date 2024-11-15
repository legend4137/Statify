from flask import Flask, jsonify, request
from services.recommendation_service import get_recommendations

app = Flask(__name__)

@app.route('/predict', methods=['POST'])
def get_recommendations_endpoint():
    data = request.get_json()
    user_id = data.get('user_id')
    recommendations = get_recommendations(user_id)
    print(recommendations)
    return jsonify({'songs': recommendations})

if __name__ == '__main__':
    app.run(debug=True)
