package server

import (
	"bytes"
	"os"
	"testing"

	"github.com/mkaminski/goaim/oscar"
	"github.com/stretchr/testify/assert"
)

func TestReceiveAndSendAuthChallenge_OK(t *testing.T) {
	cfg := Config{}

	const testFile string = "aim_test.db"
	defer func() {
		if err := os.Remove(testFile); err != nil {
			assert.NoError(t, err)
		}
	}()

	fs, err := NewFeedbagStore(testFile)
	if err != nil {
		assert.NoError(t, err)
	}

	newUser := User{
		ScreenName: "the_screen_name",
		AuthKey:    "the_auth_key",
	}
	if err := newUser.HashPassword("the_password"); err != nil {
		assert.NoError(t, err)
	}
	if err := fs.UpsertUser(newUser); err != nil {
		assert.NoError(t, err)
	}

	input := &bytes.Buffer{}

	snacPayloadOut := oscar.SNAC_0x17_0x06_BUCPChallengeRequest{}
	snacPayloadOut.AddTLV(oscar.TLV{TType: TLVScreenName, Val: newUser.ScreenName})

	var seq uint32
	writeOutSNAC(oscar.SnacFrame{}, oscar.SnacFrame{}, snacPayloadOut, &seq, input)

	output := &bytes.Buffer{}

	if err := ReceiveAndSendAuthChallenge(cfg, fs, input, output, &seq); err != nil {
		assert.NoError(t, err)
	}

	flap := oscar.FlapFrame{}
	if err := oscar.Unmarshal(&flap, output); err != nil {
		assert.NoError(t, err)
	}
	snac := oscar.SnacFrame{}
	if err := oscar.Unmarshal(&snac, output); err != nil {
		assert.NoError(t, err)
	}
	expectSnacFrame := oscar.SnacFrame{
		FoodGroup: BUCP,
		SubGroup:  BUCPChallengeResponse,
	}
	assert.Equal(t, expectSnacFrame, snac)

	snacPayload := oscar.SNAC_0x17_0x07_BUCPChallengeResponse{}
	if err := oscar.Unmarshal(&snacPayload, output); err != nil {
		assert.NoError(t, err)
	}
	assert.Equal(t, newUser.AuthKey, snacPayload.AuthKey)
}

func TestReceiveAndSendAuthChallenge_BadUser_DisableAuth(t *testing.T) {
	cfg := Config{
		DisableAuth: true,
	}

	const testFile string = "aim_test.db"
	defer func() {
		if err := os.Remove(testFile); err != nil {
			assert.NoError(t, err)
		}
	}()

	fs, err := NewFeedbagStore(testFile)
	if err != nil {
		assert.NoError(t, err)
	}

	input := &bytes.Buffer{}

	snacPayloadOut := oscar.SNAC_0x17_0x06_BUCPChallengeRequest{}
	snacPayloadOut.AddTLV(oscar.TLV{TType: TLVScreenName, Val: "bad-user"})

	var seq uint32
	writeOutSNAC(oscar.SnacFrame{}, oscar.SnacFrame{}, snacPayloadOut, &seq, input)

	output := &bytes.Buffer{}

	if err := ReceiveAndSendAuthChallenge(cfg, fs, input, output, &seq); err != nil {
		assert.NoError(t, err)
	}

	flap := oscar.FlapFrame{}
	if err := oscar.Unmarshal(&flap, output); err != nil {
		assert.NoError(t, err)
	}
	snac := oscar.SnacFrame{}
	if err := oscar.Unmarshal(&snac, output); err != nil {
		assert.NoError(t, err)
	}
	expectSnacFrame := oscar.SnacFrame{
		FoodGroup: BUCP,
		SubGroup:  BUCPChallengeResponse,
	}
	assert.Equal(t, expectSnacFrame, snac)

	snacPayload := oscar.SNAC_0x17_0x07_BUCPChallengeResponse{}
	if err := oscar.Unmarshal(&snacPayload, output); err != nil {
		assert.NoError(t, err)
	}
	assert.NotEmpty(t, snacPayload.AuthKey)
}

func TestReceiveAndSendAuthChallenge_BadUser_EnableAuth(t *testing.T) {
	cfg := Config{}

	const testFile string = "aim_test.db"
	defer func() {
		if err := os.Remove(testFile); err != nil {
			assert.NoError(t, err)
		}
	}()

	fs, err := NewFeedbagStore(testFile)
	if err != nil {
		assert.NoError(t, err)
	}

	input := &bytes.Buffer{}

	snacPayloadOut := oscar.SNAC_0x17_0x06_BUCPChallengeRequest{}
	snacPayloadOut.AddTLV(oscar.TLV{TType: TLVScreenName, Val: "bad-user"})

	var seq uint32
	writeOutSNAC(oscar.SnacFrame{}, oscar.SnacFrame{}, snacPayloadOut, &seq, input)

	output := &bytes.Buffer{}

	if err := ReceiveAndSendAuthChallenge(cfg, fs, input, output, &seq); err != nil {
		assert.NoError(t, err)
	}

	flap := oscar.FlapFrame{}
	if err := oscar.Unmarshal(&flap, output); err != nil {
		assert.NoError(t, err)
	}
	snac := oscar.SnacFrame{}
	if err := oscar.Unmarshal(&snac, output); err != nil {
		assert.NoError(t, err)
	}
	expectSnacFrame := oscar.SnacFrame{
		FoodGroup: BUCP,
		SubGroup:  BUCPLoginResponse,
	}
	assert.Equal(t, expectSnacFrame, snac)

	snacPayload := oscar.SNAC_0x17_0x03_BUCPLoginResponse{}
	if err := oscar.Unmarshal(&snacPayload, output); err != nil {
		assert.NoError(t, err)
	}
	_, hasError := snacPayload.GetTLV(TLVErrorSubcode)
	assert.True(t, hasError)
}

