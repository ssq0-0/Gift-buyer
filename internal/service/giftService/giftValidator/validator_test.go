package giftValidator

import (
	"gift-buyer/internal/config"
	"testing"

	"github.com/gotd/td/tg"
	"github.com/stretchr/testify/assert"
)

func TestNewGiftValidator(t *testing.T) {
	criterias := []config.Criterias{
		{
			MinPrice:    100,
			MaxPrice:    1000,
			TotalSupply: 50,
			Count:       10,
		},
	}

	validator := NewGiftValidator(criterias, 5000, false, false)
	assert.NotNil(t, validator)

	gv, ok := validator.(*GiftValidator)
	assert.True(t, ok)
	assert.Equal(t, criterias, gv.criteria)
	assert.Equal(t, int64(5000), gv.totalStarCap)
	assert.Equal(t, false, gv.testMode)
}

func TestGiftValidator_IsEligible_ValidGift(t *testing.T) {
	criterias := []config.Criterias{
		{
			MinPrice:    100,
			MaxPrice:    1000,
			TotalSupply: 50,
			Count:       10,
		},
	}
	validator := NewGiftValidator(criterias, 5000, false, false)

	gift := &tg.StarGift{
		ID:    1,
		Stars: 500, // Within price range
	}

	count, result := validator.IsEligible(gift)
	assert.True(t, result)
	assert.Equal(t, int64(10), count)
}

func TestGiftValidator_IsEligible_PriceTooLow(t *testing.T) {
	criterias := []config.Criterias{
		{
			MinPrice:    100,
			MaxPrice:    1000,
			TotalSupply: 50,
			Count:       10,
		},
	}
	validator := NewGiftValidator(criterias, 5000, false, false)

	gift := &tg.StarGift{
		ID:    1,
		Stars: 50, // Below minimum price
	}

	_, result := validator.IsEligible(gift)
	assert.False(t, result)
}

func TestGiftValidator_IsEligible_PriceTooHigh(t *testing.T) {
	criterias := []config.Criterias{
		{
			MinPrice:    100,
			MaxPrice:    1000,
			TotalSupply: 50,
			Count:       10,
		},
	}
	validator := NewGiftValidator(criterias, 5000, false, false)

	gift := &tg.StarGift{
		ID:    1,
		Stars: 1500, // Above maximum price
	}

	_, result := validator.IsEligible(gift)
	assert.False(t, result)
}

func TestGiftValidator_IsEligible_ExactMinPrice(t *testing.T) {
	criterias := []config.Criterias{
		{
			MinPrice:    100,
			MaxPrice:    1000,
			TotalSupply: 50,
			Count:       10,
		},
	}
	validator := NewGiftValidator(criterias, 5000, false, false)

	gift := &tg.StarGift{
		ID:    1,
		Stars: 100, // Exactly minimum price
	}

	_, result := validator.IsEligible(gift)
	assert.True(t, result)
}

func TestGiftValidator_IsEligible_ExactMaxPrice(t *testing.T) {
	criterias := []config.Criterias{
		{
			MinPrice:    100,
			MaxPrice:    1000,
			TotalSupply: 50,
			Count:       10,
		},
	}
	validator := NewGiftValidator(criterias, 5000, true, true)

	gift := &tg.StarGift{
		ID:      1,
		Stars:   1000, // Exactly maximum price
		Limited: true, // Must match LimitedStatus = true
	}

	_, result := validator.IsEligible(gift)
	assert.True(t, result)
}

func TestGiftValidator_IsEligible_ZeroCriteria(t *testing.T) {
	criterias := []config.Criterias{
		{
			MinPrice:    0,
			MaxPrice:    0,
			TotalSupply: 0,
			Count:       0,
		},
	}
	validator := NewGiftValidator(criterias, 0, true, true)

	gift := &tg.StarGift{
		ID:      1,
		Stars:   0,    // Zero price should be valid for zero criteria
		Limited: true, // Must match LimitedStatus = true
	}

	_, result := validator.IsEligible(gift)
	assert.True(t, result)
}

func TestGiftValidator_IsEligible_NegativePrice(t *testing.T) {
	criterias := []config.Criterias{
		{
			MinPrice:    100,
			MaxPrice:    1000,
			TotalSupply: 50,
			Count:       10,
		},
	}
	validator := NewGiftValidator(criterias, 5000, false, false)

	gift := &tg.StarGift{
		ID:    1,
		Stars: -100, // Negative price
	}

	_, result := validator.IsEligible(gift)
	assert.False(t, result)
}

