package toc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_parseArgs(t *testing.T) {
	type testCase struct {
		name         string
		givenPayload string
		givenCmd     string
		givenArgs    []*string
		wantVarArgs  []string
		wantArgs     []string
		wantErrMsg   string
	}

	tests := []testCase{
		{
			name:         "no positional args or varargs",
			givenPayload: `toc_chat_invite`,
			givenCmd:     "toc_chat_invite",
			givenArgs:    nil,
			wantVarArgs:  nil,
		},
		{
			name:         "positional args with varargs",
			givenPayload: `toc_chat_invite 1234 "Join me!" user1 user2 user3`,
			givenCmd:     "toc_chat_invite",
			givenArgs:    []*string{new(string), new(string)},
			wantVarArgs:  []string{"user1", "user2", "user3"},
			wantArgs:     []string{"1234", "Join me!"},
		},
		{
			name:         "nil positional argument placeholders should get skipped",
			givenPayload: `toc_chat_invite 1234 "Join me!" user1 user2 user3`,
			givenCmd:     "toc_chat_invite",
			givenArgs:    []*string{nil, nil}, // still 2 placeholders, both nil
			wantVarArgs:  []string{"user1", "user2", "user3"},
			wantArgs:     []string{"", ""},
		},
		{
			name:         "positional args with no varargs",
			givenPayload: `toc_chat_invite 1234 "Join me!"`,
			givenCmd:     "toc_chat_invite",
			givenArgs:    []*string{new(string), new(string)}, // roomID + msg
			wantVarArgs:  nil,
			wantArgs:     []string{"1234", "Join me!"},
		},
		{
			name:         "varargs only",
			givenPayload: `toc_chat_invite user1 user2 user3`,
			givenCmd:     "toc_chat_invite",
			givenArgs:    nil,
			wantVarArgs:  []string{"user1", "user2", "user3"},
		},
		{
			name:         "command mismatch",
			givenPayload: `toc_chat_invite user1 user2 user3`,
			givenCmd:     "toc_chat_accept",
			givenArgs:    nil,
			wantVarArgs:  nil,
			wantErrMsg:   "mismatch",
		},
		{
			name:         "too many positional arg placeholders",
			givenPayload: `toc_chat_invite`,
			givenCmd:     "toc_chat_invite",
			givenArgs:    []*string{new(string), new(string)},
			wantVarArgs:  nil,
			wantErrMsg:   "command contains fewer arguments than expected",
		},
		{
			name:         "CSV parser error",
			givenPayload: ``,
			givenCmd:     "toc_chat_invite",
			givenArgs:    []*string{nil},
			wantVarArgs:  nil,
			wantErrMsg:   "CSV reader error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			varArgs, err := parseArgs([]byte(tt.givenPayload), tt.givenCmd, tt.givenArgs...)

			if tt.wantErrMsg != "" {
				assert.ErrorContains(t, err, tt.wantErrMsg)
				return
			}

			assert.NoError(t, err)

			// verify the placeholder pointers got populated
			for i, want := range tt.wantArgs {
				if want == "" {
					assert.Nil(t, tt.givenArgs[i])
				} else {
					got := *tt.givenArgs[i]
					assert.Equal(t, want, got)
				}
			}

			// verify we have the same varargs
			assert.Equal(t, tt.wantVarArgs, varArgs)
			assert.Equal(t, len(tt.wantArgs), len(tt.givenArgs))

		})
	}
}
