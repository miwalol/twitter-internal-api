package twitterinternalapi

import (
	"encoding/json"
	"fmt"
)

// TweetsService handles tweet-related API operations.
type TweetsService struct {
	client *Client
}

// Create creates a new tweet.
func (s *TweetsService) Create(text string, opts *CreateTweetOptions) (*Tweet, error) {
	if text == "" {
		return nil, fmt.Errorf("tweet text cannot be empty")
	}

	if opts == nil {
		opts = &CreateTweetOptions{}
	}

	variables := map[string]interface{}{
		"tweet_text": text,
	}

	if len(opts.MediaIDs) > 0 {
		variables["media_ids"] = opts.MediaIDs
	}
	if opts.ReplyTo != nil {
		variables["reply_to_id"] = *opts.ReplyTo
	}
	if opts.Sensitive {
		variables["possibly_sensitive"] = true
	}

	query := `mutation CreateTweet($tweet_text: String!, $media_ids: [String!], $reply_to_id: String, $possibly_sensitive: Boolean) {
		create_tweet(input: {
			tweet_text: $tweet_text
			media_ids: $media_ids
			reply_to_id: $reply_to_id
			possibly_sensitive: $possibly_sensitive
		}) {
			data {
				create_tweet_response {
					tweet_results {
						result {
							__typename
							rest_id
							core {
								user_results {
									result {
										id
										legacy {
											name
											screen_name
										}
									}
								}
							}
							legacy {
								full_text
								created_at
								public_metrics {
									retweet_count
									reply_count
									like_count
									bookmark_count
								}
							}
						}
					}
				}
			}
		}
	}`

	result, err := s.client.executeGraphQL(query, variables)
	if err != nil {
		return nil, err
	}

	// Extract tweet from response
	data, ok := result["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response structure")
	}

	respBytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	var tweet Tweet
	if err := json.Unmarshal(respBytes, &tweet); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tweet: %w", err)
	}

	return &tweet, nil
}

// Schedule schedules a tweet for future posting.
// scheduledAt should be an ISO 8601 timestamp (e.g., "2025-12-25T10:30:00Z")
func (s *TweetsService) Schedule(text string, scheduledAt string) (*ScheduledTweet, error) {
	if text == "" {
		return nil, fmt.Errorf("tweet text cannot be empty")
	}
	if scheduledAt == "" {
		return nil, fmt.Errorf("scheduledAt cannot be empty")
	}

	variables := map[string]interface{}{
		"tweet_text":   text,
		"scheduled_at": scheduledAt,
	}

	query := `mutation CreateScheduledTweet($tweet_text: String!, $scheduled_at: String!) {
		create_scheduled_tweet(input: {
			tweet_text: $tweet_text
			scheduled_at: $scheduled_at
		}) {
			data {
				scheduled_tweet_response {
					scheduled_tweet {
						id
						text
						scheduled_at
					}
				}
			}
		}
	}`

	result, err := s.client.executeGraphQL(query, variables)
	if err != nil {
		return nil, err
	}

	data, ok := result["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response structure")
	}

	respBytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	var scheduled ScheduledTweet
	if err := json.Unmarshal(respBytes, &scheduled); err != nil {
		return nil, fmt.Errorf("failed to unmarshal scheduled tweet: %w", err)
	}

	return &scheduled, nil
}

// GetScheduled fetches all scheduled tweets.
func (s *TweetsService) GetScheduled() ([]*ScheduledTweet, error) {
	query := `query FetchScheduledTweets {
		scheduled_tweets {
			scheduled_tweets {
				id
				text
				scheduled_at
			}
		}
	}`

	result, err := s.client.executeGraphQL(query, make(map[string]interface{}))
	if err != nil {
		return nil, err
	}

	data, ok := result["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response structure")
	}

	respBytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	var tweets []*ScheduledTweet
	if err := json.Unmarshal(respBytes, &tweets); err != nil {
		return nil, fmt.Errorf("failed to unmarshal scheduled tweets: %w", err)
	}

	return tweets, nil
}

// DeleteScheduled deletes a scheduled tweet.
func (s *TweetsService) DeleteScheduled(scheduledTweetID string) error {
	variables := map[string]interface{}{
		"id": scheduledTweetID,
	}

	query := `mutation DeleteScheduledTweet($id: String!) {
		delete_scheduled_tweet(input: {
			id: $id
		}) {
			data {
				delete_scheduled_tweet_response {
					success
				}
			}
		}
	}`

	_, err := s.client.executeGraphQL(query, variables)
	return err
}
