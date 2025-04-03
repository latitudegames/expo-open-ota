package config

import (
	"expo-open-ota/internal/helpers"
	"flag"
	"github.com/joho/godotenv"
	"log"
	"os"
)

func validateStorageMode(storageMode string) bool {
	return storageMode == "local" || storageMode == "s3"
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

func LoadConfig() {
	err := godotenv.Load()
	if err != nil {
		log.Printf("No .env file found, continuing with runtime environment variables.")
	}
	
	// Set default BASE_URL if not set and we're in Heroku
	if os.Getenv("BASE_URL") == "" && os.Getenv("PORT") != "" {
		DefaultEnvValues["BASE_URL"] = "http://localhost:" + os.Getenv("PORT")
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
	"BASE_URL":                    "http://localhost:3000",
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
