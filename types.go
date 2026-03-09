package twitterinternalapi

import "time"

// Tweet represents a Twitter tweet.
type Tweet struct {
	ID          string      `json:"id"`
	Text        string      `json:"text"`
	CreatedAt   string      `json:"created_at"`
	Stats       *TweetStats `json:"stats"`
	ScheduledAt *string     `json:"scheduled_at,omitempty"`
}

// TweetStats contains tweet statistics.
type TweetStats struct {
	Retweets  int `json:"retweets"`
	Likes     int `json:"likes"`
	Replies   int `json:"replies"`
	Bookmarks int `json:"bookmarks"`
}

// CreateTweetOptions contains options for creating a tweet.
type CreateTweetOptions struct {
	MediaIDs  []string `json:"media_ids,omitempty"`
	ReplyTo   *string  `json:"reply_to,omitempty"`
	Sensitive bool     `json:"sensitive,omitempty"`
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
