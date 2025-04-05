package wire

import (
	"fmt"
	"testing"
	"time"
)

func TestCheckRateLimit(t *testing.T) {
	class := class5
	curLevel := class.MaxLevel
	last := time.Now()
	now := last

	for i := 0; i < 500; i++ {
		var status RateLimitStatus
		status, curLevel = CheckRateLimit(last, now, class, curLevel)
		fmt.Printf("Delta: %dms Iteration: %d Time: %s Level: %d Status: ", now.Sub(last).Milliseconds(), i, now.Format("15:04:05"), curLevel)
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
		now = now.Add(1 * time.Second)
	}
}