func TestReceiveAndSendBUCPLoginRequest_OK(t *testing.T) {
	cfg := Config{}
	sm := NewSessionManager()

	const testFile string = "aim_test.db"
	defer func() {
		if err := os.Remove(testFile); err != nil {
			assert.NoError(t, err)
		}
	}()

	fs, err := NewFeedbagStore(testFile)
	if err != nil {
		assert.NoError(t, err)
	}

	newUser := User{
		ScreenName: "the_screen_name",
		AuthKey:    "the_auth_key",
	}
	if err := newUser.HashPassword("the_password"); err != nil {
		assert.NoError(t, err)
	}
	if err := fs.UpsertUser(newUser); err != nil {
		assert.NoError(t, err)
	}

	input := &bytes.Buffer{}

	snacPayloadOut := oscar.SNAC_0x17_0x02_BUCPLoginRequest{}
	snacPayloadOut.AddTLV(oscar.TLV{TType: TLVPasswordHash, Val: newUser.PassHash})
	snacPayloadOut.AddTLV(oscar.TLV{TType: TLVScreenName, Val: newUser.ScreenName})

	var seq uint32
	writeOutSNAC(oscar.SnacFrame{}, oscar.SnacFrame{}, snacPayloadOut, &seq, input)

	output := &bytes.Buffer{}

	if err := ReceiveAndSendBUCPLoginRequest(cfg, sm, fs, input, output, &seq); err != nil {
		assert.NoError(t, err)
	}

	flap := oscar.FlapFrame{}
	if err := oscar.Unmarshal(&flap, output); err != nil {
		assert.NoError(t, err)
	}
	snac := oscar.SnacFrame{}
	if err := oscar.Unmarshal(&snac, output); err != nil {
		assert.NoError(t, err)
	}
	expectSnacFrame := oscar.SnacFrame{
		FoodGroup: BUCP,
		SubGroup:  BUCPLoginResponse,
	}
	assert.Equal(t, expectSnacFrame, snac)

	snacPayload := oscar.SNAC_0x17_0x03_BUCPLoginResponse{}
	if err := oscar.Unmarshal(&snacPayload, output); err != nil {
		assert.NoError(t, err)
	}
	_, hasError := snacPayload.GetTLV(TLVErrorSubcode)
	assert.False(t, hasError)
}

func TestReceiveAndSendBUCPLoginRequest_BadUser_EnableAuth(t *testing.T) {
	cfg := Config{}
	sm := NewSessionManager()

	const testFile string = "aim_test.db"
	defer func() {
		if err := os.Remove(testFile); err != nil {
			assert.NoError(t, err)
		}
	}()

	fs, err := NewFeedbagStore(testFile)
	if err != nil {
		assert.NoError(t, err)
	}

	newUser := User{
		ScreenName: "the_screen_name",
		AuthKey:    "the_auth_key",
	}
	if err := newUser.HashPassword("the_password"); err != nil {
		assert.NoError(t, err)
	}
	if err := fs.UpsertUser(newUser); err != nil {
		assert.NoError(t, err)
	}

	input := &bytes.Buffer{}

	snacPayloadOut := oscar.SNAC_0x17_0x02_BUCPLoginRequest{}
	snacPayloadOut.AddTLV(oscar.TLV{TType: TLVPasswordHash, Val: newUser.PassHash})
	snacPayloadOut.AddTLV(oscar.TLV{TType: TLVScreenName, Val: "bad-screen-name"})

	var seq uint32
	writeOutSNAC(oscar.SnacFrame{}, oscar.SnacFrame{}, snacPayloadOut, &seq, input)

	output := &bytes.Buffer{}

	if err := ReceiveAndSendBUCPLoginRequest(cfg, sm, fs, input, output, &seq); err != nil {
		assert.NoError(t, err)
	}

	flap := oscar.FlapFrame{}
	if err := oscar.Unmarshal(&flap, output); err != nil {
		assert.NoError(t, err)
	}
	snac := oscar.SnacFrame{}
	if err := oscar.Unmarshal(&snac, output); err != nil {
		assert.NoError(t, err)
	}
	expectSnacFrame := oscar.SnacFrame{
		FoodGroup: BUCP,
		SubGroup:  BUCPLoginResponse,
	}
	assert.Equal(t, expectSnacFrame, snac)

	snacPayload := oscar.SNAC_0x17_0x02_BUCPLoginRequest{}
	if err := oscar.Unmarshal(&snacPayload, output); err != nil {
		assert.NoError(t, err)
	}
	_, hasError := snacPayload.GetTLV(TLVErrorSubcode)
	assert.True(t, hasError)
}

