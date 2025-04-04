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
	AlertLevel:      1000,
	LimitLevel:      500,
	DisconnectLevel: 300,
	MaxLevel:        8000,
}

var class2 = RateClass{
	ID:              2,
	WindowSize:      80,
	ClearLevel:      3000,
	AlertLevel:      1000,
	LimitLevel:      500,
	DisconnectLevel: 300,
	MaxLevel:        7000,
}

var class3 = RateClass{
	ID:              3,
	WindowSize:      20,
	ClearLevel:      4100,
	AlertLevel:      4000,
	LimitLevel:      3000,
	DisconnectLevel: 2000,
	MaxLevel:        7000,
}

var class4 = RateClass{
	ID:              4,
	WindowSize:      20,
	ClearLevel:      4500,
	AlertLevel:      4300,
	LimitLevel:      3200,
	DisconnectLevel: 2000,
	MaxLevel:        8000,
}

var class5 = RateClass{
	ID:              5,
	WindowSize:      10,
	ClearLevel:      4500,
	AlertLevel:      4300,
	LimitLevel:      3200,
	DisconnectLevel: 2000,
	MaxLevel:        9000,
}

// CheckRateLimit calculates moving average
func CheckRateLimit(last time.Time, now time.Time, class RateClass, curAvg int64) (status RateLimitStatus, newAvg int64) {
	delta := now.Sub(last).Milliseconds()

	curAvg = (curAvg*(class.WindowSize-1) + delta) / class.WindowSize

	if curAvg > class.MaxLevel {
		curAvg = class.MaxLevel
	}

	switch {
	case curAvg > class.ClearLevel:
		return RateLimitStatusClear, curAvg
	case curAvg < class.DisconnectLevel:
		return RateLimitStatusDisconnect, curAvg
	case curAvg < class.LimitLevel:
		return RateLimitStatusLimited, curAvg
	case curAvg < class.AlertLevel:
		return RateLimitStatusAlert, curAvg
	}

	return RateLimitStatusClear, curAvg
}
