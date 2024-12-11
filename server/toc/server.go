package toc

import (
	"bufio"
	"bytes"
	"context"
	"encoding/csv"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"sync"
	"time"

	"github.com/mk6i/retro-aim-server/wire"
)

// Server provides client connection lifecycle management for the BOS
// service.
type Server struct {
	ListenAddr string
	Logger     *slog.Logger
}

// Start starts a TCP server and listens for connections. The initial
// authentication handshake sequences are handled by this method. The remaining
// requests are relayed to BOSRouter.
func (rt Server) Start(ctx context.Context) error {
	listener, err := net.Listen("tcp", rt.ListenAddr)
	if err != nil {
		return fmt.Errorf("unable to start TOC server: %w", err)
	}

	rt.Logger.Info("starting server", "listen_host", rt.ListenAddr)

	go func() {
		<-ctx.Done()
		listener.Close()
	}()

	wg := sync.WaitGroup{}
	for {
		conn, err := listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				break
			}
			rt.Logger.Error("accept failed", "err", err.Error())
			continue
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			connCtx := context.WithValue(ctx, "ip", conn.RemoteAddr().String())
			//rt.Logger.DebugContext(connCtx, "accepted connection")
			if err := rt.handleNewConnection(connCtx, conn); err != nil {
				rt.Logger.Info("user session failed", "err", err.Error())
			}
		}()
	}

	if !waitForShutdown(&wg) {
		rt.Logger.Error("shutdown complete, but connections didn't close cleanly")
	} else {
		rt.Logger.Info("shutdown complete")
	}

	return nil
}

func (rt Server) handleNewConnection(ctx context.Context, rwc io.ReadWriteCloser) error {

	go func() {
		<-ctx.Done()
		rwc.Close()
	}()
	reader := bufio.NewReader(rwc)

	flap := wire.NewFlapClient(0, rwc, rwc)

	line, _, err := reader.ReadLine()
	if err != nil {
		return fmt.Errorf("read line failed: %w", err)
	}
	line = bytes.TrimSpace(line)

	if string(line) != "FLAPON" {
		return fmt.Errorf("unexpected line: %s", string(line))
	}

	fmt.Printf("sending signon frame\n")

	if err := flap.SendSignonFrame(nil); err != nil {
		return fmt.Errorf("send flapon signal failed: %w", err)
	}

	signonFrame, err := flap.ReceiveSignonFrame()
	if err != nil {
		return fmt.Errorf("send flapon signal failed: %w", err)
	}

	fmt.Printf("received signon frame: %v\n", signonFrame)

	for {
		frame, err := flap.ReceiveFLAP()
		if err != nil {
			return fmt.Errorf("send flapon signal failed: %w", err)
		}

		elems, err := receiveCmd(frame.Payload)
		if err != nil {
			return fmt.Errorf("receive cmd failed: %w %s", err, frame.Payload)
		}

		if len(elems) == 0 {
			return errors.New("no cmd in flapon signal")
		}

		fmt.Printf("client: %v (%s)\n", elems, frame.Payload)

		switch elems[0] {
		case "toc_signon":

			username := elems[3]
			passwordHash, err := hex.DecodeString(elems[4][2:])
			if err != nil {
				return fmt.Errorf("decode password hash failed: %w", err)
			}
			unroasted := wire.RoastTOCPassword(passwordHash)

			host, port, err := rt.signon(username, unroasted)
			if err != nil {
				return fmt.Errorf("signon failed: %w", err)
			}

			fmt.Printf("signon: %s %s\n", host, port)

			if err := flap.SendDataFrame([]byte("SIGN_ON:1")); err != nil {
				return fmt.Errorf("send signon signal failed: %w", err)
			}

		}
	}
	return nil
}

