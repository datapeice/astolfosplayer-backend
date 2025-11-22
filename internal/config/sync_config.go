package config

type SyncConfig struct {
	DatabaseURL string
	Port        string
}

func LoadSyncConfig() *SyncConfig {
	return &SyncConfig{
		DatabaseURL: getEnv("DATABASE_URL", "metadata.db"),
		Port:        getEnv("PORT", "50053"),
	}
}
