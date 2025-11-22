# Astolfo's Player Backend (Go/gRPC)

A microservices-based backend for synchronizing music files between devices, built with Go and gRPC.

## Features

✅ **Microservices Architecture**: Auth, File, and Sync services  
✅ **User Authentication**: JWT-based with security key protection  
✅ **File Storage**: MinIO (S3-compatible) for large music files  
✅ **File Synchronization**: Hash-based deduplication using SHA256  
✅ **Streaming**: Efficient file upload/download with gRPC streams  
✅ **Lightweight**: SQLite for metadata (Raspberry Pi friendly)  
✅ **Containerized**: Docker Compose and Kubernetes ready  
✅ **Format Support**: MP3, FLAC, WAV, OGG, Opus, M4A, AAC, WMA

## Quick Start

### 1. Clone the Repository

```bash
git clone https://github.com/datapeice/astolfosplayer-backend.git
cd astolfosplayer-backend
```

### 2. Start with Docker Compose

```bash
docker-compose up -d
```

Services will be available at:
- **Auth Service**: `localhost:50051`
- **File Service**: `localhost:50052`
- **Sync Service**: `localhost:50053`
- **MinIO Console**: `http://localhost:9001` (admin/minioadmin)

## Architecture

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Auth      │     │    File     │     │    Sync     │
│  Service    │     │  Service    │     │  Service    │
│  :50051     │     │  :50052     │     │  :50053     │
└──────┬──────┘     └──────┬──────┘     └──────┬──────┘
       │                   │                   │
       ├───────────────────┴───────────────────┤
       │            SQLite (Metadata)          │
       └───────────────────────────────────────┘
                          │
                  ┌───────┴────────┐
                  │   MinIO (S3)   │
                  │  Music Files   │
                  └────────────────┘
```

## API Services

### Auth Service (Port 50051)

- `Register(username, password, security_key)` → `token`
- `Login(username, password)` → `token`

### File Service (Port 50052)

- `Upload(stream)` → `hash` (Client streaming)
- `Download(hash)` → `stream` (Server streaming)
- `Delete(hash)` → `success`

### Sync Service (Port 50053)

- `GetSync()` → `[hashes]`

## Development

### Prerequisites

- Go 1.24+
- Docker & Docker Compose
- `protoc` compiler

### Build Locally

```bash
# Install dependencies
go mod download

# Generate proto files
cd protos
protoc -I proto --go_out=gen/go --go_opt=paths=source_relative \
  --go-grpc_out=gen/go --go-grpc_opt=paths=source_relative \
  proto/auth/auth.proto proto/file/file.proto proto/sync/sync.proto

# Build services
go build -o bin/auth-service cmd/auth/main.go
go build -o bin/file-service cmd/file/main.go
go build -o bin/sync-service cmd/sync/main.go

# Run
./bin/auth-service
```

### Environment Variables

#### Auth Service
- `DATABASE_URL`: SQLite database path (default: `auth.db`)
- `SECRET_KEY`: JWT signing key
- `SECURITY_KEY`: Required key for registration
- `PORT`: gRPC port (default: `50051`)

#### File Service
- `DATABASE_URL`: SQLite database path (default: `metadata.db`)
- `S3_ENDPOINT`: MinIO endpoint (default: `localhost:9000`)
- `S3_ACCESS_KEY`: MinIO access key
- `S3_SECRET_KEY`: MinIO secret key
- `S3_BUCKET`: Bucket name (default: `music`)
- `S3_USE_SSL`: Use SSL for S3 (default: `false`)
- `PORT`: gRPC port (default: `50052`)

#### Sync Service
- `DATABASE_URL`: SQLite database path (default: `metadata.db`)
- `PORT`: gRPC port (default: `50053`)

## Deployment

### Docker Compose (Recommended)

See [`docker-compose.yml`](docker-compose.yml) for the full configuration.

### Kubernetes

See [`docs/kubernetes_deployment.md`](docs/kubernetes_deployment.md) for detailed Kubernetes deployment instructions.

## Client Migration

For Android developers using Ktor, see [`docs/android_ktor_migration.md`](docs/android_ktor_migration.md) for migration guide from REST to gRPC.

## Testing

### Using grpcurl

Install [`grpcurl`](https://github.com/fullstorydev/grpcurl):

```bash
# Register a user
grpcurl -plaintext -d '{\"username\":\"test\",\"password\":\"password\",\"security_key\":\"dev-security-key\"}' \
  localhost:50051 auth.AuthService/Register

# Login
grpcurl -plaintext -d '{\"username\":\"test\",\"password\":\"password\"}' \
  localhost:50051 auth.AuthService/Login

# Get sync hashes
grpcurl -plaintext localhost:50053 sync.SyncService/GetSync
```

## Project Structure

```
astolfosplayer-backend/
├── cmd/                    # Service entrypoints
│   ├── auth/
│   ├── file/
│   └── sync/
├── internal/               # Internal packages
│   ├── auth/              # Auth service logic
│   ├── file/              # File service logic
│   ├── sync/              # Sync service logic
│   ├── config/            # Configuration
│   └── db/                # Database connection
├── protos/                # Protocol Buffers
│   ├── proto/             # .proto definitions
│   └── gen/               # Generated code
├── build/package/         # Dockerfiles
├── docs/                  # Documentation
├── docker-compose.yml
└── README.md
```

## Security

⚠️ **Important for Production:**

1. Change `SECRET_KEY` and `SECURITY_KEY` to strong random values
2. Enable TLS for gRPC (replace `insecure` credentials)
3. Use S3 with SSL (`S3_USE_SSL=true`)
4. Secure MinIO with strong credentials
5. Use Kubernetes Secrets for sensitive data
6. Regular backups of SQLite and S3 data

## License

MIT
