package server

import (
	"github.com/mkaminski/goaim/oscar"
)

func NewAlertRouter() AlertRouter {
	return AlertRouter{}
}

type AlertRouter struct {
}

func (rt *AlertRouter) RouteAlert(SNACFrame oscar.SnacFrame) error {
	switch SNACFrame.SubGroup {
	case oscar.AlertNotifyCapabilities:
		fallthrough
	case oscar.AlertNotifyDisplayCapabilities:
		// just read the request to placate the client. no need to send a
		// response.
		return nil
	default:
		return ErrUnsupportedSubGroup
	}
}
