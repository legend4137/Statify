import streamlit as st
import requests
import json

GO_PREDICT_URL = 'http://localhost:4567/predict'
SEARCH_URL = 'http://localhost:4567/search_song'
RATE_SONG_URL = 'http://localhost:4567/rate_song'

def get_user_id():
    user_id = st.text_input('Enter your User ID', '')
    return user_id

def search_song(song_name, artist_name):
    search_payload = {
        "song_name": song_name,
        "artist_name": artist_name
    }
    response = requests.post(SEARCH_URL, json=search_payload)
    if response.status_code == 200:
        return response.json()
    else:
        st.error(f"Error: {response.json().get('error')}")
        return None

def get_recommendations(user_id):
    response = requests.post(GO_PREDICT_URL, json={"user_id": user_id})
    if response.status_code == 200:
        return response.json()
    else:
        st.error(f"Error: {response.json().get('error')}")
        return []

def rate_song(user_id, track_id):
    rate_payload = {
        "user_id": user_id,
        "track_id": track_id
    }
    response = requests.post(RATE_SONG_URL, json=rate_payload)
    if response.status_code == 200:
        st.success('Song rated successfully')
    else:
        st.error(f"Error: {response.json().get('error')}")

def main():
    st.title('Music Recommendation System')
    
    user_id = get_user_id()
    
    if user_id:
        st.subheader('Recommended Songs')
        recommended_songs = get_recommendations(user_id)
        
        if recommended_songs:
            for song in recommended_songs:
                track_name = song['track']
                artist_name = song['artist']
                album = song['album']
                image_url = song['image_url']
                track_uri = song['track_uri']
                spotify_url = song['spotify_url']
                
                st.image(image_url, caption=f"{track_name} by {artist_name}", width=200)
                st.write(f"Album: {album}")
                st.write(f"[Listen on Spotify]({spotify_url})")
                
                if st.button(f"Rate {track_name}"):
                    rate_song(user_id, track_uri)

        st.subheader('Search for a Song')
        search_song_name = st.text_input('Song Name', '')
        search_artist_name = st.text_input('Artist Name', '')
        
        if st.button('Search Song'):
            if search_song_name and search_artist_name:
                song_details = search_song(search_song_name, search_artist_name)
                if song_details:
                    st.subheader('Song Details')
                    st.image(song_details['image_url'], caption=f"{song_details['track']} by {song_details['artist']}", width=200)
                    st.write(f"Album: {song_details['album']}")
                    st.write(f"[Listen on Spotify]({song_details['spotify_url']})")
            else:
                st.error("Please enter both song name and artist name")

if __name__ == '__main__':
    main()
