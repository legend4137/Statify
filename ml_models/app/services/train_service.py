import pandas as pd
import numpy as np
from pymongo import MongoClient
import torch
import torch.nn as nn
import torch.optim as optim
from torch.utils.data import DataLoader
import os
from dotenv import load_dotenv, dotenv_values
from pyspark.sql import SparkSession
from pyspark.sql.functions import col
from datetime import datetime

load_dotenv()
DATABASE = os.getenv('MONGO_DATABASE')
MONGODB_URI = os.getenv('MONGO_URI')

client = MongoClient(MONGODB_URI)
db = client[DATABASE]

spark = SparkSession.builder \
    .appName("RetrieveDataFromMongoDB") \
    .config("spark.mongodb.input.uri", MONGODB_URI) \
    .config("spark.mongodb.output.uri", MONGODB_URI) \
    .getOrCreate()

def retrieve_all_new_data_spark(since):
    since_timestamp = since.strftime('%Y-%m-%dT%H:%M:%S')

    ratings_data = spark.read.format("mongo").option("collection", "ratings").load().filter(col("created_at") > since_timestamp)
    songs_data = spark.read.format("mongo").option("collection", "songs").load().filter(col("created_at") > since_timestamp)
    activity_data = spark.read.format("mongo").option("collection", "user_activity").load().filter(col("created_at") > since_timestamp)

    user_song_ratings = ratings_data.groupBy("user_id", "track_id").pivot("track_id").agg({"rating": "first"}).fillna(0)

    user_ids = user_song_ratings.select("user_id").distinct().rdd.flatMap(lambda x: x).collect()
    track_ids = user_song_ratings.columns[1:] 
    ratings = user_song_ratings.rdd.flatMap(lambda row: [row[i] for i in range(1, len(row))]).collect()

    genres = activity_data.select("preferred_genre").distinct().rdd.flatMap(lambda x: x).collect()
    languages = activity_data.select("preferred_language").distinct().rdd.flatMap(lambda x: x).collect()
    ages = activity_data.select("user_age").distinct().rdd.flatMap(lambda x: x).collect()
    genders = activity_data.select("user_gender").distinct().rdd.flatMap(lambda x: x).collect()

    return user_ids, track_ids, ratings, songs_data, genres, languages, ages, genders

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

def update_model_embeddings(model, n_users, n_items, n_genres, n_languages):
    if n_users > model.user_embedding.num_embeddings:
        model.user_embedding = nn.Embedding(n_users, model.user_embedding.embedding_dim)
        
    if n_items > model.item_embedding.num_embeddings:
        model.item_embedding = nn.Embedding(n_items, model.item_embedding.embedding_dim)
        
    if n_genres > model.genre_embedding.num_embeddings:
        model.genre_embedding = nn.Embedding(n_genres, model.genre_embedding.embedding_dim)
        
    if n_languages > model.language_embedding.num_embeddings:
        model.language_embedding = nn.Embedding(n_languages, model.language_embedding.embedding_dim)
        
    return model

def fine_tune_model(model, new_user_ids, new_track_ids, new_ratings, new_songs_data, genres, languages, ages, genders, n_users, n_items, n_genres, n_languages):
    model = update_model_embeddings(model, n_users, n_items, n_genres, n_languages)

    for param in model.parameters():
        param.requires_grad = False
    for param in model.output.parameters():
        param.requires_grad = True

    if new_user_ids:
        for new_user_id in new_user_ids:
            model.user_embedding.weight.data[new_user_id] = torch.randn(model.user_embedding.embedding_dim)

    new_train_dataset = RatingDataset(
        new_user_ids, new_track_ids, new_ratings,
        genres, languages, ages, genders
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
            model.item_embedding.weight.data[song_id] = torch.randn(model.item_embedding.embedding_dim)

    return model

def update_env_file(key, value):
    base_dir = os.path.dirname(__file__)
    env_file_path = os.path.join(base_dir, '.env')
    env_data = dotenv_values(env_file_path)

    env_data[key] = value

    with open(env_file_path, "w") as file:
        for k, v in env_data.items():
            file.write(f"{k}={v}\n")

def save_model(model):
    base_dir = os.path.dirname(__file__)
    models_dir = os.path.join(base_dir, '../models')
    model_path = os.path.join(models_dir, 'ncf_model.pth')
    torch.save(model.state_dict(), model_path)

def load_model(model):
    base_dir = os.path.dirname(__file__)
    models_dir = os.path.join(base_dir, '../models')
    model_path = os.path.join(models_dir, 'ncf_model.pth')
    model.load_state_dict(torch.load(model_path))
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

since_date = datetime(2024, 11, 16)
new_user_ids, new_track_ids, new_ratings, new_songs_data, genres, languages, ages, genders = retrieve_all_new_data_spark(since_date)

n_users = len(db["users"].distinct("user_id"))
num_items = len(db["songs"].distinct("track_id"))
n_genres = len(db["user_activity"].distinct("preferred_genre"))
n_languages = len(db["user_activity"].distinct("preferred_language"))

update_env_file("n_users", n_users)
update_env_file("num_items", num_items)
update_env_file("n_genres", n_genres)
update_env_file("n_languages", n_languages)

n_factors = 20
model = NCFWithDemographics(n_users, num_items, n_factors, n_genres, n_languages)
model = load_model(model)

model = fine_tune_model(model, new_user_ids, new_track_ids, new_ratings, new_songs_data, genres, languages, ages, genders, n_users, num_items, n_genres, n_languages)

save_model(model)
