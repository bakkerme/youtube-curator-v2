# yt-dlp Caching

The yt-dlp enricher now includes a file-based caching layer to speed up development and reduce API calls to YouTube.

## Configuration

### Environment Variables

- `YTDLP_CACHE_DIR`: Directory to store cache files (default: `./cache/ytdlp`)
- `YTDLP_DISABLE_CACHE`: Set to `"true"` to disable caching (default: enabled)

### Examples

```bash
# Use custom cache directory
export YTDLP_CACHE_DIR="/tmp/my-ytdlp-cache"

# Disable caching completely
export YTDLP_DISABLE_CACHE="true"

# Default behavior (caching enabled in ./cache/ytdlp)
# No environment variables needed
```

## How It Works

1. **Cache Hit**: If video data exists in cache, it's loaded instantly without calling yt-dlp
2. **Cache Miss**: Video data is fetched from YouTube via yt-dlp and cached for future use
3. **Cache Storage**: JSON responses are stored as hashed filenames (e.g., `a1b2c3d4e5f6g7h8.json`)
4. **Development**: Greatly speeds up repeated testing with the same videos

## Benefits

- **Faster Development**: No need to wait for yt-dlp on repeated requests
- **Reduced API Load**: Fewer calls to YouTube's servers
- **Offline Testing**: Can test with cached data even without internet
- **Cost Effective**: Reduces load on YouTube's APIs

## Cache Management

The cache automatically:
- Creates the cache directory on startup
- Handles corrupted cache files gracefully
- Provides logging for cache hits/misses
- Includes `ClearCache()` method for cleanup

## File Structure

```
./cache/ytdlp/
├── a1b2c3d4e5f6g7h8.json  # Cached yt-dlp response for video 1
├── b2c3d4e5f6g7h8i9.json  # Cached yt-dlp response for video 2
└── ...
```

Cache files contain the raw JSON response from yt-dlp, including:
- Video duration, tags, comments
- Subtitle URLs
- All other metadata returned by yt-dlp