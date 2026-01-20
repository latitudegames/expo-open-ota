# Expo Open OTA
![Expo Open OTA Deployment](docs/static/img/social_card.png)

[![Push workflow](https://github.com/axelmarciano/expo-open-ota/actions/workflows/push.yml/badge.svg)](https://github.com/axelmarciano/expo-open-ota/actions/workflows/push.yml)

🚀 **An open-source Go implementation of the Expo Updates protocol, designed for production with support for cloud storage like S3 and CDN integration, delivering fast and reliable OTA updates for React Native apps.**

## ⚠️ Disclaimer

**Expo Open OTA is not officially supported or affiliated with [Expo](https://expo.dev/).**  
This is an independent open-source project.

## 📖 Documentation

The full documentation is available at:  
➡️ [Documentation](https://axelmarciano.github.io/expo-open-ota/)

## 🛠 Features

- **Self-hosted OTA update server** for Expo applications.
- **Cloud storage support**: AWS S3, local storage, and more.
- **CDN integration**: Optimized for CloudFront and other CDN providers.
- **Secure key management**: Supports AWS Secrets Manager and environment-based key storage.
- **Production-ready**: Designed for scalability and performance.

## 🔨 Building

### Prerequisites

- Go 1.23+
- Node.js 18+
- npm

### Build Everything

All commands assume you start from the **repository root directory** (`expo-open-ota/`).

**1. Build the dashboard (frontend):**

```bash
# From: expo-open-ota/
cd dashboard
npm ci
npm run build
```

**2. Build the Go backend:**

```bash
# From: expo-open-ota/ (return to root if you were in dashboard/)
cd ..
go build -o main ./cmd/api
```

### Build with Docker

```bash
# From: expo-open-ota/ (repository root)
docker build -t expo-open-ota .
```

### Run Locally

```bash
# From: expo-open-ota/ (repository root)

# Run the built binary
./main

# Or run directly with Go
go run ./cmd/api

# Or run with Docker
docker run -p 3000:3000 expo-open-ota
```

## 🚀 Deploy to Heroku

This project is deployed to Heroku as **latitude-updates** using Docker containers.

All Heroku commands should be run from the **repository root directory** (`expo-open-ota/`).

### Initial Setup (already configured)

The Heroku remote is already configured for this repo:

```bash
# Heroku Git URL: https://git.heroku.com/latitude-updates.git
```

If you need to set it up again:

```bash
# From: expo-open-ota/ (repository root)

# Login to Heroku CLI
heroku login

# Add Heroku remote to your git repo
git remote add heroku https://git.heroku.com/latitude-updates.git

# Verify the stack is set to container
heroku stack:set container --app latitude-updates
```

### Configure Environment Variables

Set your required environment variables on Heroku (can be run from any directory):

```bash
heroku config:set BASE_URL=https://latitude-updates.herokuapp.com --app latitude-updates
heroku config:set EXPO_APP_ID=your-expo-app-id --app latitude-updates
heroku config:set EXPO_ACCESS_TOKEN=your-expo-access-token --app latitude-updates
heroku config:set USE_DASHBOARD=true --app latitude-updates
heroku config:set JWT_SECRET=your-secret-key --app latitude-updates
# Add other environment variables as needed (see Documentation for full list)
```

### Deploy

```bash
# From: expo-open-ota/ (repository root)

# Push to Heroku (builds and deploys automatically)
git push heroku main
```

### Useful Heroku Commands

These commands can be run from any directory:

```bash
# View logs
heroku logs --tail --app latitude-updates

# Check app status
heroku ps --app latitude-updates

# Restart the app
heroku restart --app latitude-updates

# Open the app in browser
heroku open --app latitude-updates
```

## 📜 License

This project is licensed under the MIT License - see the [LICENSE](./LICENSE.md) file for details.

## Contact

✉️ [E-mail](mailto:expoopenota@gmail.com)