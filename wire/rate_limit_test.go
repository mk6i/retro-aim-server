package wire

import (
	"fmt"
	"testing"
	"time"
)

func TestCheckRateLimit(t *testing.T) {
	class := class4
	curLevel := class.MaxLevel
	last := time.Now()
	now := time.Now()

	for i := 0; i < 100; i++ {
		var status RateLimitStatus
		status, curLevel = CheckRateLimit(last, now, class, curLevel)
		fmt.Printf("Iteration: %d Time: %s Level: %d Status: ", i, now.Format("15:04:05"), curLevel)
		switch status {
		case RateLimitStatusLimited:
			fmt.Println("limited")
		case RateLimitStatusAlert:
			fmt.Println("alert")
		case RateLimitStatusClear:
			fmt.Println("clear")
		case RateLimitStatusDisconnect:
			fmt.Println("disconnect")
		}
		last = now
		now = now.Add(10 * time.Millisecond)
	}
}
