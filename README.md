# Twitter Internal API

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/miwalol/twitter-internal-api)
[![Go Reference](https://pkg.go.dev/badge/github.com/miwalol/twitter-internal-api.svg)](https://pkg.go.dev/github.com/miwalol/twitter-internal-api)
[![Go Report Card](https://goreportcard.com/badge/github.com/miwalol/twitter-internal-api)](https://goreportcard.com/report/github.com/miwalol/twitter-internal-api)

A lightweight Go client to interact with the internal Twitter API used on the web version, with just your `auth_token` and `ct0` cookies.

Originally built for [Miwa.lol](https://miwa.lol/), this package is open for anyone to use and contribute to.

## Installation

```bash
go get github.com/miwalol/twitter-internal-api
```

## Quick Start

```go
package main

import (
	"log"
	"github.com/miwalol/twitter-internal-api"
)

func main() {
	client := twitterinternalapi.NewClient("your-auth-token", "your-csrf-token")
	
	tweet, err := client.Tweets.Create("Hello from Go!", nil)
	if err != nil {
		log.Fatal(err)
	}
	
	log.Println("Tweet ID:", tweet.ID)
	log.Println("Tweet Text:", tweet.Legacy.FullText)
}
```

## Usage Examples

### Authentication

```go
import "github.com/miwalol/twitter-internal-api"

// Create a client with auth token and CSRF token
client := twitterinternalapi.NewClient(
	"your-bearer-token-here",
	"your-csrf-token-here",
)

// Set cookies if needed
client.SetCookies("auth_token=xxx; ct0=yyy")

// Set additional CSRF token
client.SetCSRFToken("your-csrf-token")

// Register a callback for when ct0 is refreshed (e.g. to persist it)
client.OnCSRFRefreshed(func(newToken string) {
	// save newToken to cache, database, etc.
})
```

### Creating Tweets

```go
// Simple tweet
tweet, err := client.Tweets.Create("Hello Twitter!", nil)
if err != nil {
	log.Fatal(err)
}
log.Println("Posted:", tweet.ID, tweet.Legacy.FullText)

// Tweet with sensitivity flag
opts := &twitterinternalapi.CreateTweetOptions{
	Sensitive: true,
}
tweet, err := client.Tweets.Create("Sensitive content", opts)
```

### Uploading Media

```go
// Upload from a file path
mediaID, err := client.UploadMedia("path/to/image.png", "image/png")
if err != nil {
	log.Fatal(err)
}

// Upload from a byte slice
data, _ := os.ReadFile("path/to/image.png")
mediaID, err := client.UploadMediaBytes(data, "image/png")
if err != nil {
	log.Fatal(err)
}

// Create tweet with uploaded media
opts := &twitterinternalapi.CreateTweetOptions{
	MediaEntities: []twitterinternalapi.MediaEntity{
		{MediaID: mediaID, TaggedUsers: []interface{}{}},
	},
}
tweet, err := client.Tweets.Create("Posted with image!", opts)
```

### Deleting Tweets

```go
// Delete a tweet by ID
err := client.Tweets.Delete("2031054061076685198")
if err != nil {
	log.Fatal(err)
}
log.Println("Tweet deleted")
```

### GraphQL Requests

For advanced use cases, you can make custom GraphQL queries:

```go
variables := map[string]interface{}{
	"tweet_text": "Custom query",
}

features := map[string]bool{
	"responsive_web_edit_tweet_api_enabled": true,
}

result, err := client.ExecuteGraphQL(
	variables,
	"sb6vH7FMb090KdK6IZaakw", // queryId
	"CreateTweet",              // operationName
	features,
)
if err != nil {
	log.Fatal(err)
}
log.Println(result)
```

## License

MIT License - see [LICENSE](LICENSE)