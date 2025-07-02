package oscar

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/netip"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/patrickmn/go-cache"
	"golang.org/x/sync/errgroup"
	"golang.org/x/time/rate"

	"github.com/mk6i/retro-aim-server/server/oscar/middleware"
	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"
)

// OnlineNotifier returns a OServiceHostOnline SNAC that is sent to the client
// at the beginning of the protocol sequence which lists all food groups
// managed by the server.
type OnlineNotifier interface {
	HostOnline() wire.SNACMessage
}

// BuddyListRegistry is the interface for keeping track of users with active
// buddy lists. Once registered, a user becomes visible to other users' buddy
// lists and vice versa.
type BuddyListRegistry interface {
	ClearBuddyListRegistry(ctx context.Context) error
	RegisterBuddyList(ctx context.Context, user state.IdentScreenName) error
	UnregisterBuddyList(ctx context.Context, user state.IdentScreenName) error
}

// DepartureNotifier is the interface for sending buddy departure notifications
// when a client disconnects.
type DepartureNotifier interface {
	BroadcastBuddyDeparted(ctx context.Context, sess *state.Session) error
}

// ChatSessionManager is the interface for closing chat sessions
// when a client disconnects.
type ChatSessionManager interface {
	RemoveUserFromAllChats(user state.IdentScreenName)
}

// RateLimitUpdater provides rate limit updates for subscribed rate limit classes.
type RateLimitUpdater interface {
	RateLimitUpdates(ctx context.Context, sess *state.Session, now time.Time) []wire.SNACMessage
}

type AuthService interface {
	BUCPChallenge(ctx context.Context, bodyIn wire.SNAC_0x17_0x06_BUCPChallengeRequest, newUUID func() uuid.UUID) (wire.SNACMessage, error)
	BUCPLogin(ctx context.Context, bodyIn wire.SNAC_0x17_0x02_BUCPLoginRequest, newUserFn func(screenName state.DisplayScreenName) (state.User, error)) (wire.SNACMessage, error)
	FLAPLogin(ctx context.Context, frame wire.FLAPSignonFrame, newUserFn func(screenName state.DisplayScreenName) (state.User, error)) (wire.TLVRestBlock, error)
	KerberosLogin(ctx context.Context, inBody wire.SNAC_0x050C_0x0002_KerberosLoginRequest, newUserFn func(screenName state.DisplayScreenName) (state.User, error)) (wire.SNACMessage, error)
	RegisterBOSSession(ctx context.Context, authCookie []byte) (*state.Session, error)
	RegisterChatSession(ctx context.Context, authCookie []byte) (*state.Session, error)
	RetrieveBOSSession(ctx context.Context, authCookie []byte) (*state.Session, error)
	Signout(ctx context.Context, sess *state.Session)
	SignoutChat(ctx context.Context, sess *state.Session)
	GetSession(ctx context.Context, authCookie []byte) (uint16, *state.Session, error)
}

// IPRateLimiter enforces a per-IP rate limit using a token bucket algorithm.
// It caches individual rate limiters by IP address and supports tagging requests
// as originating from the BUCP or FLAP auth.
//
// The limiter uses an in-memory cache with TTL expiration, so rate limits reset
// after the TTL if no activity is observed for a given IP.
type IPRateLimiter struct {
	cache *cache.Cache // In-memory cache mapping IPs to rate limiters with optional BUCP tag
	rate  rate.Limit   // Requests allowed per second
	burst int          // Maximum burst size allowed
}

type rateLimitEntry struct {
	isBUCP  bool
	limiter *rate.Limiter
}

// NewIPRateLimiter initializes a new IPRateLimiter with the specified rate,
// burst size, and TTL for each IP's limiter. Entries expire after 2×TTL.
func NewIPRateLimiter(rate rate.Limit, burst int, ttl time.Duration) *IPRateLimiter {
	return &IPRateLimiter{
		cache: cache.New(ttl, 2*ttl),
		rate:  rate,
		burst: burst,
	}
}

