from fastapi import FastAPI, HTTPException, Depends, UploadFile, File, Form
from fastapi.security import HTTPBearer, HTTPAuthorizationCredentials
from fastapi.responses import FileResponse, StreamingResponse
from sqlalchemy import create_engine, Column, String, Integer, BigInteger, DateTime, ForeignKey
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.orm import sessionmaker, Session, relationship
from passlib.context import CryptContext
from pydantic import BaseModel
from typing import Optional, List
from datetime import datetime, timedelta
import jwt
import hashlib
import os
import shutil
from pathlib import Path
import uuid

# Конфигурация
DATABASE_URL = os.getenv("DATABASE_URL", "postgresql://astolfo:astolfopass@localhost:5432/astolfodb")
SECRET_KEY = os.getenv("SECRET_KEY", "astolfo-secret-key-change-in-production")
STORAGE_PATH = Path(os.getenv("STORAGE_PATH", "./storage"))
STORAGE_PATH.mkdir(exist_ok=True)

# Database setup
engine = create_engine(DATABASE_URL)
SessionLocal = sessionmaker(autocommit=False, autoflush=False, bind=engine)
Base = declarative_base()

# Security
pwd_context = CryptContext(schemes=["bcrypt"], deprecated="auto")
security = HTTPBearer()

# Models
class User(Base):
    __tablename__ = "users"
    
    id = Column(String, primary_key=True, default=lambda: str(uuid.uuid4()))
    username = Column(String, unique=True, nullable=False)
    password_hash = Column(String, nullable=False)
    created_at = Column(DateTime, default=datetime.utcnow)
    
    tracks = relationship("Track", back_populates="user", cascade="all, delete-orphan")

class Track(Base):
    __tablename__ = "tracks"
    
    id = Column(String, primary_key=True, default=lambda: str(uuid.uuid4()))
    user_id = Column(String, ForeignKey("users.id"), nullable=False)
    filename = Column(String, nullable=False)
    file_hash = Column(String(64), nullable=False, index=True)
    title = Column(String)
    artist = Column(String)
    album = Column(String)
    duration = Column(Integer)
    file_size = Column(BigInteger)
    mime_type = Column(String)
    uploaded_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)
    
    user = relationship("User", back_populates="tracks")

Base.metadata.create_all(bind=engine)

# Pydantic schemas
class UserCreate(BaseModel):
    username: str
    password: str

class UserLogin(BaseModel):
    username: str
    password: str

class Token(BaseModel):
    access_token: str
    token_type: str

class TrackMetadata(BaseModel):
    id: str
    filename: str
    file_hash: str
    title: Optional[str]
    artist: Optional[str]
    album: Optional[str]
    duration: Optional[int]
    file_size: int
    mime_type: Optional[str]
    uploaded_at: datetime
    updated_at: datetime

class SyncStatus(BaseModel):
    file_hash: str
    filename: str
    uploaded_at: datetime

# FastAPI app
app = FastAPI(title="Astolfo's Player API", description="Music sync API for Astolfo's Player")

# Dependencies
def get_db():
    db = SessionLocal()
    try:
        yield db
    finally:
        db.close()

def create_access_token(user_id: str) -> str:
    expire = datetime.utcnow() + timedelta(days=30)
    to_encode = {"sub": user_id, "exp": expire}
    return jwt.encode(to_encode, SECRET_KEY, algorithm="HS256")

def verify_token(credentials: HTTPAuthorizationCredentials = Depends(security)) -> str:
    try:
        payload = jwt.decode(credentials.credentials, SECRET_KEY, algorithms=["HS256"])
        user_id: str = payload.get("sub")
        if user_id is None:
            raise HTTPException(status_code=401, detail="Invalid token")
        return user_id
    except jwt.PyJWTError:
        raise HTTPException(status_code=401, detail="Invalid token")

def get_current_user(user_id: str = Depends(verify_token), db: Session = Depends(get_db)) -> User:
    user = db.query(User).filter(User.id == user_id).first()
    if not user:
        raise HTTPException(status_code=404, detail="User not found")
    return user

