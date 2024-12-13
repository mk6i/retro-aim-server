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
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
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

func (rt Server) handleNewConnection(ctx context.Context, clientConn io.ReadWriteCloser) error {

	go func() {
		<-ctx.Done()
		clientConn.Close()
	}()
	reader := bufio.NewReader(clientConn)

	clientFlap := wire.NewFlapClient(0, clientConn, clientConn)

	line, _, err := reader.ReadLine()
	if err != nil {
		return fmt.Errorf("read line failed: %w", err)
	}
	line = bytes.TrimSpace(line)

	if string(line) != "FLAPON" {
		return fmt.Errorf("unexpected line: %s", string(line))
	}

	fmt.Printf("sending signon frame\n")

	if err := clientFlap.SendSignonFrame(nil); err != nil {
		return fmt.Errorf("send flapon signal failed: %w", err)
	}

	signonFrame, err := clientFlap.ReceiveSignonFrame()
	if err != nil {
		return fmt.Errorf("send flapon signal failed: %w", err)
	}

	fmt.Printf("received signon frame: %v\n", signonFrame)

	var serverFlap *wire.FlapClient
	var serverConn net.Conn

	defer func() {
		if serverConn != nil {
			serverConn.Close()
		}
	}()
	for {
		clientFrame, err := clientFlap.ReceiveFLAP()
		if err != nil {
			return fmt.Errorf("send flapon signal failed: %w", err)
		}

		if clientFrame.FrameType == wire.FLAPFrameSignoff {
			break // client disconnected
		}
		if clientFrame.FrameType == wire.FLAPFrameKeepAlive {
			continue // keep alive heartbeat
		}
		if clientFrame.FrameType != wire.FLAPFrameData {
			return fmt.Errorf("unexpected clientFlap clientFrame type: %s", clientFrame.FrameType)
		}

		elems, err := receiveCmd(clientFrame.Payload)
		if err != nil {
			return fmt.Errorf("receive cmd failed: %w %s", err, clientFrame.Payload)
		}

		if len(elems) == 0 {
			return errors.New("no cmd in flapon signal")
		}

		fmt.Printf("client: %+v\n", elems)

		switch elems[0] {
		case "toc_signon":
			username := elems[3]
			passwordHash, err := hex.DecodeString(elems[4][2:])
			if err != nil {
				return fmt.Errorf("decode password hash failed: %w", err)
			}
			unroasted := wire.RoastTOCPassword(passwordHash)

			host, cookie, err := rt.signon(username, unroasted)
			if err != nil {
				return fmt.Errorf("signon failed: %w", err)
			}

			fmt.Printf("signon: %s %s\n", host, cookie)

			serverConn, err = net.Dial("tcp", host)
			if err != nil {
				return fmt.Errorf("dial failed: %w", err)
			}

			rt.Logger.Info("connected to BOS server", "host", host)

			serverFlap = wire.NewFlapClient(0, serverConn, serverConn)

			if _, err := serverFlap.ReceiveSignonFrame(); err != nil {
				return err
			}

			tlv := []wire.TLV{
				wire.NewTLVBE(wire.OServiceTLVTagsLoginCookie, []byte(cookie)),
			}
			if err := serverFlap.SendSignonFrame(tlv); err != nil {
				return err
			}

			hostOnlineFrame := wire.SNACFrame{}
			hostOnlineSNAC := wire.SNAC_0x01_0x03_OServiceHostOnline{}
			if err := serverFlap.ReceiveSNAC(&hostOnlineFrame, &hostOnlineSNAC); err != nil {
				return err
			}
			if err := clientFlap.SendDataFrame([]byte("SIGN_ON:1")); err != nil {
				return fmt.Errorf("send signon signal failed: %w", err)
			}

			go func() {
				if err := rt.receiveFromServer(serverFlap, clientFlap); err != nil {
					fmt.Sprintf("%w\n", err)
					//if err != io.EOF {
					//	panic("receiveFromServer err: " + err.Error())
					return
				}
			}()
		case "toc_send_im":
			recip := elems[1]
			msg := elems[2]
			snac, err := sendMessageSNAC(0, recip, msg)
			if err != nil {
				return fmt.Errorf("getting message snac failed failed: %w", err)
			}
			err = serverFlap.SendSNAC(snac.Frame, snac.Body)
			if err != nil {
				return fmt.Errorf("send snac failed failed: %w", err)
			}
		case "toc_init_done":
			frame := wire.SNACFrame{
				FoodGroup: wire.OService,
				SubGroup:  wire.OServiceClientOnline,
			}
			snac := wire.SNAC_0x01_0x02_OServiceClientOnline{}
			if err := serverFlap.SendSNAC(frame, snac); err != nil {
				return err
			}
		case "toc_add_buddy":
			frame := wire.SNACFrame{
				FoodGroup: wire.Buddy,
				SubGroup:  wire.BuddyAddBuddies,
			}
			snac := wire.SNAC_0x03_0x04_BuddyAddBuddies{}
			elems = elems[1:]
			for _, sn := range elems {
				snac.Buddies = append(snac.Buddies, struct {
					ScreenName string `oscar:"len_prefix=uint8"`
				}{ScreenName: sn})
			}
			if err := serverFlap.SendSNAC(frame, snac); err != nil {
				return err
			}
		case "toc_remove_buddy":
			frame := wire.SNACFrame{
				FoodGroup: wire.Buddy,
				SubGroup:  wire.BuddyDelBuddies,
			}
			snac := wire.SNAC_0x03_0x05_BuddyDelBuddies{}
			elems = elems[1:]
			for _, sn := range elems {
				snac.Buddies = append(snac.Buddies, struct {
					ScreenName string `oscar:"len_prefix=uint8"`
				}{ScreenName: sn})
			}
			if err := serverFlap.SendSNAC(frame, snac); err != nil {
				return err
			}
		case "toc_add_permit":
			frame := wire.SNACFrame{
				FoodGroup: wire.PermitDeny,
				SubGroup:  wire.PermitDenyAddPermListEntries,
			}
			snac := wire.SNAC_0x09_0x05_PermitDenyAddPermListEntries{}
			elems = elems[1:]
			for _, sn := range elems {
				snac.Users = append(snac.Users, struct {
					ScreenName string `oscar:"len_prefix=uint8"`
				}{ScreenName: sn})
			}
			if err := serverFlap.SendSNAC(frame, snac); err != nil {
				return err
			}
		case "toc_add_deny":
			frame := wire.SNACFrame{
				FoodGroup: wire.PermitDeny,
				SubGroup:  wire.PermitDenyAddDenyListEntries,
			}
			snac := wire.SNAC_0x09_0x07_PermitDenyAddDenyListEntries{}
			elems = elems[1:]
			for _, sn := range elems {
				snac.Users = append(snac.Users, struct {
					ScreenName string `oscar:"len_prefix=uint8"`
				}{ScreenName: sn})
			}
			if err := serverFlap.SendSNAC(frame, snac); err != nil {
				return err
			}
		case "toc_set_away":
			frame := wire.SNACFrame{
				FoodGroup: wire.Locate,
				SubGroup:  wire.LocateSetInfo,
			}
			snac := wire.SNAC_0x02_0x04_LocateSetInfo{
				TLVRestBlock: wire.TLVRestBlock{
					TLVList: wire.TLVList{
						wire.NewTLVBE(wire.LocateTLVTagsInfoUnavailableData, elems[1]),
					},
				},
			}
			if err := serverFlap.SendSNAC(frame, snac); err != nil {
				return err
			}
		case "toc_set_caps":
			elems = elems[1:]
			caps := make([]uuid.UUID, 0, len(elems))
			for _, capStr := range elems {
				uid, err := uuid.Parse(capStr)
				if err != nil {
					return fmt.Errorf("parse caps failed: %w", err)
				}
				caps = append(caps, uid)
			}

			frame := wire.SNACFrame{
				FoodGroup: wire.Locate,
				SubGroup:  wire.LocateSetInfo,
			}
			snac := wire.SNAC_0x02_0x04_LocateSetInfo{
				TLVRestBlock: wire.TLVRestBlock{
					TLVList: wire.TLVList{
						wire.NewTLVBE(wire.LocateTLVTagsInfoCapabilities, caps),
					},
				},
			}
			if err := serverFlap.SendSNAC(frame, snac); err != nil {
				return err
			}
		case "toc_evil":
			frame := wire.SNACFrame{
				FoodGroup: wire.ICBM,
				SubGroup:  wire.ICBMEvilRequest,
			}
			snac := wire.SNAC_0x04_0x08_ICBMEvilRequest{
				SendAs:     0,
				ScreenName: elems[1],
			}
			if elems[2] == "anon" {
				snac.SendAs = 1
			}
			if err := serverFlap.SendSNAC(frame, snac); err != nil {
				return err
			}
		case "toc_get_info":
			//frame := wire.SNACFrame{
			//	FoodGroup: wire.Locate,
			//	SubGroup:  wire.LocateUserInfoQuery,
			//}
			//snac := wire.SNAC_0x02_0x05_LocateUserInfoQuery{
			//	Type: uint16(wire.LocateTypeSig),
			//	ScreenName: elems[1],
			//}
			//if err := serverFlap.SendSNAC(frame, snac); err != nil {
			//	return err
			//}
			if err := clientFlap.SendDataFrame([]byte("GOTO_URL:hello:http://frogfind.com:80")); err != nil {
				return fmt.Errorf("send info signal failed: %w", err)
			}
		}
	}
	return nil
}

