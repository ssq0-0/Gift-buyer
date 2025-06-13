package giftService

import (
	"gift-buyer/internal/config"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFactory(t *testing.T) {
	cfg := &config.SoftConfig{
		TgSettings: config.TgSettings{
			AppId:    123456,
			ApiHash:  "test_hash",
			Phone:    "+1234567890",
			Password: "test_password",
		},
		Criterias: []config.Criterias{
			{
				MinPrice:    100,
				MaxPrice:    1000,
				TotalSupply: 50,
			},
		},
		Receiver: config.ReceiverParams{
			Type:           []int{1},
			UserReceiverID: []int{987654321},
		},
		Ticker: 30.0,
	}

	factory := NewFactory(cfg)
	assert.NotNil(t, factory)
	assert.Equal(t, cfg, factory.cfg)
}

func TestNewFactory_NilConfig(t *testing.T) {
	factory := NewFactory(nil)
	assert.NotNil(t, factory)
	assert.Nil(t, factory.cfg)
}

func TestFactory_Structure(t *testing.T) {
	cfg := &config.SoftConfig{
		TgSettings: config.TgSettings{
			AppId:    123456,
			ApiHash:  "test_hash",
			Phone:    "+1234567890",
			Password: "test_password",
		},
		Criterias: []config.Criterias{
			{
				MinPrice:    100,
				MaxPrice:    1000,
				TotalSupply: 50,
			},
		},
		Receiver: config.ReceiverParams{
			Type:           []int{1},
			UserReceiverID: []int{987654321},
		},
		Ticker: 30.0,
	}

	factory := NewFactory(cfg)
	assert.NotNil(t, factory)
	assert.Equal(t, cfg, factory.cfg)

	// Verify factory has CreateSystem method
	assert.NotNil(t, factory.CreateSystem)
}

func TestFactory_CreateSystemMethod(t *testing.T) {
	cfg := &config.SoftConfig{
		TgSettings: config.TgSettings{
			AppId:    123456,
			ApiHash:  "test_hash",
			Phone:    "+1234567890",
			Password: "test_password",
		},
		Criterias: []config.Criterias{
			{
				MinPrice:    100,
				MaxPrice:    1000,
				TotalSupply: 50,
			},
		},
		Receiver: config.ReceiverParams{
			Type:           []int{1},
			UserReceiverID: []int{987654321},
		},
		Ticker: 30.0,
	}

	factory := NewFactory(cfg)

	// Test that CreateSystem method exists and can be called
	assert.NotPanics(t, func() {
		// Don't actually call CreateSystem as it will try to connect to Telegram
		// Just verify the method signature exists
		_ = factory.CreateSystem
	})
}

func TestFactory_ConfigMutability(t *testing.T) {
	cfg := &config.SoftConfig{
		TgSettings: config.TgSettings{
			AppId:    123456,
			ApiHash:  "original_hash",
			Phone:    "+1234567890",
			Password: "original_password",
		},
		Criterias: []config.Criterias{
			{
				MinPrice:    100,
				MaxPrice:    1000,
				TotalSupply: 50,
			},
		},
		Receiver: config.ReceiverParams{
			Type:           []int{1},
			UserReceiverID: []int{987654321},
		},
		Ticker: 30.0,
	}

	factory := NewFactory(cfg)

	// Modify original config
	cfg.TgSettings.ApiHash = "modified_hash"
	cfg.TgSettings.AppId = 999999
	cfg.Ticker = 60.0

	// Factory should still reference the same config object
	assert.Equal(t, "modified_hash", factory.cfg.TgSettings.ApiHash)
	assert.Equal(t, 999999, factory.cfg.TgSettings.AppId)
	assert.Equal(t, 60.0, factory.cfg.Ticker)
}

func TestFactory_MinimalConfig(t *testing.T) {
	cfg := &config.SoftConfig{
		TgSettings: config.TgSettings{
			AppId:   1,
			ApiHash: "h",
			Phone:   "1",
		},
		Criterias: []config.Criterias{
			{
				MinPrice:    1,
				MaxPrice:    2,
				TotalSupply: 1,
			},
		},
		Receiver: config.ReceiverParams{
			Type:           []int{0},
			UserReceiverID: []int{0},
		},
		Ticker: 1.0,
	}

	factory := NewFactory(cfg)
	assert.NotNil(t, factory)
	assert.Equal(t, cfg, factory.cfg)
	assert.Equal(t, 1, factory.cfg.TgSettings.AppId)
	assert.Equal(t, "h", factory.cfg.TgSettings.ApiHash)
	assert.Equal(t, "1", factory.cfg.TgSettings.Phone)
	assert.Equal(t, 1, len(factory.cfg.Criterias))
	assert.Equal(t, 1.0, factory.cfg.Ticker)
}

func TestFactory_MultipleCriterias(t *testing.T) {
	cfg := &config.SoftConfig{
		TgSettings: config.TgSettings{
			AppId:    123456,
			ApiHash:  "test_hash",
			Phone:    "+1234567890",
			Password: "test_password",
		},
		Criterias: []config.Criterias{
			{
				MinPrice:    100,
				MaxPrice:    500,
				TotalSupply: 25,
			},
			{
				MinPrice:    1000,
				MaxPrice:    5000,
				TotalSupply: 100,
			},
			{
				MinPrice:    10000,
				MaxPrice:    50000,
				TotalSupply: 10,
			},
		},
		Receiver: config.ReceiverParams{
			Type:           []int{1},
			UserReceiverID: []int{987654321},
		},
		Ticker: 15.0,
	}

	factory := NewFactory(cfg)
	assert.NotNil(t, factory)
	assert.Equal(t, cfg, factory.cfg)
	assert.Equal(t, 3, len(factory.cfg.Criterias))

	// Verify first criteria
	assert.Equal(t, int64(100), factory.cfg.Criterias[0].MinPrice)
	assert.Equal(t, int64(500), factory.cfg.Criterias[0].MaxPrice)
	assert.Equal(t, int64(25), factory.cfg.Criterias[0].TotalSupply)

	// Verify second criteria
	assert.Equal(t, int64(1000), factory.cfg.Criterias[1].MinPrice)
	assert.Equal(t, int64(5000), factory.cfg.Criterias[1].MaxPrice)
	assert.Equal(t, int64(100), factory.cfg.Criterias[1].TotalSupply)

	// Verify third criteria
	assert.Equal(t, int64(10000), factory.cfg.Criterias[2].MinPrice)
	assert.Equal(t, int64(50000), factory.cfg.Criterias[2].MaxPrice)
	assert.Equal(t, int64(10), factory.cfg.Criterias[2].TotalSupply)
}

func TestFactory_EdgeCaseValues(t *testing.T) {
	cfg := &config.SoftConfig{
		TgSettings: config.TgSettings{
			AppId:    -1, // Negative app ID
			ApiHash:  "test_hash",
			Phone:    "+1234567890",
			Password: "",
		},
		Criterias: []config.Criterias{
			{
				MinPrice:    -100, // Negative prices
				MaxPrice:    -50,
				TotalSupply: -10, // Negative supply
			},
		},
		Receiver: config.ReceiverParams{
			Type:           []int{-1},         // Negative type
			UserReceiverID: []int{-987654321}, // Negative receiver
		},
		Ticker: -30.0, // Negative ticker
	}

	factory := NewFactory(cfg)
	assert.NotNil(t, factory)
	assert.Equal(t, cfg, factory.cfg)

	// Verify edge case values are preserved
	assert.Equal(t, -1, factory.cfg.TgSettings.AppId)
	assert.Equal(t, "", factory.cfg.TgSettings.Password)
	assert.Equal(t, int64(-100), factory.cfg.Criterias[0].MinPrice)
	assert.Equal(t, int64(-50), factory.cfg.Criterias[0].MaxPrice)
	assert.Equal(t, int64(-10), factory.cfg.Criterias[0].TotalSupply)
	assert.Equal(t, []int{-1}, factory.cfg.Receiver.Type)
	assert.Equal(t, []int{-987654321}, factory.cfg.Receiver.UserReceiverID)
	assert.Equal(t, -30.0, factory.cfg.Ticker)
}

func TestFactory_ZeroValues(t *testing.T) {
	cfg := &config.SoftConfig{
		TgSettings: config.TgSettings{
			AppId:    0,
			ApiHash:  "",
			Phone:    "",
			Password: "",
		},
		Criterias: []config.Criterias{
			{
				MinPrice:    0,
				MaxPrice:    0,
				TotalSupply: 0,
			},
		},
		Receiver: config.ReceiverParams{
			Type:           []int{0},
			UserReceiverID: []int{0},
		},
		Ticker: 0.0,
	}

	factory := NewFactory(cfg)
	assert.NotNil(t, factory)
	assert.Equal(t, cfg, factory.cfg)

	// Verify zero values are preserved
	assert.Equal(t, 0, factory.cfg.TgSettings.AppId)
	assert.Equal(t, "", factory.cfg.TgSettings.ApiHash)
	assert.Equal(t, "", factory.cfg.TgSettings.Phone)
	assert.Equal(t, "", factory.cfg.TgSettings.Password)
	assert.Equal(t, int64(0), factory.cfg.Criterias[0].MinPrice)
	assert.Equal(t, int64(0), factory.cfg.Criterias[0].MaxPrice)
	assert.Equal(t, int64(0), factory.cfg.Criterias[0].TotalSupply)
	assert.Equal(t, []int{0}, factory.cfg.Receiver.Type)
	assert.Equal(t, []int{0}, factory.cfg.Receiver.UserReceiverID)
	assert.Equal(t, 0.0, factory.cfg.Ticker)
}
