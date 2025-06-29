// Package giftManager provides gift management functionality for the gift buying system.
// It handles communication with the Telegram API to retrieve available star gifts
// and manages the conversion of API responses to internal data structures.
package giftManager

import (
	"context"
	"gift-buyer/pkg/errors"

	"github.com/gotd/td/tg"
)

// giftManagerImpl implements the Giftmanager interface for managing gift operations.
// It provides methods to retrieve available gifts from the Telegram API and
// handles the parsing of API responses into usable data structures.
type giftManagerImpl struct {
	// api is the Telegram client used for API communication
	api *tg.Client
}

// NewGiftManager creates a new GiftManager instance with the specified Telegram API client.
// The manager will use this client to communicate with Telegram's gift API endpoints.
//
// Parameters:
//   - api: configured Telegram API client for making requests
//
// Returns:
//   - giftInterfaces.Giftmanager: configured gift manager instance
func NewGiftManager(api *tg.Client) *giftManagerImpl {
	return &giftManagerImpl{api: api}
}

// GetAvailableGifts retrieves all currently available star gifts from Telegram.
// It makes an API call to fetch the gift catalog and parses the response
// to extract individual StarGift objects.
//
// The method handles:
//   - API communication with Telegram's gift endpoints
//   - Response parsing and type validation
//   - Conversion of API response to internal gift structures
//
// Parameters:
//   - ctx: context for request cancellation and timeout control
//
// Returns:
//   - []*tg.StarGift: slice of available star gifts from Telegram
//   - error: API communication error, parsing error, or unexpected response type
//
// Possible errors:
//   - Network communication errors with Telegram API
//   - Unexpected response type from the API
//   - Context cancellation or timeout
func (gm *giftManagerImpl) GetAvailableGifts(ctx context.Context) ([]*tg.StarGift, error) {
	gifts, err := gm.api.PaymentsGetStarGifts(ctx, 0)
	if err != nil {
		return nil, err
	}

	starGifts, ok := gifts.(*tg.PaymentsStarGifts)
	if !ok {
		return nil, errors.Wrap(errors.New("unexpected response type"), "unexpected response type")
	}

	giftList := make([]*tg.StarGift, 0, len(starGifts.Gifts))
	for _, gift := range starGifts.Gifts {
		if starGift, ok := gift.(*tg.StarGift); ok {
			giftList = append(giftList, starGift)
		}
	}
	return giftList, nil
}
