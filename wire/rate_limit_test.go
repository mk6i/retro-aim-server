package wire

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCheckRateLimit(t *testing.T) {
	type args struct {
		last       time.Time
		now        time.Time
		class      RateClass
		currentAvg int64
	}
	tests := []struct {
		name  string
		args  args
		want  RateLimitStatus
		want1 int64
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := CheckRateLimit(tt.args.last, tt.args.now, tt.args.class, tt.args.currentAvg)
			assert.Equalf(t, tt.want, got, "CheckRateLimit(%v, %v, %v, %v)", tt.args.last, tt.args.now, tt.args.class, tt.args.currentAvg)
			assert.Equalf(t, tt.want1, got1, "CheckRateLimit(%v, %v, %v, %v)", tt.args.last, tt.args.now, tt.args.class, tt.args.currentAvg)
		})
	}
}
