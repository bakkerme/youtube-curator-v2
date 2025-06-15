package store

import (
	"testing"
	"time"

	"youtube-curator-v2/internal/rss"
)

func TestVideoStore_WatchedStatePersistence_AcrossRestarts(t *testing.T) {
	// Create a temporary database for testing
	tmpDir := t.TempDir()
	dbPath := tmpDir + "/test.db"

	// Create first database instance
	db1, err := NewStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create first database instance: %v", err)
	}

	// Create first video store
	videoStore1 := NewVideoStoreWithStore(1*time.Hour, db1)

	// Add a test video
	testVideo := rss.Entry{
		ID:        "yt:video:test_persistence_video",
		Title:     "Test Persistence Video",
		Published: time.Now().Add(-2 * time.Hour),
	}
	videoStore1.AddVideo("channel1", testVideo)

	// Verify video is initially unwatched
	videos := videoStore1.GetAllVideos()
	if len(videos) != 1 {
		t.Fatalf("Expected 1 video in store, got %d", len(videos))
	}
	if videos[0].Watched {
		t.Fatalf("Expected video to be initially unwatched")
	}

	// Mark video as watched
	videoStore1.MarkVideoAsWatched("yt:video:test_persistence_video")

	// Verify video is now watched in memory
	videos = videoStore1.GetAllVideos()
	if !videos[0].Watched {
		t.Fatalf("Expected video to be marked as watched in memory")
	}

	// Verify watched state is persisted in database
	isWatched, err := db1.IsVideoWatched("yt:video:test_persistence_video")
	if err != nil {
		t.Fatalf("Error checking watched state in database: %v", err)
	}
	if !isWatched {
		t.Fatalf("Expected video to be marked as watched in database")
	}

	// Close first database and video store (simulating container shutdown)
	db1.Close()

	// Create second database instance with same path (simulating container restart)
	db2, err := NewStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create second database instance: %v", err)
	}
	defer db2.Close()

	// Create second video store (simulating new container startup)
	videoStore2 := NewVideoStoreWithStore(1*time.Hour, db2)

	// Add the same video again (this would happen when RSS feeds are refreshed on restart)
	videoStore2.AddVideo("channel1", testVideo)

	// Verify watched state is preserved after "container restart"
	videosAfterRestart := videoStore2.GetAllVideos()
	if len(videosAfterRestart) != 1 {
		t.Fatalf("Expected 1 video in store after restart, got %d", len(videosAfterRestart))
	}

	if videosAfterRestart[0].Entry.ID != "yt:video:test_persistence_video" {
		t.Errorf("Expected video ID 'yt:video:test_persistence_video', got '%s'", videosAfterRestart[0].Entry.ID)
	}

	// This is the key test - watched state should survive the "container restart"
	if !videosAfterRestart[0].Watched {
		t.Error("Expected video to remain marked as watched after container restart simulation")
	}
}

func TestVideoStore_MultiplePersistentWatchedVideos(t *testing.T) {
	// Create a temporary database for testing
	tmpDir := t.TempDir()
	dbPath := tmpDir + "/test_multiple.db"

	// Create database instance
	db, err := NewStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database instance: %v", err)
	}
	defer db.Close()

	// Create video store
	videoStore := NewVideoStoreWithStore(1*time.Hour, db)

	// Add multiple test videos
	testVideo1 := rss.Entry{
		ID:        "yt:video:test_video_1",
		Title:     "Test Video 1",
		Published: time.Now().Add(-3 * time.Hour),
	}
	testVideo2 := rss.Entry{
		ID:        "yt:video:test_video_2",
		Title:     "Test Video 2",
		Published: time.Now().Add(-2 * time.Hour),
	}
	testVideo3 := rss.Entry{
		ID:        "yt:video:test_video_3",
		Title:     "Test Video 3",
		Published: time.Now().Add(-1 * time.Hour),
	}

	videoStore.AddVideo("channel1", testVideo1)
	videoStore.AddVideo("channel1", testVideo2)
	videoStore.AddVideo("channel2", testVideo3)

	// Mark some videos as watched
	videoStore.MarkVideoAsWatched("yt:video:test_video_1")
	videoStore.MarkVideoAsWatched("yt:video:test_video_3")

	// Verify in-memory state
	videos := videoStore.GetAllVideos()
	if len(videos) != 3 {
		t.Fatalf("Expected 3 videos in store, got %d", len(videos))
	}

	watchedCount := 0
	for _, video := range videos {
		if video.Watched {
			watchedCount++
		}
	}
	if watchedCount != 2 {
		t.Fatalf("Expected 2 watched videos in memory, got %d", watchedCount)
	}

	// Verify persistent state
	watchedVideos, err := db.GetWatchedVideos()
	if err != nil {
		t.Fatalf("Error getting watched videos from database: %v", err)
	}
	if len(watchedVideos) != 2 {
		t.Fatalf("Expected 2 watched videos in database, got %d", len(watchedVideos))
	}

	// Check individual videos
	isWatched1, _ := db.IsVideoWatched("yt:video:test_video_1")
	isWatched2, _ := db.IsVideoWatched("yt:video:test_video_2")
	isWatched3, _ := db.IsVideoWatched("yt:video:test_video_3")

	if !isWatched1 {
		t.Error("Expected video 1 to be watched")
	}
	if isWatched2 {
		t.Error("Expected video 2 to be unwatched")
	}
	if !isWatched3 {
		t.Error("Expected video 3 to be watched")
	}
}

func TestVideoStore_WatchedStateWithoutPersistentStore(t *testing.T) {
	// Test that VideoStore still works when no persistent store is provided
	videoStore := NewVideoStore(1 * time.Hour) // Old constructor without store

	testVideo := rss.Entry{
		ID:        "yt:video:test_no_persistence",
		Title:     "Test No Persistence Video",
		Published: time.Now().Add(-1 * time.Hour),
	}

	videoStore.AddVideo("channel1", testVideo)
	videoStore.MarkVideoAsWatched("yt:video:test_no_persistence")

	// Should work in memory even without persistence
	videos := videoStore.GetAllVideos()
	if len(videos) != 1 {
		t.Fatalf("Expected 1 video, got %d", len(videos))
	}
	if !videos[0].Watched {
		t.Error("Expected video to be watched in memory")
	}
}