import numpy as np
import pandas as pd
from sklearn.decomposition import PCA
from sklearn.neighbors import NearestNeighbors
import tensorflow as tf
from pymongo import MongoClient

# MongoDB setup
client = MongoClient("mongodb+srv://b22ai015:uwhNGGuRUOGYSpij@musicrmdvic.5p7x3.mongodb.net/?retryWrites=true&w=majority&appName=MusicRmdVic")
db = client["data_ms"]
songs_data = pd.DataFrame(list(db["songs"].find()))
activity_data = pd.DataFrame(list(db["user_activity"].find()))

# Load TensorFlow Lite model
model_path = "./models/ncf_model.tflite"
interpreter = tf.lite.Interpreter(model_path=model_path)
interpreter.allocate_tensors()
input_details = interpreter.get_input_details()
output_details = interpreter.get_output_details()

def predict_with_tflite_model(user_vec, item_vec):
    input_data = np.array([user_vec, item_vec], dtype=np.float32)
    interpreter.set_tensor(input_details[0]['index'], input_data)
    interpreter.invoke()
    output_data = interpreter.get_tensor(output_details[0]['index'])
    return output_data.flatten()

def get_recommendations(user_id):
    top_20_songs = recommend_top_n(user_id)
    final_10_songs = filter_with_mood(user_id, top_20_songs)
    final_songs_details = get_final_song_details(user_id, final_10_songs)
    return final_songs_details

def recommend_top_n(user_id, n=20):
    item_ids = []
    num_users = 999
    num_items = 395532
    user_vec = np.full((num_items,), user_id)
    item_vec = np.arange(num_items)
    predicted_ratings = predict_with_tflite_model(user_vec, item_vec)
    top_n_indices = np.argsort(predicted_ratings)[-n:][::-1]
    return item_ids[top_n_indices]

def filter_with_mood(user_id, user_top_20):
    top_20_songs = songs_data[songs_data['Track'].isin(user_top_20)]
    features = top_20_songs[['Energy', 'Valence']].values
    pca = PCA(n_components=1)
    reduced_features = pca.fit_transform(features)
    
    user_mood = activity_data.loc[activity_data["user_id"] == user_id].iloc[0]
    mood_energy = user_mood["mood_energy"]
    mood_valence = user_mood["mood_valence"]
    
    user_mood_point = pca.transform([[mood_energy, mood_valence]])
    nn = NearestNeighbors(n_neighbors=10, algorithm='auto').fit(reduced_features)
    distances, indices = nn.kneighbors(user_mood_point)
    
    recommended_10_songs = top_20_songs.iloc[indices.flatten()]['Track'].values
    return recommended_10_songs

def get_final_song_details(user_id, final_10_songs):
    final_songs_details = {}
    for song_id in final_10_songs:
        song_data = songs_data.get(song_id)
        if song_data:
            song_name = song_data['Track']
            artist_name = song_data['Artist']
            final_songs_details[song_id] = {
                'track': song_name,
                'artist': artist_name
            }
    return final_songs_details
