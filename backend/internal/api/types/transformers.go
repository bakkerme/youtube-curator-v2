package types

import (
	"time"
	"youtube-curator-v2/internal/rss"
	"youtube-curator-v2/internal/store"
)

// transformChannel converts a store.Channel to ChannelResponse
func TransformChannel(channel store.Channel) ChannelResponse {
	return ChannelResponse{
		ID:         channel.ID,
		Title:      channel.Title,
		CreatedAt:  time.Now(), // TODO: Add CreatedAt to store.Channel if needed
		IsActive:   true,       // TODO: Add IsActive to store.Channel if needed
		VideoCount: 0,          // TODO: Calculate video count if needed
	}
}

// transformChannels converts a slice of store.Channel to ChannelsResponse
func TransformChannels(channels []store.Channel) ChannelsResponse {
	channelResponses := make([]ChannelResponse, len(channels))
	for i, channel := range channels {
		channelResponses[i] = TransformChannel(channel)
	}

	return ChannelsResponse{
		Channels:    channelResponses,
		TotalCount:  len(channels),
		LastUpdated: time.Now(),
	}
}

// transformVideoEntry converts a store.VideoEntry to VideoResponse matching the frontend VideoEntry interface
func TransformVideoEntry(videoEntry store.VideoEntry) VideoResponse {
	entry := videoEntry.Entry

	return VideoResponse{
		ID:         entry.ID,
		ChannelID:  videoEntry.ChannelID,
		CachedAt:   videoEntry.CachedAt,
		Watched:    videoEntry.Watched,
		Title:      entry.Title,
		Link:       TransformVideoLink(entry.Link),
		Published:  entry.Published,
		Content:    entry.Content,
		Author:     TransformVideoAuthor(entry.Author),
		MediaGroup: TransformVideoMediaGroup(entry.MediaGroup),
	}
}

// transformVideoLink converts rss.Link to VideoLinkResponse
func TransformVideoLink(link rss.Link) VideoLinkResponse {
	return VideoLinkResponse{
		Href: link.Href,
		Rel:  link.Rel,
	}
}

// transformVideoAuthor converts rss.Author to VideoAuthorResponse
func TransformVideoAuthor(author rss.Author) VideoAuthorResponse {
	return VideoAuthorResponse{
		Name: author.Name,
		URI:  author.URI,
	}
}

// transformVideoMediaGroup converts rss.MediaGroup to VideoMediaGroupResponse
func TransformVideoMediaGroup(mediaGroup rss.MediaGroup) VideoMediaGroupResponse {
	return VideoMediaGroupResponse{
		MediaThumbnail:   TransformVideoMediaThumbnail(mediaGroup.MediaThumbnail),
		MediaTitle:       mediaGroup.MediaTitle,
		MediaContent:     TransformVideoMediaContent(mediaGroup.MediaContent),
		MediaDescription: mediaGroup.MediaDescription,
	}
}

// transformVideoMediaThumbnail converts rss.MediaThumbnail to VideoMediaThumbnailResponse
func TransformVideoMediaThumbnail(thumbnail rss.MediaThumbnail) VideoMediaThumbnailResponse {
	return VideoMediaThumbnailResponse{
		URL:    thumbnail.URL,
		Width:  thumbnail.Width,
		Height: thumbnail.Height,
	}
}

// transformVideoMediaContent converts rss.MediaContent to VideoMediaContentResponse
func TransformVideoMediaContent(content rss.MediaContent) VideoMediaContentResponse {
	return VideoMediaContentResponse{
		URL:    content.URL,
		Type:   content.Type,
		Width:  content.Width,
		Height: content.Height,
	}
}

// transformVideos converts video entries to VideosResponse
func TransformVideos(videoEntries []store.VideoEntry, lastRefresh time.Time) VideosResponse {
	videoResponses := make([]VideoResponse, len(videoEntries))
	for i, videoEntry := range videoEntries {
		videoResponses[i] = TransformVideoEntry(videoEntry)
	}

	return VideosResponse{
		Videos:      videoResponses,
		TotalCount:  len(videoEntries),
		LastRefresh: lastRefresh,
	}
}
