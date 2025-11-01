# Astolfo's Player Server

REST API сервер для синхронизации музыкальных файлов между устройствами для Astolfo's Player.

## Возможности

- ✅ Регистрация и аутентификация пользователей (JWT)
- ✅ Загрузка музыкальных файлов (MP3, FLAC, WAV, OGG, Opus, M4A, AAC, WMA)
- ✅ Скачивание треков
- ✅ Синхронизация по SHA256 хешам
- ✅ Метаданные треков (название, исполнитель, альбом)
- ✅ Массовая загрузка файлов
- ✅ Автоматическое определение дубликатов

## Быстрый старт

### 1. Клонировать и перейти в директорию

```bash
mkdir astolfo-player-server
cd astolfo-player-server
```

### 2. Создать файлы

Создайте файлы из артефактов:
- `main.py` - основной код сервера
- `requirements.txt` - зависимости Python
- `Dockerfile` - образ Docker
- `docker-compose.yml` - конфигурация контейнеров

### 3. Запустить с Docker Compose

```bash
docker-compose up -d
```

Сервер будет доступен по адресу: `http://localhost:8000`

### 4. Проверить работу

```bash
curl http://localhost:8000
```

## API Endpoints

### Аутентификация

**Регистрация**
```http
POST /api/auth/register
Content-Type: application/json

{
  "username": "user",
  "password": "password123"
}
```

**Вход**
```http
POST /api/auth/login
Content-Type: application/json

{
  "username": "user",
  "password": "password123"
}
```

Ответ:
```json
{
  "access_token": "eyJ...",
  "token_type": "bearer"
}
```

### Работа с треками

**Получить все треки**
```http
GET /api/tracks
Authorization: Bearer {token}
```

**Получить конкретный трек**
```http
GET /api/tracks/{track_id}
Authorization: Bearer {token}
```

**Скачать файл**
```http
GET /api/tracks/{track_id}/file
Authorization: Bearer {token}
```

**Загрузить трек**
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

**Удалить трек**
```http
DELETE /api/tracks/{track_id}
Authorization: Bearer {token}
```

### Синхронизация

**Получить статус синхронизации**
```http
GET /api/sync/status
Authorization: Bearer {token}
```

Ответ:
```json
[
  {
    "file_hash": "a1b2c3...",
    "filename": "song.mp3",
    "uploaded_at": "2025-11-01T10:00:00"
  }
]
```

**Массовая загрузка**
```http
POST /api/sync/batch-upload
Authorization: Bearer {token}
Content-Type: multipart/form-data

files: [file1, file2, file3, ...]
```

## Примеры использования с Python

### Регистрация и получение токена

```python
import requests

BASE_URL = "http://localhost:8000"

# Регистрация
response = requests.post(
    f"{BASE_URL}/api/auth/register",
    json={"username": "testuser", "password": "testpass123"}
)
token = response.json()["access_token"]

# Заголовки с авторизацией
headers = {"Authorization": f"Bearer {token}"}
```

### Загрузка трека

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

### Синхронизация файлов

```python
import hashlib
from pathlib import Path

def calculate_hash(file_path):
    sha256 = hashlib.sha256()
    with open(file_path, "rb") as f:
        for chunk in iter(lambda: f.read(4096), b""):
            sha256.update(chunk)
    return sha256.hexdigest()

# Получить хеши с сервера
response = requests.get(f"{BASE_URL}/api/sync/status", headers=headers)
server_hashes = {item["file_hash"] for item in response.json()}

# Сканировать локальную папку
music_folder = Path("./music")
for music_file in music_folder.glob("*.mp3"):
    file_hash = calculate_hash(music_file)
    
    if file_hash not in server_hashes:
        # Загрузить новый файл
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

### Скачивание всех треков

```python
# Получить список всех треков
response = requests.get(f"{BASE_URL}/api/tracks", headers=headers)
tracks = response.json()

# Скачать каждый трек
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

## Конфигурация

Настройки задаются через переменные окружения в `docker-compose.yml`:

- `DATABASE_URL` - строка подключения к PostgreSQL
- `SECRET_KEY` - секретный ключ для JWT (обязательно изменить в продакшене!)
- `STORAGE_PATH` - путь для хранения файлов

## Разработка без Docker

```bash
# Установить зависимости
pip install -r requirements.txt

# Установить PostgreSQL локально
# Создать базу данных musicdb

# Запустить сервер
export DATABASE_URL="postgresql://user:pass@localhost:5432/musicdb"
export SECRET_KEY="dev-secret-key"
export STORAGE_PATH="./storage"

uvicorn main:app --reload --host 0.0.0.0 --port 8000
```

## Безопасность

⚠️ **Важно для продакшена:**

1. Измените `SECRET_KEY` на случайную строку
2. Используйте HTTPS (настройте reverse proxy с nginx)
3. Добавьте rate limiting (например, с slowapi)
4. Настройте CORS если нужен доступ с веб-клиентов
5. Включите логирование и мониторинг
6. Регулярно делайте бэкапы БД и файлов

## Swagger документация

После запуска сервера документация API доступна по адресу:
- http://localhost:8000/docs (Swagger UI)
- http://localhost:8000/redoc (ReDoc)

## Логи

Просмотр логов:
```bash
docker-compose logs -f api
```

## Остановка

```bash
docker-compose down
```

Удалить все данные (БД и файлы):
```bash
docker-compose down -v
```