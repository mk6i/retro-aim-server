package wire

import "time"

type RateClass struct {
	ID              uint16
	WindowSize      int64
	ClearLevel      int64
	AlertLevel      int64
	LimitLevel      int64
	DisconnectLevel int64
	CurrentLevel    int64
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
	ID:              0x01,
	WindowSize:      0x0050,
	ClearLevel:      0x09C4,
	AlertLevel:      0x07D0,
	LimitLevel:      0x05DC,
	DisconnectLevel: 0x0320,
	CurrentLevel:    0x0D69,
	MaxLevel:        0x1770,
}

func CheckRateLimit(last time.Time, now time.Time, class RateClass, currentAvg int64) (RateLimitStatus, int64) {
	delta := last.Sub(now).Milliseconds()

	//   NewLevel = (Window - 1)/Window * OldLevel + 1/Window * CurrentTimeDiff
	currentAvg = (class.WindowSize-1)/class.WindowSize*currentAvg + 1/class.WindowSize*delta

	if currentAvg > class.MaxLevel {
		currentAvg = class.MaxLevel
	}

	switch {
	case currentAvg > class.ClearLevel:
		return RateLimitStatusClear, currentAvg
	case currentAvg < class.DisconnectLevel:
		return RateLimitStatusDisconnect, currentAvg
	case currentAvg < class.LimitLevel:
		return RateLimitStatusLimited, currentAvg
	case currentAvg < class.AlertLevel:
		return RateLimitStatusAlert, currentAvg
	}

	return RateLimitStatusClear, currentAvg
}
