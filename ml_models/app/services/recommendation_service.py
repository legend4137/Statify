import torch
import torch.nn as nn
import pandas as pd
import numpy as np
from sklearn.decomposition import PCA
from sklearn.neighbors import NearestNeighbors
from pymongo import MongoClient
from sklearn.decomposition import PCA
from sklearn.neighbors import NearestNeighbors
import pickle
import os
from dotenv import load_dotenv

load_dotenv()
DATABASE = os.getenv('MONGO_DATABASE')
MONGODB_URI = os.getenv('MONGO_URI')

client = MongoClient(MONGODB_URI)
db = client[DATABASE]

class NCFWithDemographics(nn.Module):
    def __init__(self, n_users, n_items, n_factors, n_genres, n_languages):
        super(NCFWithDemographics, self).__init__()
        self.user_embedding = nn.Embedding(n_users, n_factors)
        self.item_embedding = nn.Embedding(n_items, n_factors)
        
        self.genre_embedding = nn.Embedding(n_genres, n_factors)
        self.language_embedding = nn.Embedding(n_languages, n_factors)
        self.age_embedding = nn.Embedding(100, n_factors)
        self.gender_embedding = nn.Embedding(2, n_factors)
        
        self.fc1 = nn.Linear(n_factors * 6, 64)
        self.fc2 = nn.Linear(64, 32)
        self.output = nn.Linear(32, 1)
        
    def forward(self, user_id, item_id, genre_id, language_id, age, gender):
        user_vec = self.user_embedding(user_id)
        item_vec = self.item_embedding(item_id)
        genre_vec = self.genre_embedding(genre_id)
        language_vec = self.language_embedding(language_id)
        age_vec = self.age_embedding(age)
        gender_vec = self.gender_embedding(gender)
        
        x = torch.cat([user_vec, item_vec, genre_vec, language_vec, age_vec, gender_vec], dim=1)
        x = torch.relu(self.fc1(x))
        x = torch.relu(self.fc2(x))
        return torch.sigmoid(self.output(x))


def load_model_and_encoders(n_users, n_items, n_factors, n_genres, n_languages):
    # Get the directory of this script
    base_dir = os.path.dirname(__file__)
    models_dir = os.path.join(base_dir, '../models')

    model = NCFWithDemographics(n_users, n_items, n_factors, n_genres, n_languages)
    model_path = os.path.join(models_dir, 'ncf_model.pth')
    model.load_state_dict(torch.load(model_path))
    model.eval()
    with open(os.path.join(models_dir, 'genre_encoder.pkl'), 'rb') as f:
        genre_encoder = pickle.load(f)
    with open(os.path.join(models_dir, 'language_encoder.pkl'), 'rb') as f:
        language_encoder = pickle.load(f)
    return model, genre_encoder, language_encoder

def recommend_top_n(user_id, model, genre_id, language_id, age, gender, num_items, n=20):
    item_vec = torch.arange(num_items)
    user_vec = torch.full((num_items,), int(user_id))
    genre_vec = torch.full((num_items,), genre_id)
    language_vec = torch.full((num_items,), language_id)
    age_vec = torch.full((num_items,), age)
    gender_vec = torch.full((num_items,), gender)

    with torch.no_grad():
        predictions = model(user_vec, item_vec, genre_vec, language_vec, age_vec, gender_vec).squeeze().numpy()
    
    top_n_indices = np.argsort(predictions)[-n:][::-1].copy()
    return item_vec[top_n_indices].numpy()

def filter_with_mood(user_id, user_top_20, db):
    songs_collection = db['songs']
    activity_collection = db['user_activity']

    song_cursor = songs_collection.find(
        {"track_id": {"$in": user_top_20.tolist()}},
        {"track_id": 1, "energy": 1, "valence": 1}
    )
    top_20_songs = list(song_cursor)

    track_ids = [song['track_id'] for song in top_20_songs]
    features = np.array([[song['energy'], song['valence']] for song in top_20_songs])

    pca = PCA(n_components=1)
    reduced_features = pca.fit_transform(features)

    user_mood = activity_collection.find_one(
        {"user_id": user_id},
        {"mood_energy": 1, "mood_valence": 1}
    )

    mood_energy = user_mood["mood_energy"]
    mood_valence = user_mood["mood_valence"]
    user_mood_point = pca.transform([[mood_energy, mood_valence]])

    nn = NearestNeighbors(n_neighbors=5)
    nn.fit(reduced_features)
    distances, indices = nn.kneighbors(user_mood_point)

    final_10_tracks = [track_ids[i] for i in indices.flatten()]
    return final_10_tracks

def retrieve_user_info(user_id):
    user_data = pd.DataFrame(list(db["users"].find({"user_id": user_id})))
    activity_data = pd.DataFrame(list(db["user_activity"].find({"user_id": user_id})))

    if user_data.empty or activity_data.empty:
        raise ValueError(f"{user_id}{user_data}{activity_data}No data found for user_id: {user_id}")
    
    age = user_data.iloc[0]["user_age"]
    gender = user_data.iloc[0]["user_gender"]
    preferred_language = activity_data.iloc[0]["preferred_language"]
    preferred_genre = activity_data.iloc[0]["preferred_genre"]
    
    return age, gender, preferred_language, preferred_genre

def get_recommendations(user_id):
    user_id = int(user_id)
    n_users = int(os.getenv('n_users'))
    num_items = int(os.getenv('num_items'))
    n_factors = 20
    n_genres= int(os.getenv('n_genres'))
    n_languages = int(os.getenv('n_languages'))
    model, genre_encoder, language_encoder = load_model_and_encoders(
        n_users, num_items, n_factors, n_genres, n_languages
    )

    age, gender, preferred_language, preferred_genre = retrieve_user_info(user_id)

    gender = 1 if gender.lower() == "male" else 0
    genre_id = genre_encoder.transform([preferred_genre])[0]
    language_id = language_encoder.transform([preferred_language])[0]

    top_20_songs_per_user = recommend_top_n(user_id, model, genre_id, language_id, age, gender, num_items)
    final_10_songs = filter_with_mood(user_id, top_20_songs_per_user, db)

    return final_10_songs