// SetBUCP marks the rate limiter for the given IP as originating from BUCP auth
// (default FLAP auth).
func (l *IPRateLimiter) SetBUCP(ip string) {
	limiter, found := l.cache.Get(ip)
	if !found {
		limiter = &rateLimitEntry{
			isBUCP:  true,
			limiter: rate.NewLimiter(l.rate, l.burst),
		}
		l.cache.Set(ip, limiter, cache.DefaultExpiration)
	}
	limiter.(*rateLimitEntry).isBUCP = true
}

// Allow checks if a request from the given IP is allowed under its rate limit.
// It returns whether the request is allowed and whether the connection uses
// BUCP auth.
func (l *IPRateLimiter) Allow(ip string) (allowed bool, isBUCP bool) {
	limiter, found := l.cache.Get(ip)
	if !found {
		limiter = &rateLimitEntry{
			limiter: rate.NewLimiter(l.rate, l.burst),
		}
		l.cache.Set(ip, limiter, cache.DefaultExpiration)
	}
	entry := limiter.(*rateLimitEntry)
	return entry.limiter.Allow(), entry.isBUCP
}

type Listener struct {
	Hostname      string
	Port          string
	AdvertiseHost string
	AdvertisePort string
}

type Server struct {
	Listeners []Listener
	AuthService
	BuddyListRegistry
	ChatSessionManager *state.InMemoryChatSessionManager
	DepartureNotifier
	Logger *slog.Logger
	OnlineNotifier
	Handler
	RateLimitUpdater
	wire.SNACRateLimits
	*IPRateLimiter
}

func (s Server) Start(ctx context.Context) error {
	if err := s.BuddyListRegistry.ClearBuddyListRegistry(ctx); err != nil {
		return fmt.Errorf("unable to clear client-side buddy list: %s", err.Error())
	}

	errGroup, ctx := errgroup.WithContext(ctx)

	for _, l := range s.Listeners {
		errGroup.Go(func() error {
			return s.acceptConnection(l, ctx)
		})
	}

	return errGroup.Wait()
}

func (s Server) acceptConnection(l Listener, ctx context.Context) error {
	listener, err := net.Listen("tcp", net.JoinHostPort(l.Hostname, l.Port))
	if err != nil {
		return fmt.Errorf("unable to start BOS server: %w", err)
	}

	go func() {
		<-ctx.Done()
		listener.Close()
	}()

	s.Logger.Info("starting server", "listen_host", net.JoinHostPort(l.Hostname, l.Port),
		"advertise_host", net.JoinHostPort(l.AdvertiseHost, l.AdvertisePort))

	wg := sync.WaitGroup{}
	for {
		conn, err := listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				break
			}
			s.Logger.Error("accept failed", "err", err.Error())
			continue
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			connCtx := context.WithValue(ctx, "ip", conn.RemoteAddr().String())
			s.Logger.DebugContext(connCtx, "accepted connection")
			if err := s.routeConnection(connCtx, conn); err != nil {
				s.Logger.Info("user session failed", "err", err.Error())
			}
		}()
	}

	if !waitForShutdown(&wg) {
		s.Logger.Error("shutdown complete, but connections didn't close cleanly")
	} else {
		s.Logger.Info("shutdown complete")
	}

	return nil
}

func (s Server) routeConnection(ctx context.Context, rwc net.Conn) error {
	defer func() {
		rwc.Close()
	}()

	ip, _, err := net.SplitHostPort(rwc.RemoteAddr().String())
	if err != nil {
		s.Logger.Error("failed to parse remote address", "err", err.Error())
		return err
	}

	flapc := wire.NewFlapClient(100, rwc, rwc)

	if err := flapc.SendSignonFrame(nil); err != nil {
		return err
	}
	flap, err := flapc.ReceiveSignonFrame()
	if err != nil {
		return err
	}

	if flap.HasTag(wire.OServiceTLVTagsLoginCookie) {
		return s.connectToService(ctx, flap, flapc, ip, rwc)
	}

	return s.doAuthStuff(ctx, rwc, ip, flapc)
}

