package wire

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRateLimitClasses(t *testing.T) {
	input := [5]RateClass{
		{ID: 1, WindowSize: 10},
		{ID: 2, WindowSize: 20},
		{ID: 3, WindowSize: 30},
		{ID: 4, WindowSize: 40},
		{ID: 5, WindowSize: 50},
	}

	classes := NewRateLimitClasses(input)

	assert.Equal(t, input, classes.All())
}

func TestCheckRateLimit(t *testing.T) {
	baseTime := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	testCases := []struct {
		name        string
		lastTime    time.Time
		currentTime time.Time
		rateClass   RateClass
		currentAvg  int32
		limitedNow  bool
		wantStatus  RateLimitStatus
		wantNewAvg  int32
	}{
		{
			name:        "Already limited, but newAvg >= ClearLevel => Clear",
			lastTime:    baseTime,
			currentTime: baseTime.Add(50 * time.Millisecond),
			rateClass: RateClass{
				WindowSize:      4,
				MaxLevel:        1000,
				ClearLevel:      10,
				DisconnectLevel: 2,
				LimitLevel:      5,
				AlertLevel:      8,
			},
			currentAvg: 9,    // currentAvg close to ClearLevel
			limitedNow: true, // already limited
			// newAvg = (9*(4-1) + 50) / 4 = (27 + 50) / 4 = 77 / 4 = 19
			// 19 >= ClearLevel(10) => RateLimitStatusClear
			wantStatus: RateLimitStatusClear,
			wantNewAvg: 19,
		},
		{
			name:        "Already limited, but newAvg < ClearLevel => Remain Limited",
			lastTime:    baseTime,
			currentTime: baseTime.Add(30 * time.Millisecond),
			rateClass: RateClass{
				WindowSize:      4,
				MaxLevel:        1000,
				ClearLevel:      50,
				DisconnectLevel: 10,
				LimitLevel:      20,
				AlertLevel:      30,
			},
			currentAvg: 10,
			limitedNow: true,
			// newAvg = (10*(4-1) + 30) / 4 = (30 + 30) / 4 = 60 / 4 = 15
			// 15 < ClearLevel(50) => remain limited
			wantStatus: RateLimitStatusLimited,
			wantNewAvg: 15,
		},
		{
			name:        "Not Limited Now, New average < DisconnectLevel => Disconnect",
			lastTime:    baseTime,
			currentTime: baseTime.Add(10 * time.Millisecond),
			rateClass: RateClass{
				WindowSize:      4,
				MaxLevel:        1000,
				ClearLevel:      100,
				DisconnectLevel: 5,
				LimitLevel:      20,
				AlertLevel:      40,
			},
			currentAvg: 1,
			limitedNow: false,
			// newAvg = (1*(4-1) + 10) / 4 = (3 + 10) / 4 = 13 / 4 = 3
			// 3 < DisconnectLevel(5) => RateLimitStatusDisconnect
			wantStatus: RateLimitStatusDisconnect,
			wantNewAvg: 3,
		},
		{
			name:        "Limited Now, New average < DisconnectLevel => Disconnect",
			lastTime:    baseTime,
			currentTime: baseTime.Add(10 * time.Millisecond),
			rateClass: RateClass{
				WindowSize:      4,
				MaxLevel:        1000,
				ClearLevel:      100,
				DisconnectLevel: 5,
				LimitLevel:      20,
				AlertLevel:      40,
			},
			currentAvg: 1,
			limitedNow: true,
			// newAvg = (1*(4-1) + 10) / 4 = (3 + 10) / 4 = 13 / 4 = 3
			// 3 < DisconnectLevel(5) => RateLimitStatusDisconnect
			wantStatus: RateLimitStatusDisconnect,
			wantNewAvg: 3,
		},
		{
			name:        "New average < LimitLevel => Limited",
			lastTime:    baseTime,
			currentTime: baseTime.Add(20 * time.Millisecond),
			rateClass: RateClass{
				WindowSize:      4,
				MaxLevel:        1000,
				ClearLevel:      100,
				DisconnectLevel: 5,
				LimitLevel:      40,
				AlertLevel:      60,
			},
			currentAvg: 10,
			limitedNow: false,
			// newAvg = (10*(4-1) + 20) / 4 = (30 + 20) / 4 = 50 / 4 = 12
			// 12 < LimitLevel(40) => RateLimitStatusLimited
			wantStatus: RateLimitStatusLimited,
			wantNewAvg: 12,
		},
		{
			name:        "New average < AlertLevel => Alert",
			lastTime:    baseTime,
			currentTime: baseTime.Add(30 * time.Millisecond),
			rateClass: RateClass{
				WindowSize:      4,
				MaxLevel:        1000,
				ClearLevel:      100,
				DisconnectLevel: 5,
				LimitLevel:      20,
				AlertLevel:      40,
			},
			currentAvg: 20,
			limitedNow: false,
			// newAvg = (20*(4-1) + 30) / 4 = (60 + 30) / 4 = 90 / 4 = 22
			// 22 >= 20 => not "Limited"; 22 < 40 => "Alert"
			wantStatus: RateLimitStatusAlert,
			wantNewAvg: 22,
		},
		{
			name:        "New average >= AlertLevel => Clear",
			lastTime:    baseTime,
			currentTime: baseTime.Add(50 * time.Millisecond),
			rateClass: RateClass{
				WindowSize:      4,
				MaxLevel:        1000,
				ClearLevel:      100,
				DisconnectLevel: 5,
				LimitLevel:      20,
				AlertLevel:      40,
			},
			// Choose 39 so the resulting newAvg is 41, which is >= AlertLevel.
			currentAvg: 39,
			limitedNow: false,
			// newAvg = (39*(4-1) + 50) / 4 = (117 + 50) / 4 = 167 / 4 = 41
			// 41 >= AlertLevel(40) => RateLimitStatusClear
			wantStatus: RateLimitStatusClear,
			wantNewAvg: 41,
		},
		{
			name:        "Clamp newAvg to MaxLevel if exceeded",
			lastTime:    baseTime,
			currentTime: baseTime.Add(9999 * time.Millisecond),
			rateClass: RateClass{
				WindowSize:      4,
				MaxLevel:        100,
				ClearLevel:      80,
				DisconnectLevel: 20,
				LimitLevel:      40,
				AlertLevel:      60,
			},
			currentAvg: 95,
			limitedNow: false,
			// Without clamping, newAvg would be huge:
			// newAvg = (95*(4-1) + 9999) / 4 = (285 + 9999)/4 = 10284/4 = 2571
			// Clamped to 100 => 100 >= AlertLevel(60) => RateLimitStatusClear
			wantStatus: RateLimitStatusClear,
			wantNewAvg: 100,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gotStatus, gotNewAvg := CheckRateLimit(
				tc.lastTime,
				tc.currentTime,
				tc.rateClass,
				tc.currentAvg,
				tc.limitedNow,
			)

			assert.Equal(t, tc.wantStatus, gotStatus)
			assert.Equal(t, tc.wantNewAvg, gotNewAvg)
		})
	}
}

