package rss

import (
	"encoding/xml"
	"fmt"
	"strings"
	"time"
)

// Feedlike is an interface that can be used to represent any type that has a FeedString method, i.e. Feed and CommentFeed
type Feedlike interface {
	FeedString() string
}

// Feed is used as an intermediate type for RSS feeds
type Feed struct {
	XMLName     xml.Name `xml:"feed"` // Add XMLName to specify the root element name
	Title       string   `xml:"title"`
	URL         Link     `xml:"-"`        // Will be populated by custom unmarshaler
	Description string   `xml:"subtitle"` // RSS feeds often use 'subtitle' for description
	Entries     []Entry  `xml:"entry"`
	RawRSS      string   // Added field to store raw RSS data
}

func (f *Feed) FeedString() string {
	return f.RawRSS // Method to return the raw RSS data
}

// Entry is used throughout the codebase for RSS feeds
type Entry struct {
	Title      string     `xml:"title" json:"title"`
	Link       Link       `xml:"link" json:"link"`
	ID         string     `xml:"id" json:"id"`
	Published  time.Time  `xml:"published" json:"published"`
	Content    string     `xml:"content" json:"content"`
	Author     Author     `xml:"author" json:"author"`
	MediaGroup MediaGroup `xml:"http://search.yahoo.com/mrss/ group" json:"mediaGroup"` // Field to store media group information

	// Enhanced metadata from yt-dlp (optional fields)
	Duration      int      `json:"duration,omitempty"`      // Duration in seconds
	Tags          []string `json:"tags,omitempty"`          // Video tags
	TopComments   []string `json:"topComments,omitempty"`   // Top comments
	AutoSubtitles string   `json:"autoSubtitles,omitempty"` // Auto-generated English subtitles

	// Video summary information (optional fields)
	Summary *Summary `json:"summary,omitempty"` // Video summary data
}

// Summary represents video summary information
type Summary struct {
	Text               string    `json:"text"`               // The generated summary text
	SourceLanguage     string    `json:"sourceLanguage"`     // Language of subtitles used (e.g., "en", "es")
	SummaryGeneratedAt time.Time `json:"summaryGeneratedAt"` // When the summary was generated
}

// Author represents the author element in RSS feeds
type Author struct {
	Name string `xml:"name" json:"name"`
	URI  string `xml:"uri" json:"uri"`
}

// MediaGroup represents the media:group element in RSS feeds
type MediaGroup struct {
	MediaThumbnail   MediaThumbnail `xml:"http://search.yahoo.com/mrss/ thumbnail" json:"mediaThumbnail"`
	MediaTitle       string         `xml:"http://search.yahoo.com/mrss/ title" json:"mediaTitle"`
	MediaContent     MediaContent   `xml:"http://search.yahoo.com/mrss/ content" json:"mediaContent"`
	MediaDescription string         `xml:"http://search.yahoo.com/mrss/ description" json:"mediaDescription"`
}

// MediaContent represents the media:content element
type MediaContent struct {
	URL    string `xml:"url,attr"`
	Type   string `xml:"type,attr"`
	Width  string `xml:"width,attr"`
	Height string `xml:"height,attr"`
}

// MediaThumbnail represents the media:thumbnail element in RSS feeds
type MediaThumbnail struct {
	URL    string `xml:"url,attr"`
	Width  string `xml:"width,attr"`
	Height string `xml:"height,attr"`
}

type Link struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr"`
}

func (e *Entry) String() string {
	var s strings.Builder
	s.WriteString(fmt.Sprintf("Title: %s\nID: %s\nLink: %s\nPublished: %s\nThumbnail: %s\nContent: %s\n",
		strings.Trim(e.Title, " "),
		e.ID,
		e.Link.Href,
		e.Published.String(),
		e.MediaGroup.MediaThumbnail.URL,
		CleanContent(e.Content, 200, false), // Use exported CleanContent
	))
	return s.String()
}

// GetID returns the Entry's ID, implementing the ContentProvider interface
func (e Entry) GetID() string {
	return e.ID
}

// GetContent returns the Entry's Content, implementing the ContentProvider interface
func (e Entry) GetContent() string {
	return e.Content
}

// UnmarshalXML implements xml.Unmarshaler for custom time parsing
func (e *Entry) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type Alias Entry
	aux := &struct {
		Published string `xml:"published"`
		*Alias
	}{
		Alias: (*Alias)(e),
	}
	if err := d.DecodeElement(aux, &start); err != nil {
		return err
	}

	// Parse the time string
	if aux.Published != "" {
		t, err := time.Parse(time.RFC3339, aux.Published)
		if err != nil {
			// Try a different format if RFC3339 fails (sometimes YouTube uses this)
			t, err = time.Parse("2006-01-02T15:04:05-07:00", aux.Published)
			if err != nil {
				return fmt.Errorf("failed to parse published time '%s': %w", aux.Published, err)
			}
		}
		e.Published = t
	}
	return nil
}

// UnmarshalXML implements xml.Unmarshaler for custom link parsing
func (f *Feed) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type Alias Feed
	aux := &struct {
		Links []Link `xml:"link"`
		*Alias
	}{
		Alias: (*Alias)(f),
	}
	if err := d.DecodeElement(aux, &start); err != nil {
		return err
	}

	// Find the alternate link
	for _, link := range aux.Links {
		if link.Rel == "alternate" {
			f.URL = link
			break
		}
	}
	return nil
}