func (s Server) connectToService(ctx context.Context, flap wire.FLAPSignonFrame, flapc *wire.FlapClient, ip string, rwc net.Conn) error {
	authCookie, ok := flap.Bytes(wire.OServiceTLVTagsLoginCookie)
	if !ok {
		return errors.New("unable to get session id from payload")
	}

	fg, sess, err := s.GetSession(ctx, authCookie)
	if err != nil {
		return err
	}
	if sess == nil {
		return errors.New("session not found")
	}

	ctx = context.WithValue(ctx, "screenName", sess.IdentScreenName())

	switch fg {
	case 2: // BOS
		if err := s.BuddyListRegistry.RegisterBuddyList(ctx, sess.IdentScreenName()); err != nil {
			return fmt.Errorf("unable to init buddy list: %w", err)
		}

		defer func() {
			sess.Close()
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			if err := s.DepartureNotifier.BroadcastBuddyDeparted(ctx, sess); err != nil {
				s.Logger.ErrorContext(ctx, "error sending buddy departure notifications", "err", err.Error())
			}
			// buddy list must be cleared before session is closed, otherwise
			// there will be a race condition that could cause the buddy list
			// be prematurely deleted.
			if err := s.BuddyListRegistry.UnregisterBuddyList(ctx, sess.IdentScreenName()); err != nil {
				s.Logger.ErrorContext(ctx, "error removing buddy list entry", "err", err.Error())
			}
			s.ChatSessionManager.RemoveUserFromAllChats(sess.IdentScreenName())
			s.Signout(ctx, sess)
		}()
		remoteAddr, ok := ctx.Value("ip").(string)
		if ok {
			ip, err := netip.ParseAddrPort(remoteAddr)
			if err != nil {
				return errors.New("unable to parse ip addr")
			}
			sess.SetRemoteAddr(&ip)
		}
	case 3: // Chat
		defer func() {
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			s.SignoutChat(ctx, sess)
		}()

	}

	msg := s.OnlineNotifier.HostOnline()
	if err := flapc.SendSNAC(msg.Frame, msg.Body); err != nil {
		return err
	}

	return dispatchIncomingMessages(ctx, sess, flapc, rwc, s.Logger, s.Handler, s.RateLimitUpdater, s.SNACRateLimits)
}

func (s Server) doAuthStuff(ctx context.Context, conn net.Conn, ip string, flapc *wire.FlapClient) error {
	if ok, isBUCP := s.Allow(ip); !ok {
		s.Logger.Error("user rate limited at login", "remote", ip)
		tlv := wire.TLVRestBlock{
			TLVList: []wire.TLV{
				wire.NewTLVBE(wire.LoginTLVTagsErrorSubcode, wire.LoginErrRateLimitExceeded),
			},
		}
		// gives wrong response if you quickly switch between BUCP/FLAP clients
		if isBUCP {
			return flapc.SendSNAC(
				wire.SNACFrame{
					FoodGroup: wire.BUCP,
					SubGroup:  wire.BUCPLoginResponse,
				},
				wire.SNAC_0x17_0x03_BUCPLoginResponse{
					TLVRestBlock: tlv,
				},
			)
		} else {
			return flapc.SendSignoffFrame(tlv)
		}
	}

	// auth must complete within the next 30 seconds
	if err := conn.SetDeadline(time.Now().Add(30 * time.Second)); err != nil {
		return fmt.Errorf("failed to set deadline: %w", err)
	}

	if err := flapc.SendSignonFrame(nil); err != nil {
		return err
	}

	signonFrame, err := flapc.ReceiveSignonFrame()
	if err != nil {
		return err
	}

	// decide whether the client is using BUCP or FLAP authentication based on
	// the presence of the screen name TLV. this block used to check for the
	// presence of the roasted password TLV, however that proved an unreliable
	// indicator of FLAP-auth because older ICQ clients appear to omit the
	// roasted password TLV when the password is not stored client-side.
	if _, hasScreenName := signonFrame.Uint16BE(wire.LoginTLVTagsScreenName); hasScreenName {
		return s.processFLAPAuth(ctx, signonFrame, flapc)
	}

	s.SetBUCP(ip)

	return s.processBUCPAuth(ctx, flapc)
}