func TestRateLimitClasses_Get(t *testing.T) {
	classes := DefaultRateLimitClasses()

	// Test Get() returns correct class for each ID
	for i := 1; i <= 5; i++ {
		id := RateLimitClassID(i)
		class := classes.Get(id)
		assert.Equal(t, id, class.ID)
		assert.Equal(t, classes.All()[i-1], class)
	}
}

func TestRateLimitClasses_All(t *testing.T) {
	classes := DefaultRateLimitClasses()

	// Test All() returns exactly 5 classes with correct IDs
	all := classes.All()
	assert.Len(t, all, 5)
	for i, class := range all {
		expectedID := RateLimitClassID(i + 1)
		assert.Equal(t, expectedID, class.ID, "class ID mismatch at index %d", i)
	}
}

func TestSNACRateLimits_RateClassLookup(t *testing.T) {
	limits := DefaultSNACRateLimits()

	testCases := []struct {
		foodGroup uint16
		subGroup  uint16
		expected  RateLimitClassID
		found     bool
	}{
		{Chat, ChatUsersJoined, 1, true},
		{Chat, ChatChannelMsgToHost, 2, true},
		{0xFFFF, 0x0001, 0, false},
		{Chat, 0xFFFF, 0, false},
	}

	for _, tc := range testCases {
		classID, ok := limits.RateClassLookup(tc.foodGroup, tc.subGroup)
		assert.Equal(t, tc.found, ok)
		assert.Equal(t, tc.expected, classID)
	}
}

func TestSNACRateLimits_All(t *testing.T) {
	limits := DefaultSNACRateLimits()

	seen := map[uint16]map[uint16]RateLimitClassID{}
	for entry := range limits.All() {
		if _, ok := seen[entry.FoodGroup]; !ok {
			seen[entry.FoodGroup] = map[uint16]RateLimitClassID{}
		}
		seen[entry.FoodGroup][entry.SubGroup] = entry.RateLimitClass
	}

	// Spot-check a few values
	require.Contains(t, seen, ICBM)
	assert.Equal(t, RateLimitClassID(3), seen[ICBM][ICBMChannelMsgToHost])
	assert.Equal(t, RateLimitClassID(1), seen[ICBM][ICBMChannelMsgToClient])

	require.Contains(t, seen, Locate)
	assert.Equal(t, RateLimitClassID(4), seen[Locate][LocateSetDirInfo])
	assert.Equal(t, RateLimitClassID(3), seen[Locate][LocateUserInfoQuery])
}

func TestSNACRateLimits_All_YieldStopsEarly(t *testing.T) {
	limits := DefaultSNACRateLimits()

	count := 0
	limits.All()(func(entry struct {
		FoodGroup      uint16
		SubGroup       uint16
		RateLimitClass RateLimitClassID
	}) bool {
		count++
		// stop iteration after first item to trigger `if !yield(...) { return }`
		return false
	})

	// Should only yield one entry
	assert.Equal(t, 1, count)
}
