# Astolfo's Player Server

REST API server for synchronizing music files between devices for Astolfo's Player.

## Features

- ✅ User registration and authentication (JWT)
- ✅ Uploading music files (MP3, FLAC, WAV, OGG, Opus, M4A, AAC, WMA)
- ✅ Track downloading
- ✅ Synchronization via SHA256 hashes
- ✅ Track metadata (title, artist, album)
- ✅ Bulk file upload
- ✅ Automatic duplicate detection

## Quick Start

### 1. Clone and navigate to the directory

```bash
mkdir astolfo-player-server
cd astolfo-player-server
```

### 2. Create files

Create files from artifacts:
- `main.py` - main server code
- `requirements.txt` - Python dependencies
- `Dockerfile` - Docker image
- `docker-compose.yml` - container configuration

### 3. Start with Docker Compose

```bash
docker-compose up -d
```

The server will be available at: `http://localhost:8000`

### 4. Check the server

```bash
curl http://localhost:8000
```

## API Endpoints

### Authentication

**Registration**
```http
POST /api/auth/register
Content-Type: application/json

{
  "username": "user",
  "password": "password123"
}
```

**Login**
```http
POST /api/auth/login
Content-Type: application/json

{
  "username": "user",
  "password": "password123"
}
```

Response:
```json
{
  "access_token": "eyJ...",
  "token_type": "bearer"
}
```

### Track Management

**Get all tracks**
```http
GET /api/tracks
Authorization: Bearer {token}
```

**Get a specific track**
```http
GET /api/tracks/{track_id}
Authorization: Bearer {token}
```

**Download file**
```http
GET /api/tracks/{track_id}/file
Authorization: Bearer {token}
```

**Upload track**
```http
POST /api/tracks/upload
Authorization: Bearer {token}
Content-Type: multipart/form-data

file: [binary data]
title: "Song Title" (optional)
artist: "Artist Name" (optional)
album: "Album Name" (optional)
duration: 180 (optional, seconds)
```

**Delete track**
```http
DELETE /api/tracks/{track_id}
Authorization: Bearer {token}
```

### Synchronization

**Get synchronization status**
```http
GET /api/sync/status
Authorization: Bearer {token}
```

Response:
```json
[
  {
    "file_hash": "a1b2c3...",
    "filename": "song.mp3",
    "uploaded_at": "2025-11-01T10:00:00"
  }
]
```

**Bulk upload**
```http
POST /api/sync/batch-upload
Authorization: Bearer {token}
Content-Type: multipart/form-data

files: [file1, file2, file3, ...]
```

## Usage Examples with Python

### Registration and obtaining a token

```python
import requests

BASE_URL = "http://localhost:8000"

# Registration
response = requests.post(
    f"{BASE_URL}/api/auth/register",
    json={"username": "testuser", "password": "testpass123"}
)
token = response.json()["access_token"]

# Authorization headers
headers = {"Authorization": f"Bearer {token}"}
```

### Uploading a track

```python
with open("song.mp3", "rb") as f:
    files = {"file": ("song.mp3", f, "audio/mpeg")}
    data = {
        "title": "My Song",
        "artist": "My Artist",
        "album": "My Album",
        "duration": 180
    }
    response = requests.post(
        f"{BASE_URL}/api/tracks/upload",
        files=files,
        data=data,
        headers=headers
    )
    track = response.json()
    print(f"Uploaded track: {track['id']}")
```

### File synchronization

```python
import hashlib
from pathlib import Path

def calculate_hash(file_path):
    sha256 = hashlib.sha256()
    with open(file_path, "rb") as f:
        for chunk in iter(lambda: f.read(4096), b""):
            sha256.update(chunk)
    return sha256.hexdigest()

# Get hashes from server
response = requests.get(f"{BASE_URL}/api/sync/status", headers=headers)
server_hashes = {item["file_hash"] for item in response.json()}

# Scan local folder
music_folder = Path("./music")
for music_file in music_folder.glob("*.mp3"):
    file_hash = calculate_hash(music_file)
    
    if file_hash not in server_hashes:
        # Upload new file
        with open(music_file, "rb") as f:
            files = {"file": (music_file.name, f, "audio/mpeg")}
            requests.post(
                f"{BASE_URL}/api/tracks/upload",
                files=files,
                headers=headers
            )
        print(f"Uploaded: {music_file.name}")
    else:
        print(f"Already exists: {music_file.name}")
```

### Downloading all tracks

```python
# Get list of all tracks
response = requests.get(f"{BASE_URL}/api/tracks", headers=headers)
tracks = response.json()

# Download each track
download_folder = Path("./downloads")
download_folder.mkdir(exist_ok=True)

for track in tracks:
    response = requests.get(
        f"{BASE_URL}/api/tracks/{track['id']}/file",
        headers=headers,
        stream=True
    )
    
    file_path = download_folder / track["filename"]
    with open(file_path, "wb") as f:
        for chunk in response.iter_content(chunk_size=8192):
            f.write(chunk)
    
    print(f"Downloaded: {track['filename']}")
```

## Configuration

Settings are specified via environment variables in `docker-compose.yml`:

- `DATABASE_URL` - PostgreSQL connection string
- `SECRET_KEY` - secret key for JWT (make sure to change in production!)
- `STORAGE_PATH` - path for storing files

## Development without Docker

```bash
# Install dependencies
pip install -r requirements.txt

# Install PostgreSQL locally
# Create the musicdb database

# Start the server
export DATABASE_URL="postgresql://user:pass@localhost:5432/musicdb"
export SECRET_KEY="dev-secret-key"
export STORAGE_PATH="./storage"

uvicorn main:app --reload --host 0.0.0.0 --port 8000
```

## Security

⚠️ **Important for production:**

1. Change `SECRET_KEY` to a random string
2. Use HTTPS (set up reverse proxy with nginx)
3. Add rate limiting (e.g., with slowapi)
4. Configure CORS if web clients need access
5. Enable logging and monitoring
6. Regularly back up the database and files

## Swagger Documentation

After starting the server, the API documentation is available at:
- http://localhost:8000/docs (Swagger UI)
- http://localhost:8000/redoc (ReDoc)

## Logs

Viewing logs:
```bash
docker-compose logs -f api
```

## Stopping

```bash
docker-compose down
```

Delete all data (DB and files):
```bash
docker-compose down -v
```
