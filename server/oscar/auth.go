package oscar

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"sync"
	"time"

	"github.com/mk6i/retro-aim-server/config"
	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"

	"github.com/google/uuid"
	"github.com/patrickmn/go-cache"
	"golang.org/x/time/rate"
)

type AuthService interface {
	BUCPChallenge(ctx context.Context, bodyIn wire.SNAC_0x17_0x06_BUCPChallengeRequest, newUUID func() uuid.UUID) (wire.SNACMessage, error)
	BUCPLogin(ctx context.Context, bodyIn wire.SNAC_0x17_0x02_BUCPLoginRequest, newUserFn func(screenName state.DisplayScreenName) (state.User, error)) (wire.SNACMessage, error)
	FLAPLogin(ctx context.Context, frame wire.FLAPSignonFrame, newUserFn func(screenName state.DisplayScreenName) (state.User, error)) (wire.TLVRestBlock, error)
	RegisterBOSSession(ctx context.Context, authCookie []byte) (*state.Session, error)
	RetrieveBOSSession(ctx context.Context, authCookie []byte) (*state.Session, error)
	RegisterChatSession(ctx context.Context, authCookie []byte) (*state.Session, error)
	Signout(ctx context.Context, sess *state.Session)
	SignoutChat(ctx context.Context, sess *state.Session)
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

// AuthServer is an authentication server for both FLAP (AIM v1.0-3.0) and BUCP
// (AIM v3.5-5.9) authentication flows.
type AuthServer struct {
	AuthService
	config.Config
	Logger *slog.Logger
	*IPRateLimiter
}

// Start starts the authentication server and listens for new connections.
func (rt AuthServer) Start(ctx context.Context) error {
	addr := net.JoinHostPort("", rt.Config.AuthPort)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("unable to start auth server: %w", err)
	}

	go func() {
		<-ctx.Done()
		listener.Close()
	}()

	rt.Logger.Info("starting server", "listen_host", addr, "oscar_host", rt.Config.OSCARHost)

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
			rt.Logger.DebugContext(connCtx, "accepted connection")
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

func (rt AuthServer) handleNewConnection(ctx context.Context, conn net.Conn) error {
	defer conn.Close()

	flapc := wire.NewFlapClient(100, conn, conn)

	ip, _, err := net.SplitHostPort(conn.RemoteAddr().String())
	if err != nil {
		rt.Logger.Error("failed to parse remote address", "err", err.Error())
		return err
	}

	if ok, isBUCP := rt.Allow(ip); !ok {
		rt.Logger.Error("user rate limited at login", "remote", ip)
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
		return rt.processFLAPAuth(ctx, signonFrame, flapc)
	}

	rt.SetBUCP(ip)

	return rt.processBUCPAuth(ctx, flapc)
}

func (rt AuthServer) processFLAPAuth(ctx context.Context, signonFrame wire.FLAPSignonFrame, flapc *wire.FlapClient) error {
	tlv, err := rt.AuthService.FLAPLogin(ctx, signonFrame, state.NewStubUser)
	if err != nil {
		return err
	}
	return flapc.SendSignoffFrame(tlv)
}

func (rt AuthServer) processBUCPAuth(ctx context.Context, flapc *wire.FlapClient) error {
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
			rt.Logger.Debug("signed off mid-login")
			return io.EOF // client disconnected
		case wire.FLAPFrameKeepAlive:
			rt.Logger.Debug("received flap keepalive frame")
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
				outSNAC, err := rt.BUCPChallenge(ctx, challengeRequest, uuid.New)
				if err != nil {
					return err
				}
				if err := flapc.SendSNAC(outSNAC.Frame, outSNAC.Body); err != nil {
					return err
				}

				if outSNAC.Frame.SubGroup == wire.BUCPLoginResponse {
					screenName, _ := challengeRequest.String(wire.LoginTLVTagsScreenName)
					rt.Logger.Debug("failed BUCP challenge: user does not exist", "screen_name", screenName)
					return nil // account does not exist
				}
			case fr.FoodGroup == wire.BUCP && fr.SubGroup == wire.BUCPLoginRequest:
				loginRequest := wire.SNAC_0x17_0x02_BUCPLoginRequest{}
				if err := wire.UnmarshalBE(&loginRequest, buf); err != nil {
					return err
				}
				outSNAC, err := rt.BUCPLogin(ctx, loginRequest, state.NewStubUser)
				if err != nil {
					return err
				}

				return flapc.SendSNAC(outSNAC.Frame, outSNAC.Body)
			default:
				rt.Logger.Debug("unexpected SNAC received during login",
					"foodgroup", wire.FoodGroupName(fr.FoodGroup),
					"subgroup", wire.SubGroupName(fr.FoodGroup, fr.SubGroup))
				return io.EOF
			}
		default:
			rt.Logger.Debug("unexpected frame type received during login", "type", frame.FrameType)
			return io.EOF
		}
	}
}