func (s Server) processFLAPAuth(ctx context.Context, signonFrame wire.FLAPSignonFrame, flapc *wire.FlapClient) error {
	tlv, err := s.AuthService.FLAPLogin(ctx, signonFrame, state.NewStubUser)
	if err != nil {
		return err
	}
	return flapc.SendSignoffFrame(tlv)
}

func (s Server) processBUCPAuth(ctx context.Context, flapc *wire.FlapClient) error {
	frames := 0

	for {
		frame, err := flapc.ReceiveFLAP()
		if err != nil {
			return err
		}

		if frames > 10 {
			// a lot of frames received, the client is misbehaving
			return fmt.Errorf("too many auth flap packets received")
		}
		frames++

		switch frame.FrameType {
		case wire.FLAPFrameSignoff:
			s.Logger.Debug("signed off mid-login")
			return io.EOF // client disconnected
		case wire.FLAPFrameKeepAlive:
			s.Logger.Debug("received flap keepalive frame")
		case wire.FLAPFrameData:
			buf := bytes.NewReader(frame.Payload)
			fr := wire.SNACFrame{}
			if err := wire.UnmarshalBE(&fr, buf); err != nil {
				return err
			}
			switch {
			case fr.FoodGroup == wire.BUCP && fr.SubGroup == wire.BUCPChallengeRequest:
				challengeRequest := wire.SNAC_0x17_0x06_BUCPChallengeRequest{}
				if err := wire.UnmarshalBE(&challengeRequest, buf); err != nil {
					return err
				}
				outSNAC, err := s.BUCPChallenge(ctx, challengeRequest, uuid.New)
				if err != nil {
					return err
				}
				if err := flapc.SendSNAC(outSNAC.Frame, outSNAC.Body); err != nil {
					return err
				}

				if outSNAC.Frame.SubGroup == wire.BUCPLoginResponse {
					screenName, _ := challengeRequest.String(wire.LoginTLVTagsScreenName)
					s.Logger.Debug("failed BUCP challenge: user does not exist", "screen_name", screenName)
					return nil // account does not exist
				}
			case fr.FoodGroup == wire.BUCP && fr.SubGroup == wire.BUCPLoginRequest:
				loginRequest := wire.SNAC_0x17_0x02_BUCPLoginRequest{}
				if err := wire.UnmarshalBE(&loginRequest, buf); err != nil {
					return err
				}
				outSNAC, err := s.BUCPLogin(ctx, loginRequest, state.NewStubUser)
				if err != nil {
					return err
				}

				return flapc.SendSNAC(outSNAC.Frame, outSNAC.Body)
			default:
				s.Logger.Debug("unexpected SNAC received during login",
					"foodgroup", wire.FoodGroupName(fr.FoodGroup),
					"subgroup", wire.SubGroupName(fr.FoodGroup, fr.SubGroup))
				return io.EOF
			}
		default:
			s.Logger.Debug("unexpected frame type received during login", "type", frame.FrameType)
			return io.EOF
		}
	}
}

func sendInvalidSNACErr(frameIn wire.SNACFrame, rw ResponseWriter) error {
	frameOut := wire.SNACFrame{
		FoodGroup: frameIn.FoodGroup,
		SubGroup:  0x01, // error subgroup for all SNACs
		RequestID: frameIn.RequestID,
	}
	bodyOut := wire.SNACError{
		Code: wire.ErrorCodeInvalidSnac,
	}
	return rw.SendSNAC(frameOut, bodyOut)
}

