# Music Recommender Service

## Setup Instructions

1. Clone the repository:
    ```bash
    git clone <repository-url>
    cd ml_models
    ```

2. Install dependencies:
    ```bash
    pip install -r requirements.txt
    ```

3. Run the Flask app:
    ```bash
    python app/main.py
    ```

4. The API will be available at `http://localhost:5000`.

5. To get recommendations for a user:
    ```bash
    GET /recommendations/<user_id>
    ```

6. The response will be in JSON format with the recommended song details.

## Files Structure:
- **app/main.py**: Flask app for serving ML models
- **models/genre_model.pkl**: Pretrained genre model
- **models/behavior_model.pkl**: Pretrained behavior model
- **services/recommendation_service.py**: Logic for generating recommendations
