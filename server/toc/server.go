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
	"strconv"
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
		go func() {
			<-ctx.Done()
			conn.Close()
		}()

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

func (rt Server) handleNewConnection(ctx context.Context, clientConn io.ReadWriter) error {
	if err := rt.TOCHandshake(clientConn); err != nil {
		return fmt.Errorf("handshake failed: %w", err)
	}

	clientFlap, err := rt.initFLAP(clientConn)
	if err != nil {
		return err
	}

	clientCh := make(chan any)
	bosCh := make(chan wire.SNACMessage)
	chatNavCh := make(chan wire.SNACMessage)

	defer func() {
		close(clientCh)
		close(bosCh)
	}()

	go func() {
		if err := rt.sendToClient(ctx, clientCh, clientFlap, chatNavCh); err != nil {
			rt.Logger.Error("failed to receive from server", "err", err.Error())
			return
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
			err = rt.initBOS(ctx, elems, clientCh, bosCh)
			if err != nil {
				return err
			}
			if err := clientFlap.SendDataFrame([]byte("SIGN_ON:1")); err != nil {
				return fmt.Errorf("send sign on data frame failed: %w", err)
			}
		case "toc_send_im":
			recip := elems[1]
			msg := elems[2]
			snac, err := sendMessageSNAC(0, recip, msg)
			if err != nil {
				return fmt.Errorf("getting message snac failed failed: %w", err)
			}
			bosCh <- snac
		case "toc_init_done":
			bosCh <- wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.OService,
					SubGroup:  wire.OServiceClientOnline,
				},
				Body: wire.SNAC_0x01_0x02_OServiceClientOnline{},
			}
		case "toc_add_buddy":
			snac := wire.SNAC_0x03_0x04_BuddyAddBuddies{}
			elems = elems[1:]
			for _, sn := range elems {
				snac.Buddies = append(snac.Buddies, struct {
					ScreenName string `oscar:"len_prefix=uint8"`
				}{ScreenName: sn})
			}
			bosCh <- wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.Buddy,
					SubGroup:  wire.BuddyAddBuddies,
				},
				Body: snac,
			}
		case "toc_remove_buddy":
			snac := wire.SNAC_0x03_0x05_BuddyDelBuddies{}
			elems = elems[1:]
			for _, sn := range elems {
				snac.Buddies = append(snac.Buddies, struct {
					ScreenName string `oscar:"len_prefix=uint8"`
				}{ScreenName: sn})
			}
			bosCh <- wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.Buddy,
					SubGroup:  wire.BuddyDelBuddies,
				},
				Body: snac,
			}
		case "toc_add_permit":
			snac := wire.SNAC_0x09_0x05_PermitDenyAddPermListEntries{}
			elems = elems[1:]
			for _, sn := range elems {
				snac.Users = append(snac.Users, struct {
					ScreenName string `oscar:"len_prefix=uint8"`
				}{ScreenName: sn})
			}
			bosCh <- wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.PermitDeny,
					SubGroup:  wire.PermitDenyAddPermListEntries,
				},
				Body: snac,
			}
		case "toc_add_deny":
			snac := wire.SNAC_0x09_0x07_PermitDenyAddDenyListEntries{}
			elems = elems[1:]
			for _, sn := range elems {
				snac.Users = append(snac.Users, struct {
					ScreenName string `oscar:"len_prefix=uint8"`
				}{ScreenName: sn})
			}
			bosCh <- wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.PermitDeny,
					SubGroup:  wire.PermitDenyAddDenyListEntries,
				},
				Body: snac,
			}
		case "toc_set_away":
			bosCh <- wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.Locate,
					SubGroup:  wire.LocateSetInfo,
				},
				Body: wire.SNAC_0x02_0x04_LocateSetInfo{
					TLVRestBlock: wire.TLVRestBlock{
						TLVList: wire.TLVList{
							wire.NewTLVBE(wire.LocateTLVTagsInfoUnavailableData, elems[1]),
						},
					},
				},
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

			bosCh <- wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.Locate,
					SubGroup:  wire.LocateSetInfo,
				},
				Body: wire.SNAC_0x02_0x04_LocateSetInfo{
					TLVRestBlock: wire.TLVRestBlock{
						TLVList: wire.TLVList{
							wire.NewTLVBE(wire.LocateTLVTagsInfoCapabilities, caps),
						},
					},
				},
			}
		case "toc_evil":
			snac := wire.SNAC_0x04_0x08_ICBMEvilRequest{
				SendAs:     0,
				ScreenName: elems[1],
			}
			if elems[2] == "anon" {
				snac.SendAs = 1
			}
			bosCh <- wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.ICBM,
					SubGroup:  wire.ICBMEvilRequest,
				},
				Body: snac,
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
			clientCh <- []byte("GOTO_URL:hello:http://frogfind.com:80")
		case "toc_chat_join":
			exchange, err := strconv.Atoi(elems[1])
			if err != nil {
				return fmt.Errorf("parse exchange failed: %w", err)
			}
			bosCh <- wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.OService,
					SubGroup:  wire.OServiceServiceRequest,
				},
				Body: wire.SNAC_0x01_0x04_OServiceServiceRequest{
					FoodGroup: wire.ChatNav,
				},
			}
			chatNavCh <- wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.ChatNav,
					SubGroup:  wire.ChatNavCreateRoom,
				},
				Body: wire.SNAC_0x0E_0x02_ChatRoomInfoUpdate{
					Exchange: uint16(exchange),
					Cookie:   "create",
					TLVBlock: wire.TLVBlock{
						TLVList: wire.TLVList{
							wire.NewTLVBE(wire.ChatRoomTLVRoomName, elems[2]),
						},
					},
				},
			}
		}
	}
	return nil
}