func receiveCmd(b []byte) ([]string, error) {
	if b[len(b)-1] == '\x00' {
		b = b[:len(b)-1]
	}
	reader := csv.NewReader(bytes.NewReader(b))
	reader.Comma = ' '
	reader.LazyQuotes = false
	reader.TrimLeadingSpace = true
	return reader.Read()
}

func (rt Server) respond(s string, rwc io.ReadWriteCloser) error {
	fmt.Printf("server: %s\n", s)
	if _, err := io.WriteString(rwc, s); err != nil {
		return fmt.Errorf("error writing FLAPON: %w", err)
	}
	return nil
}

// waitForShutdown returns when either the wg completes or 5 seconds has
// passed. This is a temporary hack to ensure that the server shuts down even
// if all the TCP connections do not drain. Return true if the shutdown is
// clean.
func waitForShutdown(wg *sync.WaitGroup) bool {
	ch := make(chan struct{})

	go func() {
		wg.Wait() // goroutine leak if wg never completes
		close(ch)
	}()

	select {
	case <-ch:
		return true
	case <-time.After(time.Second * 5):
		return false
	}
}

func (rt Server) signon(screenName string, password []byte) (string, string, error) {
	host := net.JoinHostPort("127.0.0.1", "5190")
	conn, err := net.Dial("tcp", host)
	if err != nil {
		return "", "", fmt.Errorf("unable to dial into auth host: %w", err)
	}
	defer func() {
		rt.Logger.Debug("disconnected from auth service", "host", host)
		conn.Close()
	}()

	rt.Logger.Debug("connected to auth service", "host", host)

	flapc := wire.NewFlapClient(0, conn, conn)
	host, authCookie, err := authenticate(flapc, screenName, password)
	if err == nil {
		rt.Logger.Debug("authentication succeeded, proceeding to BOS host", "host", host, "authCookie", authCookie)
	}
	return host, authCookie, err
}

// authenticate performs the BUCP auth flow with the OSCAR auth server. Upon
// successful login, it returns a host name and auth cookie for connecting to
// and authenticating with the BOS service.
func authenticate(flapc *wire.FlapClient, screenName string, password []byte) (string, string, error) {
	if _, err := flapc.ReceiveSignonFrame(); err != nil {
		return "", "", fmt.Errorf("unable to receive signon frame: %w", err)
	}

	list := wire.TLVList{
		wire.NewTLVBE(wire.LoginTLVTagsScreenName, screenName),
		wire.NewTLVBE(wire.LoginTLVTagsRoastedPassword, wire.RoastPassword(password)),
	}
	if err := flapc.SendSignonFrame(list); err != nil {
		return "", "", fmt.Errorf("unable to send signon frame: %w", err)
	}

	loginFinal, err := flapc.ReceiveFLAP()
	if err != nil {
		return "", "", fmt.Errorf("unable to receive signon frame: %w", err)
	}

	loginPayload := wire.TLVList{}
	err = wire.UnmarshalBE(&loginPayload, bytes.NewBuffer(loginFinal.Payload))
	if err != nil {
		return "", "", fmt.Errorf("unable to unmarshal flap response: %w", err)
	}

	if code, hasErr := loginPayload.Uint16BE(wire.LoginTLVTagsErrorSubcode); hasErr {
		switch code {
		case wire.LoginErrInvalidUsernameOrPassword:
			return "", "", fmt.Errorf("error code from FLAP login: invalid username or password")
		default:
			return "", "", fmt.Errorf("error code from FLAP login: : %d", code)
		}
	}

	host, hasHostname := loginPayload.String(wire.LoginTLVTagsReconnectHere)
	if !hasHostname {
		return "", "", errors.New("SNAC(0x17,0x03) does not contain a hostname TLV")
	}

	authCookie, hasAuthCookie := loginPayload.String(wire.LoginTLVTagsAuthorizationCookie)
	if !hasAuthCookie {
		return "", "", errors.New("SNAC(0x17,0x03) does not contain an auth cookie TLV")
	}

	return host, authCookie, nil
}