def calculate_file_hash(file_path: Path) -> str:
    """Вычисляет SHA256 хеш файла"""
    sha256_hash = hashlib.sha256()
    with open(file_path, "rb") as f:
        for byte_block in iter(lambda: f.read(4096), b""):
            sha256_hash.update(byte_block)
    return sha256_hash.hexdigest()

# Auth endpoints
@app.post("/api/auth/register", response_model=Token)
def register(user_data: UserCreate, db: Session = Depends(get_db)):
    # Проверка существования пользователя
    if db.query(User).filter(User.username == user_data.username).first():
        raise HTTPException(status_code=400, detail="Username already exists")
    
    # Создание пользователя
    hashed_password = pwd_context.hash(user_data.password)
    user = User(username=user_data.username, password_hash=hashed_password)
    db.add(user)
    db.commit()
    db.refresh(user)
    
    # Создание токена
    token = create_access_token(user.id)
    return {"access_token": token, "token_type": "bearer"}

@app.post("/api/auth/login", response_model=Token)
def login(user_data: UserLogin, db: Session = Depends(get_db)):
    user = db.query(User).filter(User.username == user_data.username).first()
    if not user or not pwd_context.verify(user_data.password, user.password_hash):
        raise HTTPException(status_code=401, detail="Invalid credentials")
    
    token = create_access_token(user.id)
    return {"access_token": token, "token_type": "bearer"}

# Track endpoints
@app.get("/api/tracks", response_model=List[TrackMetadata])
def get_tracks(current_user: User = Depends(get_current_user)):
    return current_user.tracks

@app.get("/api/tracks/{track_id}", response_model=TrackMetadata)
def get_track(track_id: str, current_user: User = Depends(get_current_user), db: Session = Depends(get_db)):
    track = db.query(Track).filter(Track.id == track_id, Track.user_id == current_user.id).first()
    if not track:
        raise HTTPException(status_code=404, detail="Track not found")
    return track

@app.get("/api/tracks/{track_id}/file")
def download_track(track_id: str, current_user: User = Depends(get_current_user), db: Session = Depends(get_db)):
    track = db.query(Track).filter(Track.id == track_id, Track.user_id == current_user.id).first()
    if not track:
        raise HTTPException(status_code=404, detail="Track not found")
    
    file_path = STORAGE_PATH / current_user.id / track.file_hash
    if not file_path.exists():
        raise HTTPException(status_code=404, detail="File not found on server")
    
    # Определяем MIME type, если не был сохранён
    mime_type = track.mime_type
    if not mime_type:
        # Определяем по расширению файла
        ext = track.filename.lower().split('.')[-1]
        mime_types = {
            'mp3': 'audio/mpeg',
            'flac': 'audio/flac',
            'wav': 'audio/wav',
            'ogg': 'audio/ogg',
            'opus': 'audio/opus',
            'm4a': 'audio/mp4',
            'aac': 'audio/aac',
            'wma': 'audio/x-ms-wma',
        }
        mime_type = mime_types.get(ext, 'audio/mpeg')
    
    return FileResponse(
        path=file_path,
        media_type=mime_type,
        filename=track.filename
    )