func (rt Server) TOCHandshake(clientConn io.ReadWriter) error {
	reader := bufio.NewReader(clientConn)

	line, _, err := reader.ReadLine()
	if err != nil {
		return fmt.Errorf("read line failed: %w", err)
	}
	line = bytes.TrimSpace(line)

	if string(line) != "FLAPON" {
		return fmt.Errorf("unexpected line: %s", string(line))
	}
	return nil
}

func (rt Server) initFLAP(clientConn io.ReadWriter) (*wire.FlapClient, error) {
	clientFlap := wire.NewFlapClient(0, clientConn, clientConn)

	fmt.Printf("sending signon frame\n")
	if err := clientFlap.SendSignonFrame(nil); err != nil {
		return nil, fmt.Errorf("send flapon signal failed: %w", err)
	}

	signonFrame, err := clientFlap.ReceiveSignonFrame()
	if err != nil {
		return nil, fmt.Errorf("send flapon signal failed: %w", err)
	}

	fmt.Printf("received signon frame: %v\n", signonFrame)
	return clientFlap, nil
}

func (rt Server) initBOS(ctx context.Context, elems []string, clientCh chan<- any, ch <-chan wire.SNACMessage) error {
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

	serverConn, err := net.Dial("tcp", host)
	if err != nil {
		return fmt.Errorf("dial failed: %w", err)
	}

	go func() {
		defer serverConn.Close()
		<-ctx.Done()
	}()

	rt.Logger.Info("connected to BOS server", "host", host)

	serverFlap := wire.NewFlapClient(0, serverConn, serverConn)

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

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-ch:
				if err := serverFlap.SendSNAC(msg.Frame, msg.Body); err != nil {
					rt.Logger.Error("send snac failed", "err", err)
				}
			}
		}
	}()
	go func() {
		for {
			flap, err := serverFlap.ReceiveFLAP()
			if err != nil {
				if err == io.EOF {
					return
				}
				rt.Logger.Error("receive signon frame failed", "err", err)
			}
			clientCh <- flap
		}
	}()
	return nil
}