// dispatchIncomingMessages receives incoming messages and sends them to the
// appropriate message handler. Messages from the client are sent to the
// router. Messages relayed from the user session are forwarded to the client.
// This function ensures that the same sequence number is incremented for both
// types of messages. The function terminates upon receiving a connection error
// or when the session closes.
//
// todo: this method has too many params and should be folded into a new type
func dispatchIncomingMessages(ctx context.Context, sess *state.Session, flapc *wire.FlapClient, r io.Reader, logger *slog.Logger, router Handler, rateLimitUpdater RateLimitUpdater, snacRateLimits wire.SNACRateLimits) error {

	defer func() {
		logger.InfoContext(ctx, "user disconnected")
	}()

	// buffered so that the go routine has room to exit
	msgCh := make(chan wire.FLAPFrame, 1)
	errCh := make(chan error, 1)

	// consume flap frames
	go func() {
		defer close(msgCh)
		defer close(errCh)

		for {
			frame := wire.FLAPFrame{}
			if err := wire.UnmarshalBE(&frame, r); err != nil {
				errCh <- err
				return
			}
			msgCh <- frame
		}
	}()

	for {
		select {
		case flap, ok := <-msgCh:
			if !ok {
				return nil
			}
			switch flap.FrameType {
			case wire.FLAPFrameData:
				flapBuf := bytes.NewBuffer(flap.Payload)

				inFrame := wire.SNACFrame{}
				if err := wire.UnmarshalBE(&inFrame, flapBuf); err != nil {
					return err
				}

				rateClassID, ok := snacRateLimits.RateClassLookup(inFrame.FoodGroup, inFrame.SubGroup)
				if ok {
					if status := sess.EvaluateRateLimit(time.Now(), rateClassID); status == wire.RateLimitStatusLimited {
						logger.DebugContext(ctx, "rate limit exceeded, dropping SNAC",
							"foodgroup", wire.FoodGroupName(inFrame.FoodGroup),
							"subgroup", wire.SubGroupName(inFrame.FoodGroup, inFrame.SubGroup))
						break
					}
				} else {
					logger.ErrorContext(ctx, "rate limit not found, allowing request through")
				}

				// route a client request to the appropriate service handler. the
				// handler may write a response to the client connection.
				if err := router.Handle(ctx, sess, inFrame, flapBuf, flapc); err != nil {
					middleware.LogRequestError(ctx, logger, inFrame, err)
					if errors.Is(err, ErrRouteNotFound) {
						if err1 := sendInvalidSNACErr(inFrame, flapc); err1 != nil {
							return errors.Join(err1, err)
						}
						break
					}
					return err
				}
			case wire.FLAPFrameSignon:
				return fmt.Errorf("shouldn't get FLAPFrameSignon. flap: %v", flap)
			case wire.FLAPFrameError:
				return fmt.Errorf("got FLAPFrameError. flap: %v", flap)
			case wire.FLAPFrameSignoff:
				logger.InfoContext(ctx, "got FLAPFrameSignoff", "flap", flap)
				return nil
			case wire.FLAPFrameKeepAlive:
				logger.DebugContext(ctx, "keepalive heartbeat")
			default:
				return fmt.Errorf("got unknown FLAP frame type. flap: %v", flap)
			}
		case m := <-sess.ReceiveMessage():
			// forward a notification sent from another client to this client
			if err := flapc.SendSNAC(m.Frame, m.Body); err != nil {
				middleware.LogRequestError(ctx, logger, m.Frame, err)
				return err
			}
			middleware.LogRequest(ctx, logger, m.Frame, m.Body)
		case <-time.After(1 * time.Second):
			msgs := rateLimitUpdater.RateLimitUpdates(ctx, sess, time.Now())
			for _, rate := range msgs {
				if err := flapc.SendSNAC(rate.Frame, rate.Body); err != nil {
					middleware.LogRequestError(ctx, logger, rate.Frame, err)
					return err
				}
			}
		case <-sess.Closed():
			block := wire.TLVRestBlock{}
			// error code indicating user signed in a different location
			block.Append(wire.NewTLVBE(0x0009, wire.OServiceDiscErrNewLogin))
			// "more info" button
			block.Append(wire.NewTLVBE(0x000b, "https://github.com/mk6i/retro-aim-server"))
			if err := flapc.SendSignoffFrame(block); err != nil {
				return fmt.Errorf("unable to gracefully disconnect user. %w", err)
			}
			return nil
		case <-ctx.Done():
			// application is shutting down
			if err := flapc.Disconnect(); err != nil {
				return fmt.Errorf("unable to gracefully disconnect user. %w", err)
			}
			return nil
		case err := <-errCh:
			if !errors.Is(io.EOF, err) {
				logger.ErrorContext(ctx, "client disconnected with error", "err", err)
			}
			return nil
		}
	}
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
