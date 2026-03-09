package twitterinternalapi

import (
	"log"
	"testing"
)

// Example: Post a simple tweet
func TestPostTweet(t *testing.T) {
	client := NewClient("auth_token cookie", "ct0 cookie")

	// Post a simple tweet
	tweet, err := client.Tweets.Create("test", nil)
	if err != nil {
		log.Fatal(err)
	}
	if tweet == nil {
		log.Fatal("tweet is nil")
	}
	log.Println("Posted tweet with ID:", tweet.ID)
	if tweet.Legacy != nil {
		log.Println("Tweet text:", tweet.Legacy.FullText)
		log.Println("Created at:", tweet.Legacy.CreatedAt)
	}
}

// TODO: Update these tests to use new ExecuteGraphQL API
// func TestScheduleTweet(t *testing.T) { ... }
// func TestGetScheduledTweets(t *testing.T) { ... }
// func TestPostTweetWithOptions(t *testing.T) { ... }
