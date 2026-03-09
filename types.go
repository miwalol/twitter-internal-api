package twitterinternalapi

import "time"

// Tweet represents a Twitter tweet.
type Tweet struct {
	ID     string       `json:"rest_id"`
	Legacy *TweetLegacy `json:"legacy"`
}

// TweetLegacy contains the legacy tweet data
type TweetLegacy struct {
	FullText      string         `json:"full_text"`
	CreatedAt     string         `json:"created_at"`
	PublicMetrics *PublicMetrics `json:"public_metrics"`
}

// PublicMetrics contains tweet metrics
type PublicMetrics struct {
	RetweetCount  int `json:"retweet_count"`
	ReplyCount    int `json:"reply_count"`
	LikeCount     int `json:"like_count"`
	BookmarkCount int `json:"bookmark_count"`
}

// MediaEntity represents a media attachment in a tweet
type MediaEntity struct {
	MediaID     string        `json:"media_id"`
	TaggedUsers []interface{} `json:"tagged_users"`
}

// CreateTweetOptions contains options for creating a tweet.
type CreateTweetOptions struct {
	MediaIDs      []string      `json:"media_ids,omitempty"`
	MediaEntities []MediaEntity `json:"media_entities,omitempty"`
	ReplyTo       *string       `json:"reply_to,omitempty"`
	Sensitive     bool          `json:"sensitive,omitempty"`
}

// ScheduledTweet represents a scheduled tweet.
type ScheduledTweet struct {
	ID          string    `json:"id"`
	Text        string    `json:"text"`
	ScheduledAt time.Time `json:"scheduled_at"`
}

// StringPtr returns a pointer to a string.
func StringPtr(s string) *string {
	return &s
}

// IntPtr returns a pointer to an int.
func IntPtr(i int) *int {
	return &i
}
