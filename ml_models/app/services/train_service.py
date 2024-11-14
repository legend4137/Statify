import pandas as pd
import numpy as np
from pymongo import MongoClient
import torch
import torch.nn as nn
import torch.optim as optim
from torch.utils.data import DataLoader

client = MongoClient('mongodb+srv://b22ai015:mel7iKthsBpNR6d3@msrmd.tgazz.mongodb.net/?retryWrites=true&w=majority&appName=MsRmd')
db = client['music_recommendation_system']

def retrieve_all_new_data():
    ratings_data = pd.DataFrame(list(db["buffer_ratings"].find()))
    songs_data = pd.DataFrame(list(db["buffer_songs"].find()))
    activity_data = pd.DataFrame(list(db["buffer_user_activity"].find()))
    user_data = pd.DataFrame(list(db["buffer_users"].find()))

    user_song_ratings = pd.pivot_table(ratings_data, values='rating', index='user_id', columns='track_id').fillna(0)

    user_ids, track_ids = np.nonzero(user_song_ratings)
    ratings = [user_song_ratings.iloc[user, track] for user, track in zip(user_ids, track_ids)]
    merged_data = ratings_data.merge(activity_data, on="user_id", how="left").merge(user_data, on="user_id", how="left")
    
    return user_ids, track_ids, np.array(ratings), songs_data, merged_data

class RatingDataset(torch.utils.data.Dataset):
    def __init__(self, user_ids, track_ids, ratings, genres, languages, ages, genders):
        self.user_ids = user_ids
        self.track_ids = track_ids
        self.ratings = ratings
        self.genres = genres
        self.languages = languages
        self.ages = ages
        self.genders = genders

    def __len__(self):
        return len(self.user_ids)

    def __getitem__(self, idx):
        return (
            self.user_ids[idx],
            self.track_ids[idx],
            self.ratings[idx],
            self.genres[idx],
            self.languages[idx],
            self.ages[idx],
            self.genders[idx]
        )
    
def fine_tune_model(model, new_user_ids, new_track_ids, new_ratings, new_songs_data, new_merged_data):
    for param in model.parameters():
        param.requires_grad = False
    for param in model.output.parameters():
        param.requires_grad = True

    if new_user_ids:
        for new_user_id in new_user_ids:
            model.user_embeddings[new_user_id] = torch.randn(1, model.user_embedding_dim)

    new_train_dataset = RatingDataset(
        new_user_ids, new_track_ids, new_ratings,
        new_merged_data.get('preferred_genre', []),
        new_merged_data.get('preferred_language', []),
        new_merged_data.get('user_age', []),
        new_merged_data.get('user_gender', [])
    )
    new_train_loader = DataLoader(new_train_dataset, batch_size=256, shuffle=True)

    criterion = nn.MSELoss()
    optimizer = optim.Adam(model.parameters(), lr=0.001)

    model.train()
    for user_id, item_id, rating, genre_id, language_id, age, gender in new_train_loader:
        optimizer.zero_grad()
        prediction = model(user_id, item_id, genre_id, language_id, age, gender)
        loss = criterion(prediction.squeeze(), rating)
        loss.backward()
        optimizer.step()

    if new_songs_data:
        for song_data in new_songs_data:
            song_id = song_data['track_id']
            model.song_embeddings[song_id] = torch.randn(1, model.song_embedding_dim)

    return model

def save_model(model, file_path):
    torch.save(model.state_dict(), file_path)

def load_model(model, file_path):
    model.load_state_dict(torch.load(file_path))
    model.eval()
    return model

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

new_user_ids, new_track_ids, new_ratings, new_songs_data, new_merged_data = retrieve_all_new_data()

n_users = len(new_merged_data['user_id'].unique())
n_items = len(new_songs_data['track_id'].unique())
n_factors = 20
n_genres = len(new_merged_data['preferred_genre'].unique())
n_languages = len(new_merged_data['preferred_language'].unique())

model = NCFWithDemographics(n_users, n_items, n_factors, n_genres, n_languages)
model = load_model(model, 'ncf_model.pth')

model = fine_tune_model(model, new_user_ids, new_track_ids, new_ratings, new_songs_data, new_merged_data)
save_model(model, 'ncf_model.pth')
