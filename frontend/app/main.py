import streamlit as st
import requests
import json

BASE_URL = 'http://localhost:4567'
LOGIN_URL = f'{BASE_URL}/login'
REGISTER_URL = f'{BASE_URL}/register'
GO_PREDICT_URL = f'{BASE_URL}/predict'
SEARCH_URL = f'{BASE_URL}/search_song'
RATE_SONG_URL = f'{BASE_URL}/rate_song'

def login(email, password):
    try:
        response = requests.post(LOGIN_URL, json={
            "email": email,
            "password": password
        })
        if response.status_code == 200:
            return response.json()
        else:
            print(response)
            st.error("Invalid credentials")
            return None
    except Exception as e:
        st.error(f"Error: {str(e)}")
        return None

def register(email, password, user_name, user_age, user_gender, user_song_language):
    try:
        response = requests.post(REGISTER_URL, json={
            "email": email,
            "password": password,
            "user_name": user_name,
            "user_age": user_age,
            "user_gender": user_gender,
            "user_song_language": user_song_language
        })
        if response.status_code == 200:
            st.success("Registration successful! Please login.")
            return True
        else:
            st.error(f"Registration failed: {response.json().get('error', 'Unknown error')}")
            return False
    except Exception as e:
        st.error(f"Error: {str(e)}")
        return False

def search_song(song_name, artist_name):
    search_payload = {
        "song_name": song_name,
        "artist_name": artist_name
    }
    response = requests.get(SEARCH_URL, json=search_payload)
    if response.status_code == 200:
        return response.json()
    else:
        st.error(f"Error: {response.json().get('error')}")
        return None

def get_recommendations(user_id):
    try:
        response = requests.post(GO_PREDICT_URL, json={"user_id": str(user_id)})
        if response.status_code == 200:
            return response.json().get('songs', [])
        else:
            st.error("Failed to fetch recommendations")
            return []
    except Exception as e:
        st.error(f"Error: {str(e)}")
        return []
    
def main():
    st.title('Music Recommendation System')

    if 'logged_in' not in st.session_state:
        st.session_state.logged_in = False
    if 'user_data' not in st.session_state:
        st.session_state.user_data = None

    if not st.session_state.logged_in:
        tab1, tab2 = st.tabs(["Login", "Register"])
        
        with tab1:
            st.subheader("Login")
            login_email = st.text_input("Email", key="login_email")
            login_password = st.text_input("Password", type="password", key="login_password")
            
            if st.button("Login"):
                if login_email and login_password:
                    user_data = login(login_email, login_password)
                    if user_data:
                        st.session_state.logged_in = True
                        st.session_state.user_data = user_data
                        st.rerun()
                else:
                    st.error("Please fill in all fields")

        with tab2:
            st.subheader("Register")
            reg_email = st.text_input("Email", key="reg_email")
            reg_password = st.text_input("Password", type="password", key="reg_password")
            user_name = st.text_input("Full Name")
            user_age = st.number_input("Age", min_value=1, max_value=120)
            user_gender = st.selectbox("Gender", ["Male", "Female", "Other"])
            user_song_language = st.selectbox("Preferred Song Language", ["English", "Hindi", "Spanish", "Other"])

            if st.button("Register"):
                if all([reg_email, reg_password, user_name, user_age, user_gender, user_song_language]):
                    if register(reg_email, reg_password, user_name, user_age, user_gender, user_song_language):
                        st.success("Please go to login tab to sign in")
                else:
                    st.error("Please fill in all fields")

    else:
        st.write(f"Welcome, {st.session_state.user_data['user_name']}!")
        st.write(f"User ID: {st.session_state.user_data['user_id']}")
        
        if st.button("Logout"):
            st.session_state.logged_in = False
            st.session_state.user_data = None
            st.rerun()

        search_col, recommend_col = st.columns(2)

        with search_col:
            st.subheader('Search for a Song')
            search_song_name = st.text_input('Song Name', '')
            search_artist_name = st.text_input('Artist Name', '')

            if st.button('Search Song'):
                if search_song_name and search_artist_name:
                    song_details = search_song(search_song_name, search_artist_name)
                    if song_details:
                        # st.subheader('Song Details')
                        # st.image(song_details['image_url'], caption=f"{song_details['track']} by {song_details['artist']}", width=200)
                        # st.write(f"Album: {song_details['album']}")
                        embed_url = f"https://open.spotify.com/embed/track/{song_details['id']}"
                        st.components.v1.iframe(embed_url, width=300, height=380)
                else:
                    st.error("Please enter both song name and artist name")

        # Right column for recommendations
        with recommend_col:
            st.subheader('Recommended Songs')
            if st.button("Get Recommendations"):
                recommended_songs = get_recommendations(st.session_state.user_data['user_id'])
                
                if recommended_songs:
                    for song in recommended_songs:
                        # st.image(song['image_url'], caption=f"{song['track']} by {song['artist']}", width=200)
                        # st.write(f"Album: {song['album']}")
                        # st.write(f"[Listen on Spotify]({song['spotify_url']})")
                        embed_url = f"https://open.spotify.com/embed/track/{song['id']}"
                        st.components.v1.iframe(embed_url, width=300, height=380)
                        # if st.button(f"Rate {song['track']}"):
                            # rate_song(st.session_state.user_data['_id'], song['track_uri'])

if __name__ == '__main__':
    main()