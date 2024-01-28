package server

import (
	"io"
	"testing"

	"github.com/mk6i/retro-aim-server/oscar"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestChatServiceRouter_Route(t *testing.T) {
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
			name: "Chat no error",
			frame: oscar.SNACFrame{
				FoodGroup: oscar.Chat,
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
			name: "Chat with error",
			frame: oscar.SNACFrame{
				FoodGroup: oscar.Chat,
			},
			err: io.EOF,
		},
		{
			name: "ICBM (unsupported route), expect error",
			frame: oscar.SNACFrame{
				FoodGroup: oscar.ICBM,
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

			router := ChatServiceRouter{}

			switch tt.frame.FoodGroup {
			case oscar.OService:
				router.OServiceChatRouter = fnNewRouter()
			case oscar.Chat:
				router.ChatRouter = fnNewRouter()
			}

			err := router.Route(nil, nil, tt.frame, nil, nil, nil)
			assert.ErrorIs(t, err, tt.err)
		})
	}
}
