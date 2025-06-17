package accountManager

import (
	"context"
	"fmt"
	"gift-buyer/internal/service/giftService/giftInterfaces"
	"gift-buyer/pkg/errors"
	"gift-buyer/pkg/logger"

	"github.com/gotd/td/tg"
)

type AccountManager struct {
	api        *tg.Client
	userIDs    []int
	channelIDs []int
	userCache  giftInterfaces.UserCache
}

func NewAccountManager(api *tg.Client, userIDs, channelIDs []int, userCache giftInterfaces.UserCache) *AccountManager {
	return &AccountManager{
		api:        api,
		userIDs:    userIDs,
		channelIDs: channelIDs,
		userCache:  userCache,
	}
}

func (am *AccountManager) SetIds(ctx context.Context) error {
	logger.GlobalLogger.Info("Starting to load users and channels into cache")

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

	logger.GlobalLogger.Info("Successfully loaded all users and channels into cache")
	return nil
}

func (am *AccountManager) loadUsersToCache(ctx context.Context) error {
	logger.GlobalLogger.Debugf("Loading users to cache, target user IDs: %v", am.userIDs)

	contacts, err := am.api.ContactsGetContacts(ctx, 0)
	if err != nil {
		return errors.Wrap(err, "failed to get contacts")
	}

	switch v := contacts.(type) {
	case *tg.ContactsContacts:
		logger.GlobalLogger.Debugf("Found %d contacts, caching all users", len(v.Users))

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

			if targetUserMap[int(u.ID)] {
				logger.GlobalLogger.Debugf("Target user %d found and cached", u.ID)
			}
		}

		logger.GlobalLogger.Infof("Cached %d users from contacts", cachedCount)

		notFoundUsers := []int{}
		for _, userID := range am.userIDs {
			if _, err := am.userCache.GetUser(int64(userID)); err != nil {
				notFoundUsers = append(notFoundUsers, userID)
			}
		}

		if len(notFoundUsers) > 0 {
			logger.GlobalLogger.Warnf("Target users not found in contacts: %v", notFoundUsers)
			logger.GlobalLogger.Warn("These users need to be added to contacts or started a conversation with")
		}

	default:
		return errors.New(fmt.Sprintf("unexpected contacts response type: %T", contacts))
	}

	return nil
}

func (am *AccountManager) loadChannelsToCache(ctx context.Context) error {
	logger.GlobalLogger.Debugf("Loading channels to cache, target channel IDs: %v", am.channelIDs)

	cachedCount := 0
	notFoundChannels := []int{}

	for _, channelID := range am.channelIDs {
		channel, err := am.loadSingleChannel(ctx, channelID)
		if err != nil {
			logger.GlobalLogger.Errorf("Failed to load channel %d: %v", channelID, err)
			notFoundChannels = append(notFoundChannels, channelID)
			continue
		}

		am.userCache.SetChannel(channel)
		cachedCount++
		logger.GlobalLogger.Debugf("Channel %d cached successfully", channelID)
	}

	logger.GlobalLogger.Infof("Cached %d channels", cachedCount)

	if len(notFoundChannels) > 0 {
		logger.GlobalLogger.Warnf("Channels not found or inaccessible: %v", notFoundChannels)
		logger.GlobalLogger.Warn("These channels need to be joined or made accessible")
	}

	return nil
}

func (am *AccountManager) loadSingleChannel(ctx context.Context, channelID int) (*tg.Channel, error) {
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
			logger.GlobalLogger.Debugf("Channel found: real ID=%d, config ID=%d, Title=%s, AccessHash=%d", channel.ID, channelID, channel.Title, channel.AccessHash)
			// Создаем копию канала с исходным ID, но сохраняем оригинальный AccessHash
			channelCopy := *channel
			channelCopy.ID = int64(channelID) // Используем исходный ID из конфигурации
			// AccessHash остается оригинальным - это критически важно для API
			logger.GlobalLogger.Debugf("Channel copy: ID=%d, AccessHash=%d", channelCopy.ID, channelCopy.AccessHash)
			return &channelCopy, nil
		}
	}

	return nil, errors.New(fmt.Sprintf("channel %d not found in response", channelID))
}
