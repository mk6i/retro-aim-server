package oscar

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"
)

// ErrRouteNotFound is an error that indicates a failure to find a matching
// route for an OSCAR protocol request.
var ErrRouteNotFound = errors.New("route not found")

// ResponseWriter is the interface for sending a SNAC response to the client
// from the server handlers.
type ResponseWriter interface {
	SendSNAC(frame wire.SNACFrame, body any) error
}

// Handler defines an interface for routing and processing OSCAR protocol
// requests based on their food group categorization. Implementers of this
// interface should provide logic to handle incoming requests, perform
// necessary operations, and possibly generate responses.
type Handler interface {
	Handle(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, r io.Reader, rw ResponseWriter) error
}

// NewRouter creates a new Router instance.
func NewRouter() Router {
	return Router{
		entries: make(map[uint16]map[uint16]Handler),
	}
}

// Router defines a structure for routing OSCAR protocol requests to
// appropriate handlers based on group:subGroup identifiers.
type Router struct {
	entries map[uint16]map[uint16]Handler
}

// HandlerFunc defines a function type that implements the Handler interface.
// This allows using simple functions as handlers for processing OSCAR requests.
type HandlerFunc func(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, r io.Reader, rw ResponseWriter) error

// Handle executes the HandlerFunc, facilitating the Handler interface implementation.
func (f HandlerFunc) Handle(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, r io.Reader, rw ResponseWriter) error {
	return f(ctx, sess, inFrame, r, rw)
}

// Register adds a new route to the router by associating a HandlerFunc with a
// specific group:subGroup pair. If a handler is already registered for the
// group:subGroup pair, it will be overwritten with the new handler.
func (rt Router) Register(group uint16, subGroup uint16, fn HandlerFunc) {
	if _, ok := rt.entries[group]; !ok {
		rt.entries[group] = make(map[uint16]Handler)
	}
	rt.entries[group][subGroup] = fn
}

// Handle directs an incoming OSCAR request to the appropriate handler based on
// its group and subGroup identifiers found in the SNAC frame. It returns an
// ErrRouteNotFound error if no matching handler is found for the group:subGroup
// pair in the request.
func (rt Router) Handle(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, r io.Reader, rw ResponseWriter) error {
	h, ok := rt.entries[inFrame.FoodGroup][inFrame.SubGroup]
	if !ok {
		return fmt.Errorf("%w. group: %d subgroup: %d", ErrRouteNotFound, inFrame.FoodGroup, inFrame.SubGroup)
	}
	return h.Handle(ctx, sess, inFrame, r, rw)
}
