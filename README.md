# Twitter Internal API

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/miwalol/twitter-internal-api)
[![Go Reference](https://pkg.go.dev/badge/github.com/miwalol/twitter-internal-api.svg)](https://pkg.go.dev/github.com/miwalol/twitter-internal-api)

A lightweight Go client for the internal Twitter API. Post and schedule tweets programmatically with just an auth token.

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
	client := twitterinternalapi.NewClient("your-auth-token")
	
	tweet, err := client.Tweets.Create("Hello from Go!", nil)
	if err != nil {
		log.Fatal(err)
	}
	
	log.Println("Tweet ID:", tweet.ID)
}
```

## License

MIT License - see [LICENSE](LICENSE)