func (rt Server) initChatNav(ctx context.Context, host string, cookie []byte, clientCh chan<- any, navCh <-chan wire.SNACMessage) error {
	serverConn, err := net.Dial("tcp", host)
	if err != nil {
		return fmt.Errorf("dial failed: %w", err)
	}

	go func() {
		defer serverConn.Close()
		<-ctx.Done()
	}()

	rt.Logger.Info("connected to BOS server", "host", host)

	serverFlap := wire.NewFlapClient(0, serverConn, serverConn)

	if _, err := serverFlap.ReceiveSignonFrame(); err != nil {
		return err
	}

	tlv := []wire.TLV{
		wire.NewTLVBE(wire.OServiceTLVTagsLoginCookie, cookie),
	}
	if err := serverFlap.SendSignonFrame(tlv); err != nil {
		return err
	}

	hostOnlineFrame := wire.SNACFrame{}
	hostOnlineSNAC := wire.SNAC_0x01_0x03_OServiceHostOnline{}
	if err := serverFlap.ReceiveSNAC(&hostOnlineFrame, &hostOnlineSNAC); err != nil {
		return err
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-navCh:
				if err := serverFlap.SendSNAC(msg.Frame, msg.Body); err != nil {
					rt.Logger.Error("send snac failed", "err", err)
				}
			}
		}
	}()
	go func() {
		for {
			flap, err := serverFlap.ReceiveFLAP()
			if err != nil {
				if err == io.EOF {
					return
				}
				rt.Logger.Error("receive signon frame failed", "err", err)
			}
			clientCh <- flap
		}
	}()
	return nil
}

func (rt Server) sendToClient(ctx context.Context, clientCh chan any, clientFlap *wire.FlapClient, navCh chan wire.SNACMessage) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msgIn := <-clientCh:
			switch msgIn := msgIn.(type) {
			case wire.FLAPFrame:
				switch msgIn.FrameType {
				case wire.FLAPFrameData:
					flapBuf := bytes.NewBuffer(msgIn.Payload)

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
					case wire.OService:
						switch inFrame.SubGroup {
						case wire.OServiceServiceResponse:
							sn := wire.SNAC_0x01_0x05_OServiceServiceResponse{}
							if err := wire.UnmarshalBE(&sn, flapBuf); err != nil {
								return fmt.Errorf("unmarshal ICBM channel message failed: %w", err)
							}
							host, _ := sn.String(wire.OServiceTLVTagsReconnectHere)
							cookie, _ := sn.Bytes(wire.OServiceTLVTagsLoginCookie)
							group, _ := sn.Uint16BE(wire.OServiceTLVTagsGroupID)

							switch group {
							case wire.ChatNav:
								if err := rt.initChatNav(ctx, host, cookie, clientCh, navCh); err != nil {
									return fmt.Errorf("initChatNav failed: %w", err)
								}
							default:
								return fmt.Errorf("unsupported oservice response. group: %d", group)
							}

						default:
							rt.Logger.Info("unsupported snac. foodgroup: %s subgroup: %s", wire.FoodGroupName(inFrame.FoodGroup), wire.SubGroupName(inFrame.FoodGroup, inFrame.SubGroup))
						}
					case wire.ChatNav:
						switch inFrame.SubGroup {
						case wire.ChatNavNavInfo:
							sn := wire.SNAC_0x0D_0x09_ChatNavNavInfo{}
							if err := wire.UnmarshalBE(&sn, flapBuf); err != nil {
								return fmt.Errorf("unmarshal ICBM channel message failed: %w", err)
							}
							host, _ := sn.String(wire.OServiceTLVTagsReconnectHere)
							cookie, _ := sn.Bytes(wire.OServiceTLVTagsLoginCookie)
							group, _ := sn.Uint16BE(wire.OServiceTLVTagsGroupID)

						default:
							rt.Logger.Info("unsupported snac. foodgroup: %s subgroup: %s", wire.FoodGroupName(inFrame.FoodGroup), wire.SubGroupName(inFrame.FoodGroup, inFrame.SubGroup))
						}
					default:
						rt.Logger.Info("unsupported snac. foodgroup: %s subgroup: %s", wire.FoodGroupName(inFrame.FoodGroup), wire.SubGroupName(inFrame.FoodGroup, inFrame.SubGroup))
					}
				default:
				}
			}
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
