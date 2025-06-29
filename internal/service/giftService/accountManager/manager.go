package accountManager

import (
	"context"
	"fmt"
	"gift-buyer/internal/service/giftService/giftInterfaces"
	"gift-buyer/pkg/errors"
	"gift-buyer/pkg/logger"

	"github.com/gotd/td/tg"
)

type accountManagerImpl struct {
	api        *tg.Client
	userIDs    []int
	channelIDs []int
	userCache  giftInterfaces.UserCache
}

func NewAccountManager(api *tg.Client, userIDs, channelIDs []int, userCache giftInterfaces.UserCache) *accountManagerImpl {
	return &accountManagerImpl{
		api:        api,
		userIDs:    userIDs,
		channelIDs: channelIDs,
		userCache:  userCache,
	}
}

func (am *accountManagerImpl) SetIds(ctx context.Context) error {
	if am.api == nil {
		return errors.New("API client is nil")
	}

	if len(am.userIDs) > 0 {
		if err := am.loadUsersToCache(ctx); err != nil {
			return errors.Wrap(err, "failed to load users to cache")
		}
	}

	if len(am.channelIDs) > 0 {
		if err := am.loadChannelsToCache(ctx); err != nil {
			return errors.Wrap(err, "failed to load channels to cache")
		}
	}

	return nil
}

func (am *accountManagerImpl) loadUsersToCache(ctx context.Context) error {
	if am.api == nil {
		return errors.New("API client is nil")
	}

	contacts, err := am.api.ContactsGetContacts(ctx, 0)
	if err != nil {
		return errors.Wrap(err, "failed to get contacts")
	}

	switch v := contacts.(type) {
	case *tg.ContactsContacts:

		cachedCount := 0
		targetUserMap := make(map[int]bool)
		for _, userID := range am.userIDs {
			targetUserMap[userID] = true
		}

		for _, user := range v.Users {
			u, ok := user.(*tg.User)
			if !ok {
				continue
			}

			am.userCache.SetUser(u)
			cachedCount++

		}

		notFoundUsers := []int{}
		for _, userID := range am.userIDs {
			if _, err := am.userCache.GetUser(int64(userID)); err != nil {
				notFoundUsers = append(notFoundUsers, userID)
			}
		}

		if len(notFoundUsers) > 0 {
			logger.GlobalLogger.Warnf("Target users not found in contacts: %v", notFoundUsers)
		}

	default:
		return errors.New(fmt.Sprintf("unexpected contacts response type: %T", contacts))
	}

	return nil
}

func (am *accountManagerImpl) loadChannelsToCache(ctx context.Context) error {

	cachedCount := 0
	notFoundChannels := []int{}

	for _, channelID := range am.channelIDs {
		channel, err := am.loadSingleChannel(ctx, channelID)
		if err != nil {
			notFoundChannels = append(notFoundChannels, channelID)
			continue
		}

		am.userCache.SetChannel(channel)
		cachedCount++
	}

	if len(notFoundChannels) > 0 {
		logger.GlobalLogger.Warnf("Channels not found or inaccessible: %v", notFoundChannels)
	}

	return nil
}

func (am *accountManagerImpl) loadSingleChannel(ctx context.Context, channelID int) (*tg.Channel, error) {
	if am.api == nil {
		return nil, errors.New("API client is nil")
	}

	var actualChannelID int64
	if channelID < -1000000000000 {
		actualChannelID = int64(-channelID - 1000000000000)
	} else {
		actualChannelID = int64(channelID)
	}

	channels, err := am.api.ChannelsGetChannels(ctx, []tg.InputChannelClass{
		&tg.InputChannel{
			ChannelID: actualChannelID,
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to get channel %d info", int(actualChannelID)))
	}

	for _, chat := range channels.GetChats() {
		if channel, ok := chat.(*tg.Channel); ok {
			channelCopy := *channel
			channelCopy.ID = int64(channelID)
			return &channelCopy, nil
		}
	}

	return nil, errors.New(fmt.Sprintf("channel %d not found in response", channelID))
}
