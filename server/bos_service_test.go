package server

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"sync"
	"testing"

	"github.com/mkaminski/goaim/state"
	"github.com/stretchr/testify/mock"

	"github.com/mkaminski/goaim/oscar"
	"github.com/stretchr/testify/assert"
)

// pipeRWC provides a mock for ReadWriteCloser that uses pipes instead of TCP
// connections
type pipeRWC struct {
	*io.PipeReader
	*io.PipeWriter
}

func (m pipeRWC) Close() error {
	if err := m.PipeReader.Close(); err != nil {
		return err
	}
	return m.PipeWriter.Close()
}

func TestBOSService_handleNewConnection(t *testing.T) {
	sess := state.NewSession()
	sess.SetID("login-cookie-1234")

	clientReader, serverWriter := io.Pipe()
	serverReader, clientWriter := io.Pipe()

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		// < receive FLAPSignonFrame
		flap := oscar.FLAPFrame{}
		assert.NoError(t, oscar.Unmarshal(&flap, serverReader))
		buf, err := flap.SNACBuffer(serverReader)
		assert.NoError(t, err)
		flapSignonFrame := oscar.FLAPSignonFrame{}
		assert.NoError(t, oscar.Unmarshal(&flapSignonFrame, buf))

		// > send FLAPSignonFrame
		flapSignonFrame = oscar.FLAPSignonFrame{
			FLAPVersion: 1,
		}
		flapSignonFrame.Append(oscar.NewTLV(oscar.OServiceTLVTagsLoginCookie, []byte(sess.ID())))
		buf = &bytes.Buffer{}
		assert.NoError(t, oscar.Marshal(flapSignonFrame, buf))
		flap = oscar.FLAPFrame{
			StartMarker:   42,
			FrameType:     oscar.FLAPFrameSignon,
			PayloadLength: uint16(buf.Len()),
		}
		assert.NoError(t, oscar.Marshal(flap, serverWriter))
		_, err = serverWriter.Write(buf.Bytes())
		assert.NoError(t, err)

		// < receive SNAC_0x01_0x03_OServiceHostOnline
		flap = oscar.FLAPFrame{}
		assert.NoError(t, oscar.Unmarshal(&flap, serverReader))
		buf, err = flap.SNACBuffer(serverReader)
		assert.NoError(t, err)
		frame := oscar.SNACFrame{}
		assert.NoError(t, oscar.Unmarshal(&frame, buf))
		body := oscar.SNAC_0x01_0x03_OServiceHostOnline{}
		assert.NoError(t, oscar.Unmarshal(&body, buf))

		// send the first request that should get relayed to BOSRouter.Route
		var seq uint32
		frame = oscar.SNACFrame{
			FoodGroup: oscar.OService,
			SubGroup:  oscar.OServiceClientOnline,
		}
		assert.NoError(t, sendSNAC(frame, struct{}{}, &seq, serverWriter))

		assert.NoError(t, serverWriter.Close())
	}()

	authHandler := newMockAuthHandler(t)
	authHandler.EXPECT().
		RetrieveBOSSession(sess.ID()).
		Return(sess, nil)
	authHandler.EXPECT().
		Signout(mock.Anything, sess).
		Run(func(ctx context.Context, sess *state.Session) {
			wg.Done()
		}).
		Return(nil)

	bosHandler := newMockOServiceBOSHandler(t)
	bosHandler.EXPECT().
		WriteOServiceHostOnline().
		Return(oscar.SNACMessage{
			Frame: oscar.SNACFrame{
				FoodGroup: oscar.OService,
				SubGroup:  oscar.OServiceHostOnline,
			},
			Body: oscar.SNAC_0x01_0x03_OServiceHostOnline{},
		})

	bosRouter := newMockBOSRouter(t)
	bosRouter.EXPECT().
		Route(mock.Anything, sess, mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	rt := BOSService{
		AuthHandler:       authHandler,
		OServiceBOSRouter: NewOServiceRouterForBOS(slog.Default(), nil, bosHandler),
		BOSRouter:         bosRouter,
	}
	rwc := pipeRWC{
		PipeReader: clientReader,
		PipeWriter: clientWriter,
	}
	rt.handleNewConnection(context.Background(), rwc)

	wg.Wait() // wait for server to drain the connection
}
