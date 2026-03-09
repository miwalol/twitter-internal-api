package twitterinternalapi

import (
	"log"
	"time"
)

func ExampleClient_PostTweet() {
	client := NewClient("your-auth-token-here")

	// Post a simple tweet
	tweet, err := client.Tweets.Create("Hello from Go!", nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Posted tweet with ID:", tweet.ID)
}

func ExampleClient_ScheduleTweet() {
	client := NewClient("your-auth-token-here")

	// Schedule a tweet for tomorrow at 10 AM
	tomorrow := time.Now().AddDate(0, 0, 1)
	scheduledTime := tomorrow.Format("2006-01-02T15:04:05Z")

	scheduled, err := client.Tweets.Schedule("Good morning everyone!", scheduledTime)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Tweet scheduled for:", scheduled.ScheduledAt)
}

func ExampleClient_GetScheduledTweets() {
	client := NewClient("your-auth-token-here")

	// Get all scheduled tweets
	tweets, err := client.Tweets.GetScheduled()
	if err != nil {
		log.Fatal(err)
	}

	for _, tweet := range tweets {
		log.Printf("Tweet: %s\nScheduled at: %v\n", tweet.Text, tweet.ScheduledAt)
	}
}

func ExampleClient_PostTweetWithOptions() {
	client := NewClient("your-auth-token-here")

	// Post a tweet with options
	opts := &CreateTweetOptions{
		MediaIDs:  []string{"media-id-1"},
		Sensitive: true,
	}

	tweet, err := client.Tweets.Create("Check out this image!", opts)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Posted tweet with ID:", tweet.ID)
}
