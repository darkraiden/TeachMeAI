package config

import "os"

type Config struct {
	OllamaHost string
	MongoURI   string
	ServerPort string
}

func Load() *Config {
	cfg := &Config{
		OllamaHost: getEnv("OLLAMA_HOST", "http://localhost:11434"),
		MongoURI:   getEnv("MONGO_URI", "mongodb://localhost:27017/teachme"),
		ServerPort: getEnv("PORT", "8080"),
	}
	return cfg
}

func getEnv(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}
