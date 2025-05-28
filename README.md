# YouTube Curator v2

A self-hosted application that monitors YouTube channels via RSS feeds and sends email notifications when new videos are published.

## Features

- Monitor multiple YouTube channels for new videos
- Send email notifications with video details and thumbnails
- Persistent storage to track last checked videos
- Configurable check intervals
- Docker support for easy deployment
- Debug mode with mock RSS feeds

## Quick Start with Docker

### Prerequisites

- Docker and Docker Compose installed
- SMTP email credentials (Gmail, Outlook, etc.)
- YouTube channel IDs you want to monitor

### Setup

1. **Clone the repository:**
   ```bash
   git clone <repository-url>
   cd youtube-curator-v2
   ```

2. **Validate Docker setup:**
   ```bash
   ./scripts/validate-docker.sh
   ```
   This script will check your Docker installation and create necessary directories.

3. **Configure channels:**
   Edit `channels.txt` and add your YouTube channel IDs (one per line):
   ```
   # Some Channel
   UCAYF6ZY9gWBR1GW3R7PX7yw
   # Another Channel
   UCxxxxxxxxxxxxxxxxxx
   ```

4. **Create environment file:**
   ```bash
   cp env.example .env
   ```
   Then edit `.env` with your SMTP settings:
   ```bash
   # SMTP Configuration (Required)
   SMTP_SERVER=smtp.gmail.com
   SMTP_PORT=587
   SMTP_USERNAME=your-email@gmail.com
   SMTP_PASSWORD=your-app-password
   RECIPIENT_EMAIL=recipient@example.com
   
   # Optional Configuration
   CHECK_INTERVAL=1h
   DEBUG_MOCK_RSS=false
   DEBUG_SKIP_CRON=false
   ```

5. **Build and run:**
   ```bash
   docker compose up -d
   ```

### Docker Commands

**Build the image:**
```bash
docker build -t youtube-curator-v2 .
```

**Run with Docker Compose (recommended):**
```bash
# Start the service
docker compose up -d

# View logs
docker compose logs -f

# Stop the service
docker compose down
```

**Run with Docker directly:**
```bash
docker run -d \
  --name youtube-curator-v2 \
  -v $(pwd)/data/youtubecurator.db:/app/youtubecurator.db \
  -v $(pwd)/channels.txt:/app/channels.txt:ro \
  -e CHANNELS_FILE=/app/channels.txt \
  -e SMTP_SERVER=smtp.gmail.com \
  -e SMTP_PORT=587 \
  -e SMTP_USERNAME=your-email@gmail.com \
  -e SMTP_PASSWORD=your-app-password \
  -e RECIPIENT_EMAIL=recipient@example.com \
  youtube-curator-v2
```

### Using Make Commands

For convenience, you can use the provided Makefile:

```bash
# Setup project (creates directories, copies env template)
make setup

# Validate Docker configuration
make validate

# Build and start with Docker Compose
make docker-up

# View logs
make docker-logs

# Stop the service
make docker-down

# Restart the service
make docker-restart
```

## Configuration

### Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `CHANNELS_FILE` | Yes | - | Path to file containing channel IDs |
| `SMTP_SERVER` | Yes | - | SMTP server hostname |
| `SMTP_PORT` | Yes | - | SMTP server port |
| `SMTP_USERNAME` | Yes | - | SMTP username |
| `SMTP_PASSWORD` | Yes | - | SMTP password |
| `RECIPIENT_EMAIL` | Yes | - | Email address to send notifications to |
| `CHECK_INTERVAL` | No | `1h` | How often to check for new videos |
| `DEBUG_MOCK_RSS` | No | `false` | Use mock RSS feeds for testing |
| `DEBUG_SKIP_CRON` | No | `false` | Run once and exit (no scheduling) |

### Getting YouTube Channel IDs

1. Go to the YouTube channel page
2. View page source and search for `"channelId":"` or `"externalId":"`
3. Copy the ID that looks like `UCxxxxxxxxxxxxxxxxxx`

Alternatively, use online tools or browser extensions to extract channel IDs.

### SMTP Setup Examples

**Gmail:**
- Server: `smtp.gmail.com`
- Port: `587`
- Use an App Password (not your regular password)

**Outlook/Hotmail:**
- Server: `smtp-mail.outlook.com`
- Port: `587`

## Data Persistence

The application uses BadgerDB for local storage. When running with Docker:

- Database files are stored in `./data/youtubecurator.db/` on the host
- This directory is mounted as a volume to persist data between container restarts
- Mock RSS feeds (if used) are stored in `./data/feed_mocks/`

## Troubleshooting

**Validate setup:**
```bash
./scripts/validate-docker.sh
```

**Check logs:**
```bash
docker compose logs -f youtube-curator
```

**Common issues:**
- SMTP authentication errors: Verify credentials and enable "Less secure app access" or use App Passwords
- Channel not found: Verify the channel ID is correct
- Permission errors: Ensure the data directory is writable
- Docker permission denied: Add your user to the docker group or use sudo

**Debug mode:**
Set `DEBUG_SKIP_CRON=true` to run once and exit, useful for testing configuration.

## Development

### Building from Source

```bash
# Install dependencies
go mod download

# Run locally
go run main.go

# Build binary
go build -o youtube-curator-v2
```

### Running Tests

```bash
go test ./...
```