func TestGiftValidator_MultipleCriterias(t *testing.T) {
	criterias := []config.Criterias{
		{
			MinPrice:    100,
			MaxPrice:    500,
			TotalSupply: 50,
			Count:       10,
		},
		{
			MinPrice:    600,
			MaxPrice:    1000,
			TotalSupply: 100,
			Count:       20,
		},
	}
	validator := NewGiftValidator(criterias, 5000, false, false)

	tests := []struct {
		name     string
		stars    int64
		expected bool
	}{
		{"price in first range", 300, true},
		{"price in second range", 800, true},
		{"price between ranges", 550, false},
		{"price below all ranges", 50, false},
		{"price above all ranges", 1500, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gift := &tg.StarGift{
				ID:    1,
				Stars: tt.stars,
			}

			_, result := validator.IsEligible(gift)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGiftValidator_PriceValid_EdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		criterias []config.Criterias
		stars     int64
		expected  bool
	}{
		{
			name: "price within range",
			criterias: []config.Criterias{
				{MinPrice: 100, MaxPrice: 1000, TotalSupply: 100},
			},
			stars:    500,
			expected: true,
		},
		{
			name: "price at minimum boundary",
			criterias: []config.Criterias{
				{MinPrice: 100, MaxPrice: 1000, TotalSupply: 100},
			},
			stars:    100,
			expected: true,
		},
		{
			name: "price at maximum boundary",
			criterias: []config.Criterias{
				{MinPrice: 100, MaxPrice: 1000, TotalSupply: 100},
			},
			stars:    1000,
			expected: true,
		},
		{
			name: "price below minimum",
			criterias: []config.Criterias{
				{MinPrice: 100, MaxPrice: 1000, TotalSupply: 100},
			},
			stars:    99,
			expected: false,
		},
		{
			name: "price above maximum",
			criterias: []config.Criterias{
				{MinPrice: 100, MaxPrice: 1000, TotalSupply: 100},
			},
			stars:    1001,
			expected: false,
		},
		{
			name: "zero range with zero price",
			criterias: []config.Criterias{
				{MinPrice: 0, MaxPrice: 0, TotalSupply: 0},
			},
			stars:    0,
			expected: true,
		},
		{
			name: "zero range with non-zero price",
			criterias: []config.Criterias{
				{MinPrice: 0, MaxPrice: 0, TotalSupply: 0},
			},
			stars:    1,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := &GiftValidator{
				criteria:     tt.criterias,
				totalStarCap: 5000,
			}

			gift := &tg.StarGift{
				Stars: tt.stars,
			}

			result := validator.priceValid(tt.criterias[0], gift)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGiftValidator_StarCapValidation(t *testing.T) {
	criterias := []config.Criterias{
		{
			MinPrice:    100,
			MaxPrice:    1000,
			TotalSupply: 50,
			Count:       10,
		},
	}
	validator := NewGiftValidator(criterias, 1000, false, false)

	tests := []struct {
		name     string
		gift     *tg.StarGift
		expected bool
	}{
		{
			name: "gift with zero availability total",
			gift: &tg.StarGift{
				ID:    1,
				Stars: 100,
				// AvailabilityTotal not set, GetAvailabilityTotal() will return 0
				// 100 * 0 = 0 <= 1000 = true
			},
			expected: true,
		},
		{
			name: "gift with zero stars",
			gift: &tg.StarGift{
				ID:    2,
				Stars: 0,
				// AvailabilityTotal not set, GetAvailabilityTotal() will return 0
				// 0 * 0 = 0 <= 1000 = true
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.(*GiftValidator).starCapValidation(tt.gift)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGiftValidator_TestMode(t *testing.T) {
	criterias := []config.Criterias{
		{
			MinPrice:    100,
			MaxPrice:    1000,
			TotalSupply: 50,
			Count:       10,
		},
	}

	tests := []struct {
		name          string
		testMode      bool
		limitedStatus bool
		gift          *tg.StarGift
		expected      bool
	}{
		{
			name:          "test mode - accepts unlimited gift",
			testMode:      true,
			limitedStatus: false,
			gift: &tg.StarGift{
				ID:      1,
				Stars:   500,
				Limited: false,
				SoldOut: false,
			},
			expected: true,
		},
		{
			name:          "test mode - accepts limited gift",
			testMode:      true,
			limitedStatus: true,
			gift: &tg.StarGift{
				ID:      2,
				Stars:   500,
				Limited: true,
				SoldOut: false,
			},
			expected: true,
		},
		{
			name:          "test mode - rejects sold out gift",
			testMode:      true,
			limitedStatus: false,
			gift: &tg.StarGift{
				ID:      3,
				Stars:   500,
				Limited: false,
				SoldOut: true,
			},
			expected: false,
		},
		{
			name:          "production mode - rejects limited gift with low supply",
			testMode:      false,
			limitedStatus: true,
			gift: &tg.StarGift{
				ID:                  4,
				Stars:               500,
				Limited:             true,
				SoldOut:             false,
				AvailabilityRemains: 10, // Less than TotalSupply (50)
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewGiftValidator(criterias, 5000, tt.testMode, tt.limitedStatus)
			_, result := validator.IsEligible(tt.gift)
			assert.Equal(t, tt.expected, result)
		})
	}
}
