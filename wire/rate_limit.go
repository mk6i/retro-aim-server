package wire

import "time"

type RateClass struct {
	ID              uint16
	WindowSize      int64
	ClearLevel      int64
	AlertLevel      int64
	LimitLevel      int64
	DisconnectLevel int64
	MaxLevel        int64
}

type RateLimitStatus int

const (
	RateLimitStatusLimited    RateLimitStatus = 1 // You're currently rate-limited
	RateLimitStatusAlert      RateLimitStatus = 2 // You're close to being rate-limited
	RateLimitStatusClear      RateLimitStatus = 3 // You're under the rate limit; all good
	RateLimitStatusDisconnect RateLimitStatus = 4 // You're under the rate limit; all good
)

var class1 = RateClass{
	ID:              1,
	WindowSize:      80,
	ClearLevel:      2500,
	AlertLevel:      2000,
	LimitLevel:      1500,
	DisconnectLevel: 800,
	MaxLevel:        6000,
}

var class2 = RateClass{
	ID:              2,
	WindowSize:      80,
	ClearLevel:      3000,
	AlertLevel:      2000,
	LimitLevel:      1500,
	DisconnectLevel: 3000,
	MaxLevel:        6000,
}

var class3 = RateClass{
	ID:              3,
	WindowSize:      20,
	ClearLevel:      5100,
	AlertLevel:      5000,
	LimitLevel:      4000,
	DisconnectLevel: 3000,
	MaxLevel:        6000,
}

var class4 = RateClass{
	ID:              4,
	WindowSize:      20,
	ClearLevel:      5500,
	AlertLevel:      5300,
	LimitLevel:      4200,
	DisconnectLevel: 3000,
	MaxLevel:        8000,
}

var class5 = RateClass{
	ID:              5,
	WindowSize:      10,
	ClearLevel:      5500,
	AlertLevel:      5300,
	LimitLevel:      4200,
	DisconnectLevel: 3000,
	MaxLevel:        8000,
}

// CheckRateLimit calculates moving average
func CheckRateLimit(last time.Time, now time.Time, class RateClass, curAvg int64) (status RateLimitStatus, newAvg int64) {
	delta := now.Sub(last).Milliseconds()

	curAvg = (curAvg * (class.WindowSize - 1) / class.WindowSize) + (delta / class.WindowSize)

	if curAvg > class.MaxLevel {
		curAvg = class.MaxLevel
	}

	switch {
	case curAvg < class.DisconnectLevel:
		return RateLimitStatusDisconnect, curAvg
	case curAvg < class.LimitLevel:
		return RateLimitStatusLimited, curAvg
	case curAvg < class.AlertLevel:
		return RateLimitStatusAlert, curAvg
	}

	return RateLimitStatusClear, curAvg
}
