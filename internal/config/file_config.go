package config

type FileConfig struct {
	S3Endpoint  string
	S3AccessKey string
	S3SecretKey string
	S3Bucket    string
	S3UseSSL    bool
	DatabaseURL string
	Port        string
}

func LoadFileConfig() *FileConfig {
	return &FileConfig{
		S3Endpoint:  getEnv("S3_ENDPOINT", "localhost:9000"),
		S3AccessKey: getEnv("S3_ACCESS_KEY", "minioadmin"),
		S3SecretKey: getEnv("S3_SECRET_KEY", "minioadmin"),
		S3Bucket:    getEnv("S3_BUCKET", "music"),
		S3UseSSL:    getEnv("S3_USE_SSL", "false") == "true",
		DatabaseURL: getEnv("DATABASE_URL", "metadata.db"),
		Port:        getEnv("PORT", "50052"),
	}
}
