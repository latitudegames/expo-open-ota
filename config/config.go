package config

import (
	"encoding/json"
	"expo-open-ota/internal/helpers"
	"flag"
	"log"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/joho/godotenv"
)

type AppConfig struct {
	ChannelPrefix string `json:"channelPrefix"`
	AppId         string `json:"appId"`
	AccessToken   string `json:"accessToken,omitempty"`
}

var (
	appConfigs     []AppConfig
	appConfigsOnce sync.Once
)

func loadAppConfigs() {
	raw := os.Getenv("EXPO_APP_CONFIGS")
	if raw == "" {
		return
	}
	var configs []AppConfig
	if err := json.Unmarshal([]byte(raw), &configs); err != nil {
		log.Printf("WARN: EXPO_APP_CONFIGS is not valid JSON, ignoring: %v", err)
		return
	}
	sort.Slice(configs, func(i, j int) bool {
		return len(configs[i].ChannelPrefix) > len(configs[j].ChannelPrefix)
	})
	appConfigs = configs
	log.Printf("Loaded %d app config(s) from EXPO_APP_CONFIGS", len(configs))
}

func ResolveAppForChannel(channelName string) (appId string, accessToken string) {
	appConfigsOnce.Do(loadAppConfigs)
	for _, cfg := range appConfigs {
		if cfg.ChannelPrefix != "" && strings.HasPrefix(channelName, cfg.ChannelPrefix) {
			token := cfg.AccessToken
			if token == "" {
				token = GetEnv("EXPO_ACCESS_TOKEN")
			}
			return cfg.AppId, token
		}
	}
	return GetEnv("EXPO_APP_ID"), GetEnv("EXPO_ACCESS_TOKEN")
}

func validateStorageMode(storageMode string) bool {
	return storageMode == "local" || storageMode == "s3"
}

func GetPort() string {
	port := GetEnv("PORT")
	if port == "" {
		port = "3000"
	}
	return port
}

func validateBucketParams(storageMode string) bool {
	switch storageMode {
	case "s3":
		bucketName := GetEnv("S3_BUCKET_NAME")
		if bucketName == "" {
			log.Printf("S3_BUCKET_NAME not set")
			return false
		}
		region := GetEnv("AWS_REGION")
		if region == "" {
			log.Printf("AWS_REGION not set")
			return false
		}
	case "local":
		// Already handled by default values
		return true
	default:
		return false
	}
	return true
}

func validateBaseUrl(baseUrl string) bool {
	return baseUrl != "" && helpers.IsValidURL(baseUrl)
}

func IsTestMode() bool {
	return flag.Lookup("test.v") != nil
}

func resolveDefaultBaseUrl() string {
	port := os.Getenv("PORT")
	if port == "" {
		return "http://localhost:3000"
	}
	return "http://localhost:" + port
}


func LoadConfig() {
	err := godotenv.Load()
	if err != nil {
		log.Printf("No .env file found, continuing with runtime environment variables.")
	}
	storageMode := GetEnv("STORAGE_MODE")
	if !validateStorageMode(storageMode) {
		log.Fatalf("Invalid STORAGE_MODE: %s", storageMode)
	}
	bucketParamsValid := validateBucketParams(storageMode)
	if !bucketParamsValid {
		log.Fatalf("Invalid bucket parameters")
	}
	baseUrl := GetEnv("BASE_URL")
	if !validateBaseUrl(baseUrl) {
		log.Fatalf("Invalid BASE_URL: %s", baseUrl)
	}
	expoToken := GetEnv("EXPO_ACCESS_TOKEN")
	if expoToken == "" {
		log.Fatalf("EXPO_ACCESS_TOKEN not set")
	}
	expoAppId := GetEnv("EXPO_APP_ID")
	if expoAppId == "" {
		log.Fatalf("EXPO_APP_ID not set")
	}
	jwtSecret := GetEnv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatalf("JWT_SECRET not set")
	}
}

var DefaultEnvValues = map[string]string{
	"LOCAL_BUCKET_BASE_PATH":      "./updates",
	"STORAGE_MODE":                "local",
	"BASE_URL":                    resolveDefaultBaseUrl(),
	"PUBLIC_LOCAL_EXPO_KEY_PATH":  "./keyStore/public-key.pem",
	"PRIVATE_LOCAL_EXPO_KEY_PATH": "./keyStore/private-key.pem",
	"KEYS_STORAGE_TYPE":           "local",
	"JWT_SECRET":                  "",
	"AWS_REGION":                  "eu-west-3",
}


func GetEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		defaultValue := DefaultEnvValues[key]
		if defaultValue != "" {
			return defaultValue
		}
		return ""
	}
	return value
}