func (rt Server) receiveFromServer(serverFlap, clientFlap *wire.FlapClient) error {
	for {
		flap, err := serverFlap.ReceiveFLAP()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("receive flap failed: %w", err)
		}

		switch flap.FrameType {
		case wire.FLAPFrameData:
			flapBuf := bytes.NewBuffer(flap.Payload)

			inFrame := wire.SNACFrame{}
			if err := wire.UnmarshalBE(&inFrame, flapBuf); err != nil {
				return err
			}

			switch inFrame.FoodGroup {
			case wire.Buddy:
				switch inFrame.SubGroup {
				case wire.BuddyArrived:
					sn := wire.TOCBuddyArrived{}
					if err := wire.UnmarshalBE(&sn, flapBuf); err != nil {
						return fmt.Errorf("unmarshal buddy arrived: %w", err)
					}
					if err := clientFlap.SendDataFrame([]byte(sn.String())); err != nil {
						return fmt.Errorf("sending im to client failed: %w", err)
					}
				case wire.BuddyDeparted:
					sn := wire.TOCBuddyDeparted{}
					if err := wire.UnmarshalBE(&sn, flapBuf); err != nil {
						return fmt.Errorf("unmarshal buddy arrived: %w", err)
					}
					if err := clientFlap.SendDataFrame([]byte(sn.String())); err != nil {
						return fmt.Errorf("sending im to client failed: %w", err)
					}
				}
			case wire.ICBM:
				switch inFrame.SubGroup {
				case wire.ICBMChannelMsgToClient:
					sn := wire.TOCIMIN{}
					if err := wire.UnmarshalBE(&sn, flapBuf); err != nil {
						return fmt.Errorf("unmarshal ICBM channel message failed: %w", err)
					}
					if err := clientFlap.SendDataFrame([]byte(sn.String())); err != nil {
						return fmt.Errorf("sending im to client failed: %w", err)
					}
				default:
					rt.Logger.Info("unsupported snac. foodgroup: %s subgroup: %s", wire.FoodGroupName(inFrame.FoodGroup), wire.SubGroupName(inFrame.FoodGroup, inFrame.SubGroup))
				}
			default:
				rt.Logger.Info("unsupported snac. foodgroup: %s subgroup: %s", wire.FoodGroupName(inFrame.FoodGroup), wire.SubGroupName(inFrame.FoodGroup, inFrame.SubGroup))
			}
		default:

		}
	}
	return nil
}
func sendMessageSNAC(cookie uint64, screenName string, response string) (wire.SNACMessage, error) {
	msgFrame := wire.SNACFrame{
		FoodGroup: wire.ICBM,
		SubGroup:  wire.ICBMChannelMsgToHost,
	}

	// build the response message
	response = strings.ReplaceAll("@MsgContent@", "@MsgContent@", response)

	frags, err := wire.ICBMFragmentList(response)
	if err != nil {
		return wire.SNACMessage{}, fmt.Errorf("unable to create ICBM fragment list: %w", err)
	}

	return wire.SNACMessage{
		Frame: msgFrame,
		Body: wire.SNAC_0x04_0x06_ICBMChannelMsgToHost{
			Cookie:     cookie,
			ChannelID:  1,
			ScreenName: screenName,
			TLVRestBlock: wire.TLVRestBlock{
				TLVList: wire.TLVList{
					wire.NewTLVBE(wire.ICBMTLVAOLIMData, frags),
				},
			},
		},
	}, nil
}

func receiveCmd(b []byte) ([]string, error) {
	if b[len(b)-1] == '\x00' {
		b = b[:len(b)-1]
	}
	if bytes.HasPrefix(b, []byte("toc_set_config")) {
		// gaim uses braces instead of quotes for some reason
		first := bytes.IndexByte(b, '{')
		if first != -1 {
			b[first] = '"'
		}
		last := bytes.LastIndexByte(b, '}')
		if last != -1 {
			b[last] = '"'
		}
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
