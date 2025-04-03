package wire

import "time"

type RateClass struct {
	ID              uint16
	WindowSize      uint32
	ClearLevel      uint32
	AlertLevel      uint32
	LimitLevel      uint32
	DisconnectLevel uint32
	CurrentLevel    uint32
	MaxLevel        uint32
	V2Params        *struct {
		LastTime     uint32
		CurrentState uint8
	} `oscar:"optional"`
}

type RateLimitStatus int

const (
	RateLimitStatusLimited    RateLimitStatus = 1 // You're currently rate-limited
	RateLimitStatusAlert      RateLimitStatus = 2 // You're close to being rate-limited
	RateLimitStatusClear      RateLimitStatus = 3 // You're under the rate limit; all good
	RateLimitStatusDisconnect RateLimitStatus = 4 // You're under the rate limit; all good
)

func CheckRateLimit(last time.Time, now time.Time, class RateClass, currentAvg time.Duration) (state RateLimitStatus, avg time.Duration) {
	delta := last.Sub(now)
	currentAvg = (currentAvg*time.Duration(class.WindowSize-1) + delta) / time.Duration(class.WindowSize)

	if currentAvg > time.Duration(class.MaxLevel)*time.Millisecond {
		currentAvg = time.Duration(class.MaxLevel)
	}

	switch {
	case currentAvg > time.Duration(class.ClearLevel)*time.Millisecond:
		return RateLimitStatusClear, avg
	case currentAvg < time.Duration(class.DisconnectLevel)*time.Millisecond:
		return RateLimitStatusDisconnect, avg
	case currentAvg < time.Duration(class.LimitLevel)*time.Millisecond:
		return RateLimitStatusLimited, avg
	case currentAvg < time.Duration(class.AlertLevel)*time.Millisecond:
		return RateLimitStatusAlert, avg
	}

	return RateLimitStatusClear, currentAvg
}