func TestReceiveAndSendBUCPLoginRequest_BadUser_DisableAuth(t *testing.T) {
	cfg := Config{
		DisableAuth: true,
	}
	sm := NewSessionManager()

	const testFile string = "aim_test.db"
	defer func() {
		if err := os.Remove(testFile); err != nil {
			assert.NoError(t, err)
		}
	}()

	fs, err := NewFeedbagStore(testFile)
	if err != nil {
		assert.NoError(t, err)
	}

	newUser := User{
		ScreenName: "the_screen_name",
		AuthKey:    "the_auth_key",
	}
	if err := newUser.HashPassword("the_password"); err != nil {
		assert.NoError(t, err)
	}
	if err := fs.UpsertUser(newUser); err != nil {
		assert.NoError(t, err)
	}

	input := &bytes.Buffer{}

	snacPayloadOut := oscar.SNAC_0x17_0x02_BUCPLoginRequest{}
	snacPayloadOut.AddTLV(oscar.TLV{TType: TLVPasswordHash, Val: newUser.PassHash})
	snacPayloadOut.AddTLV(oscar.TLV{TType: TLVScreenName, Val: "bad-screen-name"})

	var seq uint32
	writeOutSNAC(oscar.SnacFrame{}, oscar.SnacFrame{}, snacPayloadOut, &seq, input)

	output := &bytes.Buffer{}

	if err := ReceiveAndSendBUCPLoginRequest(cfg, sm, fs, input, output, &seq); err != nil {
		assert.NoError(t, err)
	}

	flap := oscar.FlapFrame{}
	if err := oscar.Unmarshal(&flap, output); err != nil {
		assert.NoError(t, err)
	}
	snac := oscar.SnacFrame{}
	if err := oscar.Unmarshal(&snac, output); err != nil {
		assert.NoError(t, err)
	}
	expectSnacFrame := oscar.SnacFrame{
		FoodGroup: BUCP,
		SubGroup:  BUCPLoginResponse,
	}
	assert.Equal(t, expectSnacFrame, snac)

	snacPayload := oscar.SNAC_0x17_0x02_BUCPLoginRequest{}
	if err := oscar.Unmarshal(&snacPayload, output); err != nil {
		assert.NoError(t, err)
	}
	_, hasError := snacPayload.GetTLV(TLVErrorSubcode)
	assert.False(t, hasError)
}

func TestReceiveAndSendBUCPLoginRequest_BadPassword(t *testing.T) {
	cfg := Config{}
	sm := NewSessionManager()

	const testFile string = "aim_test.db"
	defer func() {
		if err := os.Remove(testFile); err != nil {
			assert.NoError(t, err)
		}
	}()

	fs, err := NewFeedbagStore(testFile)
	if err != nil {
		assert.NoError(t, err)
	}

	newUser := User{
		ScreenName: "the_screen_name",
		AuthKey:    "the_auth_key",
	}
	if err := newUser.HashPassword("the_password"); err != nil {
		assert.NoError(t, err)
	}
	if err := fs.UpsertUser(newUser); err != nil {
		assert.NoError(t, err)
	}

	input := &bytes.Buffer{}

	snacPayloadOut := oscar.SNAC_0x17_0x02_BUCPLoginRequest{}
	snacPayloadOut.AddTLV(oscar.TLV{TType: TLVPasswordHash, Val: []byte("bad_password")})
	snacPayloadOut.AddTLV(oscar.TLV{TType: TLVScreenName, Val: newUser.ScreenName})

	var seq uint32
	writeOutSNAC(oscar.SnacFrame{}, oscar.SnacFrame{}, snacPayloadOut, &seq, input)

	output := &bytes.Buffer{}

	if err := ReceiveAndSendBUCPLoginRequest(cfg, sm, fs, input, output, &seq); err != nil {
		assert.NoError(t, err)
	}

	flap := oscar.FlapFrame{}
	if err := oscar.Unmarshal(&flap, output); err != nil {
		assert.NoError(t, err)
	}

	snac := oscar.SnacFrame{}
	if err := oscar.Unmarshal(&snac, output); err != nil {
		assert.NoError(t, err)
	}
	expectSnacFrame := oscar.SnacFrame{
		FoodGroup: BUCP,
		SubGroup:  BUCPLoginResponse,
	}
	assert.Equal(t, expectSnacFrame, snac)

	snacPayload := oscar.SNAC_0x17_0x02_BUCPLoginRequest{}
	if err := oscar.Unmarshal(&snacPayload, output); err != nil {
		assert.NoError(t, err)
	}
	_, hasError := snacPayload.GetTLV(TLVErrorSubcode)
	assert.True(t, hasError)
}