@app.post("/api/tracks/upload", response_model=TrackMetadata)
async def upload_track(
    file: UploadFile = File(...),
    title: Optional[str] = Form(None),
    artist: Optional[str] = Form(None),
    album: Optional[str] = Form(None),
    duration: Optional[int] = Form(None),
    current_user: User = Depends(get_current_user),
    db: Session = Depends(get_db)
):
    # Создание директории пользователя
    user_storage = STORAGE_PATH / current_user.id
    user_storage.mkdir(exist_ok=True)
    
    # Временное сохранение файла для вычисления хеша
    temp_path = user_storage / f"temp_{uuid.uuid4()}"
    
    # КРИТИЧНО: Читаем весь файл порциями с явным flush
    try:
        with open(temp_path, "wb") as buffer:
            while chunk := await file.read(1024 * 1024):  # 1MB chunks
                buffer.write(chunk)
                buffer.flush()  # Гарантируем запись на диск
            # Финальный flush и fsync для гарантии записи
            buffer.flush()
            os.fsync(buffer.fileno())
    except Exception as e:
        if temp_path.exists():
            temp_path.unlink()
        raise HTTPException(status_code=500, detail=f"File upload failed: {str(e)}")
    
    # Вычисление хеша
    try:
        file_hash = calculate_file_hash(temp_path)
    except Exception as e:
        temp_path.unlink()
        raise HTTPException(status_code=500, detail=f"Hash calculation failed: {str(e)}")
    
    # Проверка существования файла
    existing_track = db.query(Track).filter(
        Track.user_id == current_user.id,
        Track.file_hash == file_hash
    ).first()
    
    if existing_track:
        temp_path.unlink()  # Удаляем временный файл
        return existing_track
    
    # Перемещение файла
    try:
        final_path = user_storage / file_hash
        temp_path.rename(final_path)
        # Проверяем, что файл действительно переместился
        if not final_path.exists():
            raise Exception("File move verification failed")
    except Exception as e:
        if temp_path.exists():
            temp_path.unlink()
        raise HTTPException(status_code=500, detail=f"File move failed: {str(e)}")
    
    # Создание записи в БД
    try:
        track = Track(
            user_id=current_user.id,
            filename=file.filename,
            file_hash=file_hash,
            title=title,
            artist=artist,
            album=album,
            duration=duration,
            file_size=final_path.stat().st_size,
            mime_type=file.content_type
        )
        db.add(track)
        db.commit()
        db.refresh(track)
    except Exception as e:
        # Откатываем изменения
        if final_path.exists():
            final_path.unlink()
        db.rollback()
        raise HTTPException(status_code=500, detail=f"Database error: {str(e)}")
    
    return track

@app.delete("/api/tracks/{track_id}")
def delete_track(track_id: str, current_user: User = Depends(get_current_user), db: Session = Depends(get_db)):
    track = db.query(Track).filter(Track.id == track_id, Track.user_id == current_user.id).first()
    if not track:
        raise HTTPException(status_code=404, detail="Track not found")
    
    # Удаление файла
    file_path = STORAGE_PATH / current_user.id / track.file_hash
    if file_path.exists():
        file_path.unlink()
    
    # Удаление записи из БД
    db.delete(track)
    db.commit()
    
    return {"message": "Track deleted successfully"}

# Sync endpoints
@app.get("/api/sync/status", response_model=List[SyncStatus])
def get_sync_status(current_user: User = Depends(get_current_user)):
    """Возвращает список всех хешей файлов для сравнения"""
    return [
        SyncStatus(
            file_hash=track.file_hash,
            filename=track.filename,
            uploaded_at=track.uploaded_at
        )
        for track in current_user.tracks
    ]

@app.post("/api/sync/batch-upload")
async def batch_upload(
    files: List[UploadFile] = File(...),
    current_user: User = Depends(get_current_user),
    db: Session = Depends(get_db)
):
    """Массовая загрузка файлов"""
    results = []
    
    for file in files:
        try:
            # Создание директории пользователя
            user_storage = STORAGE_PATH / current_user.id
            user_storage.mkdir(exist_ok=True)
            
            # Временное сохранение
            temp_path = user_storage / f"temp_{uuid.uuid4()}"
            with open(temp_path, "wb") as buffer:
                shutil.copyfileobj(file.file, buffer)
            
            # Вычисление хеша
            file_hash = calculate_file_hash(temp_path)
            
            # Проверка существования
            existing_track = db.query(Track).filter(
                Track.user_id == current_user.id,
                Track.file_hash == file_hash
            ).first()
            
            if existing_track:
                temp_path.unlink()
                results.append({"filename": file.filename, "status": "exists", "track_id": existing_track.id})
                continue
            
            # Перемещение и сохранение
            final_path = user_storage / file_hash
            temp_path.rename(final_path)
            
            track = Track(
                user_id=current_user.id,
                filename=file.filename,
                file_hash=file_hash,
                file_size=final_path.stat().st_size,
                mime_type=file.content_type
            )
            db.add(track)
            db.commit()
            db.refresh(track)
            
            results.append({"filename": file.filename, "status": "uploaded", "track_id": track.id})
        except Exception as e:
            results.append({"filename": file.filename, "status": "error", "error": str(e)})
    
    return {"results": results}

@app.get("/")
def root():
    return {"message": "Astolfo's Player API", "version": "1.0.0"}

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)