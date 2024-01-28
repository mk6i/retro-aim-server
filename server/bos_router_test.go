package server

import (
	"io"
	"testing"

	"github.com/mk6i/retro-aim-server/oscar"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestBOSRouter_Route(t *testing.T) {
	tests := []struct {
		name      string
		frame     oscar.SNACFrame
		newRouter func(t *testing.T) BOSRouter
		err       error
	}{
		{
			name: "OService no error",
			frame: oscar.SNACFrame{
				FoodGroup: oscar.OService,
			},
		},
		{
			name: "Locate no error",
			frame: oscar.SNACFrame{
				FoodGroup: oscar.Locate,
			},
		},
		{
			name: "Buddy no error",
			frame: oscar.SNACFrame{
				FoodGroup: oscar.Buddy,
			},
		},
		{
			name: "ICBM no error",
			frame: oscar.SNACFrame{
				FoodGroup: oscar.ICBM,
			},
		},
		{
			name: "ChatNav no error",
			frame: oscar.SNACFrame{
				FoodGroup: oscar.ChatNav,
			},
		},
		{
			name: "Feedbag no error",
			frame: oscar.SNACFrame{
				FoodGroup: oscar.Feedbag,
			},
		},
		{
			name: "Alert no error",
			frame: oscar.SNACFrame{
				FoodGroup: oscar.Alert,
			},
		},
		{
			name: "OService with error",
			frame: oscar.SNACFrame{
				FoodGroup: oscar.OService,
			},
			err: io.EOF,
		},
		{
			name: "Locate with error",
			frame: oscar.SNACFrame{
				FoodGroup: oscar.Locate,
			},
			err: io.EOF,
		},
		{
			name: "Buddy with error",
			frame: oscar.SNACFrame{
				FoodGroup: oscar.Buddy,
			},
			err: io.EOF,
		},
		{
			name: "ICBM with error",
			frame: oscar.SNACFrame{
				FoodGroup: oscar.ICBM,
			},
			err: io.EOF,
		},
		{
			name: "ChatNav with error",
			frame: oscar.SNACFrame{
				FoodGroup: oscar.ChatNav,
			},
			err: io.EOF,
		},
		{
			name: "Feedbag with error",
			frame: oscar.SNACFrame{
				FoodGroup: oscar.Feedbag,
			},
			err: io.EOF,
		},
		{
			name: "Alert with error",
			frame: oscar.SNACFrame{
				FoodGroup: oscar.Alert,
			},
			err: io.EOF,
		},
		{
			name: "Chat (unsupported route), expect error",
			frame: oscar.SNACFrame{
				FoodGroup: oscar.Chat,
			},
			err: ErrUnsupportedSubGroup,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fnNewRouter := func() Router {
				fgRouter := newMockRouter(t)
				fgRouter.EXPECT().
					Route(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(tt.err)
				return fgRouter
			}

			router := BOSRouter{}

			switch tt.frame.FoodGroup {
			case oscar.OService:
				router.OServiceBOSRouter = fnNewRouter()
			case oscar.Locate:
				router.LocateRouter = fnNewRouter()
			case oscar.Buddy:
				router.BuddyRouter = fnNewRouter()
			case oscar.ICBM:
				router.ICBMRouter = fnNewRouter()
			case oscar.ChatNav:
				router.ChatNavRouter = fnNewRouter()
			case oscar.Feedbag:
				router.FeedbagRouter = fnNewRouter()
			case oscar.Alert:
				router.AlertRouter = fnNewRouter()
			}

			err := router.Route(nil, nil, tt.frame, nil, nil, nil)
			assert.ErrorIs(t, err, tt.err)
		})
	}
}
