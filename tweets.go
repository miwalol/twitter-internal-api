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

	// Build media entities
	mediaEntities := []interface{}{}
	if len(opts.MediaEntities) > 0 {
		for _, entity := range opts.MediaEntities {
			mediaEntities = append(mediaEntities, map[string]interface{}{
				"media_id":     entity.MediaID,
				"tagged_users": entity.TaggedUsers,
			})
		}
	}

	possiblySensitive := false
	if opts.Sensitive {
		possiblySensitive = true
	}

	variables := map[string]interface{}{
		"tweet_text": text,
		"media": map[string]interface{}{
			"media_entities":     mediaEntities,
			"possibly_sensitive": possiblySensitive,
		},
		"semantic_annotation_ids":  []interface{}{},
		"disallowed_reply_options": nil,
	}

	// Features from Twitter's actual API
	features := map[string]bool{
		"premium_content_api_read_enabled":                                        false,
		"communities_web_enable_tweet_community_results_fetch":                    true,
		"c9s_tweet_anatomy_moderator_badge_enabled":                               true,
		"responsive_web_grok_analyze_button_fetch_trends_enabled":                 false,
		"responsive_web_grok_analyze_post_followups_enabled":                      true,
		"responsive_web_jetfuel_frame":                                            true,
		"responsive_web_grok_share_attachment_enabled":                            true,
		"responsive_web_grok_annotations_enabled":                                 true,
		"responsive_web_edit_tweet_api_enabled":                                   true,
		"graphql_is_translatable_rweb_tweet_is_translatable_enabled":              true,
		"view_counts_everywhere_api_enabled":                                      true,
		"longform_notetweets_consumption_enabled":                                 true,
		"responsive_web_twitter_article_tweet_consumption_enabled":                true,
		"tweet_awards_web_tipping_enabled":                                        false,
		"content_disclosure_indicator_enabled":                                    true,
		"content_disclosure_ai_generated_indicator_enabled":                       true,
		"responsive_web_grok_show_grok_translated_post":                           false,
		"responsive_web_grok_analysis_button_from_backend":                        true,
		"post_ctas_fetch_enabled":                                                 true,
		"longform_notetweets_rich_text_read_enabled":                              true,
		"longform_notetweets_inline_media_enabled":                                false,
		"profile_label_improvements_pcf_label_in_post_enabled":                    true,
		"responsive_web_profile_redirect_enabled":                                 false,
		"rweb_tipjar_consumption_enabled":                                         false,
		"verified_phone_label_enabled":                                            false,
		"articles_preview_enabled":                                                true,
		"responsive_web_grok_community_note_auto_translation_is_enabled":          false,
		"responsive_web_graphql_skip_user_profile_image_extensions_enabled":       false,
		"freedom_of_speech_not_reach_fetch_enabled":                               true,
		"standardized_nudges_misinfo":                                             true,
		"tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled": true,
		"responsive_web_grok_image_annotation_enabled":                            true,
		"responsive_web_grok_imagine_annotation_enabled":                          true,
		"responsive_web_graphql_timeline_navigation_enabled":                      true,
		"responsive_web_enhance_cards_enabled":                                    false,
	}

	result, err := s.client.ExecuteGraphQL(variables, "sb6vH7FMb090KdK6IZaakw", "CreateTweet", features)
	if err != nil {
		return nil, err
	}

	// Extract tweet from response
	data, ok := result["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response structure: missing data field")
	}

	createTweet, ok := data["create_tweet"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response structure: missing create_tweet")
	}

	tweetResults, ok := createTweet["tweet_results"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response structure: missing tweet_results")
	}

	resultData, ok := tweetResults["result"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response structure: missing result")
	}

	respBytes, err := json.Marshal(resultData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response: %w", err)
	}

	var tweet Tweet
	if err := json.Unmarshal(respBytes, &tweet); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tweet: %w", err)
	}

	return &tweet, nil
}

// Delete deletes a tweet by ID.
func (s *TweetsService) Delete(tweetID string) error {
	if tweetID == "" {
		return fmt.Errorf("tweet ID cannot be empty")
	}

	variables := map[string]interface{}{
		"tweet_id": tweetID,
	}

	_, err := s.client.ExecuteGraphQL(variables, "nxpZCY2K-I6QoFHAHeojFQ", "DeleteTweet", nil)
	return err
}
