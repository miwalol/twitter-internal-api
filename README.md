# Twitter Internal API Go Package

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/miwalol/twitter-internal-api)
[![Go Reference](https://pkg.go.dev/badge/github.com/miwalol/twitter-internal-api.svg)](https://pkg.go.dev/github.com/miwalol/twitter-internal-api)

A minimal Go package for posting and scheduling tweets using the internal Twitter API with an authentication token.

This package is public so that anyone can use it and contribute, but keep in mind that we made it for our usage.

## Installation

```bash
go get github.com/miwalol/twitter-internal-api
```

## Usage

```go
package main

import (
	"log"
	"github.com/miwalol/twitter-internal-api"
)

func main() {
	// Create a new client with your auth token
	client := twitterinternalapi.NewClient("your-auth-token-here")
	
	// Post a tweet
	tweet, err := client.Tweets.Create("Hello from Go!", nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Tweet posted with ID:", tweet.ID)
}
```
