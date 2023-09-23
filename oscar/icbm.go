package oscar

import (
	"errors"
	"fmt"
	"io"
)

const (
	ICBMErr                uint16 = 0x0001
	ICBMAddParameters             = 0x0002
	ICBMDelParameters             = 0x0003
	ICBMParameterQuery            = 0x0004
	ICBMParameterReply            = 0x0005
	ICBMChannelMsgTohost          = 0x0006
	ICBMChannelMsgToclient        = 0x0007
	ICBMEvilRequest               = 0x0008
	ICBMEvilReply                 = 0x0009
	ICBMMissedCalls               = 0x000A
	ICBMClientErr                 = 0x000B
	ICBMHostAck                   = 0x000C
	ICBMSinStored                 = 0x000D
	ICBMSinListQuery              = 0x000E
	ICBMSinListReply              = 0x000F
	ICBMSinRetrieve               = 0x0010
	ICBMSinDelete                 = 0x0011
	ICBMNotifyRequest             = 0x0012
	ICBMNotifyReply               = 0x0013
	ICBMClientEvent               = 0x0014
	ICBMSinReply                  = 0x0017
)

func routeICBM(sm *SessionManager, fm *FeedbagStore, sess *Session, snac snacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	switch snac.subGroup {
	case ICBMErr:
		panic("not implemented")
	case ICBMAddParameters:
		return ReceiveAddParameters(snac, r)
	case ICBMDelParameters:
		panic("not implemented")
	case ICBMParameterQuery:
		return SendAndReceiveICBMParameterReply(snac, r, w, sequence)
	case ICBMChannelMsgTohost:
		return SendAndReceiveChannelMsgTohost(sm, fm, sess, snac, r, w, sequence)
	case ICBMChannelMsgToclient:
		panic("not implemented")
	case ICBMEvilRequest:
		return SendAndReceiveEvilRequest(sm, fm, sess, snac, r, w, sequence)
	case ICBMMissedCalls:
		panic("not implemented")
	case ICBMClientErr:
		return ReceiveClientErr(snac, r)
	case ICBMHostAck:
		panic("not implemented")
	case ICBMSinStored:
		panic("not implemented")
	case ICBMSinListQuery:
		panic("not implemented")
	case ICBMSinListReply:
		panic("not implemented")
	case ICBMSinRetrieve:
		panic("not implemented")
	case ICBMSinDelete:
		panic("not implemented")
	case ICBMNotifyRequest:
		panic("not implemented")
	case ICBMNotifyReply:
		panic("not implemented")
	case ICBMClientEvent:
		return SendAndReceiveClientEvent(sm, fm, sess, snac, r)
	case ICBMSinReply:
		panic("not implemented")
	}

	return nil
}

func SendAndReceiveICBMParameterReply(snac snacFrame, _ io.Reader, w io.Writer, sequence *uint32) error {
	fmt.Printf("sendAndReceiveICBMParameterReply read SNAC frame: %+v\n", snac)

	snacFrameOut := snacFrame{
		foodGroup: ICBM,
		subGroup:  ICBMParameterReply,
	}
	snacPayloadOut := SNAC_0x04_0x05_ICBMParameterReply{
		MaxSlots:             100,
		ICBMFlags:            3,
		MaxIncomingICBMLen:   512,
		MaxSourceEvil:        999,
		MaxDestinationEvil:   999,
		MinInterICBMInterval: 0,
	}

	return writeOutSNAC(snac, snacFrameOut, snacPayloadOut, sequence, w)
}

func SendAndReceiveChannelMsgTohost(sm *SessionManager, fm *FeedbagStore, sess *Session, snac snacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	fmt.Printf("SendAndReceiveChannelMsgTohost read SNAC frame: %+v\n", snac)

	snacPayloadIn := SNAC_0x04_0x06_ICBMChannelMsgToHost{}
	if err := Unmarshal(&snacPayloadIn, r); err != nil {
		return err
	}

	blocked, err := fm.Blocked(sess.ScreenName, snacPayloadIn.ScreenName)
	if err != nil {
		return err
	}
	if blocked != BlockedNo {
		snacFrameOut := snacFrame{
			foodGroup: ICBM,
			subGroup:  ICBMErr,
		}
		snacPayloadOut := snacError{
			Code: ErrorCodeNotLoggedOn,
		}
		if blocked == BlockedA {
			snacPayloadOut.Code = ErrorCodeInLocalPermitDeny
		}
		return writeOutSNAC(snac, snacFrameOut, snacPayloadOut, sequence, w)
	}

	recipSess, err := sm.RetrieveByScreenName(snacPayloadIn.ScreenName)
	if err != nil {
		if errors.Is(err, errSessNotFound) {
			snacFrameOut := snacFrame{
				foodGroup: ICBM,
				subGroup:  ICBMErr,
			}
			snacPayloadOut := snacError{
				Code: ErrorCodeNotLoggedOn,
			}
			return writeOutSNAC(snac, snacFrameOut, snacPayloadOut, sequence, w)
		}
		return err
	}

	clientIM := SNAC_0x04_0x07_ICBMChannelMsgToClient{
		Cookie:    snacPayloadIn.Cookie,
		ChannelID: snacPayloadIn.ChannelID,
		TLVUserInfo: TLVUserInfo{
			ScreenName:   sess.ScreenName,
			WarningLevel: sess.GetWarning(),
		},
		TLVRestBlock: TLVRestBlock{
			TLVList: TLVList{
				{
					tType: 0x0B,
					val:   []byte{},
				},
			},
		},
	}
	if messagePayload, found := snacPayloadIn.TLVRestBlock.getSlice(0x02); found {
		clientIM.TLVRestBlock.addTLV(TLV{
			tType: 0x02,
			val:   messagePayload,
		})
	}
	if messagePayload, found := snacPayloadIn.TLVRestBlock.getSlice(0x05); found {
		clientIM.TLVRestBlock.addTLV(TLV{
			tType: 0x05,
			val:   messagePayload,
		})
	}
	if t, hasAutoResp := snacPayloadIn.getTLV(0x04); hasAutoResp {
		clientIM.TLVRestBlock.addTLV(t)
	}

	sm.SendToScreenName(recipSess.ScreenName, XMessage{
		snacFrame: snacFrame{
			foodGroup: ICBM,
			subGroup:  ICBMChannelMsgToclient,
		},
		snacOut: clientIM,
	})

	if _, requestedConfirmation := snacPayloadIn.TLVRestBlock.getSlice(0x03); !requestedConfirmation {
		// don't ack message
		return nil
	}

	// ack message back to sender
	snacFrameOut := snacFrame{
		foodGroup: ICBM,
		subGroup:  ICBMHostAck,
	}
	snacPayloadOut := SNAC_0x04_0x0C_ICBMHostAck{
		Cookie:     snacPayloadIn.Cookie,
		ChannelID:  snacPayloadIn.ChannelID,
		ScreenName: snacPayloadIn.ScreenName,
	}

	return writeOutSNAC(snac, snacFrameOut, snacPayloadOut, sequence, w)
}

func ReceiveAddParameters(snac snacFrame, r io.Reader) error {
	fmt.Printf("ReceiveAddParameters read SNAC frame: %+v\n", snac)

	snacPayloadIn := SNAC_0x04_0x02_ICBMAddParameters{}
	if err := Unmarshal(&snacPayloadIn, r); err != nil {
		return err
	}

	fmt.Printf("ReceiveAddParameters read SNAC: %+v\n", snacPayloadIn)
	return nil
}

func ReceiveClientErr(snac snacFrame, r io.Reader) error {
	fmt.Printf("ReceiveClientErr read SNAC frame: %+v\n", snac)

	snacPayloadIn := SNAC_0x04_0x0B_ICBMClientErr{}
	if err := Unmarshal(&snacPayloadIn, r); err != nil {
		return err
	}

	fmt.Printf("ReceiveClientErr read SNAC: %+v\n", snacPayloadIn)
	return nil
}

func SendAndReceiveClientEvent(sm *SessionManager, fm *FeedbagStore, sess *Session, snac snacFrame, r io.Reader) error {
	fmt.Printf("SendAndReceiveClientEvent read SNAC frame: %+v\n", snac)

	snacPayloadIn := SNAC_0x04_0x14_ICBMClientEvent{}
	if err := Unmarshal(&snacPayloadIn, r); err != nil {
		return err
	}

	blocked, err := fm.Blocked(sess.ScreenName, snacPayloadIn.ScreenName)
	if err != nil {
		return err
	}
	if blocked != BlockedNo {
		return nil
	}

	sm.SendToScreenName(snacPayloadIn.ScreenName, XMessage{
		snacFrame: snacFrame{
			foodGroup: ICBM,
			subGroup:  ICBMClientEvent,
		},
		snacOut: SNAC_0x04_0x14_ICBMClientEvent{
			Cookie:     snacPayloadIn.Cookie,
			ChannelID:  snacPayloadIn.ChannelID,
			ScreenName: sess.ScreenName,
			Event:      snacPayloadIn.Event,
		},
	})

	return nil
}

const (
	evilDelta     = uint16(100)
	evilDeltaAnon = uint16(30)
)

func SendAndReceiveEvilRequest(sm *SessionManager, fm *FeedbagStore, sess *Session, snac snacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	fmt.Printf("SendAndReceiveEvilRequest read SNAC frame: %+v\n", snac)

	snacPayloadIn := SNAC_0x04_0x08_ICBMEvilRequest{}
	if err := Unmarshal(&snacPayloadIn, r); err != nil {
		return err
	}

	// don't let users warn themselves, it causes the AIM client to go into a
	// weird state.
	if snacPayloadIn.ScreenName == sess.ScreenName {
		snacFrameOut := snacFrame{
			foodGroup: ICBM,
			subGroup:  ICBMErr,
		}
		snacPayloadOut := snacError{
			Code: ErrorCodeNotSupportedByHost,
		}
		return writeOutSNAC(snac, snacFrameOut, snacPayloadOut, sequence, w)
	}

	blocked, err := fm.Blocked(sess.ScreenName, snacPayloadIn.ScreenName)
	if err != nil {
		return err
	}
	if blocked != BlockedNo {
		snacFrameOut := snacFrame{
			foodGroup: ICBM,
			subGroup:  ICBMErr,
		}
		snacPayloadOut := snacError{
			Code: ErrorCodeNotLoggedOn,
		}
		return writeOutSNAC(snac, snacFrameOut, snacPayloadOut, sequence, w)
	}

	recipSess, err := sm.RetrieveByScreenName(snacPayloadIn.ScreenName)
	if err != nil {
		return err
	}

	increase := evilDelta
	if snacPayloadIn.SendAs == 1 {
		increase = evilDeltaAnon
	}
	recipSess.IncreaseWarning(increase)

	snacFrameOut := snacFrame{
		foodGroup: ICBM,
		subGroup:  ICBMEvilReply,
	}
	snacPayloadOut := SNAC_0x04_0x09_ICBMEvilReply{
		EvilDeltaApplied: increase,
		UpdatedEvilValue: recipSess.GetWarning(),
	}

	if err := writeOutSNAC(snac, snacFrameOut, snacPayloadOut, sequence, w); err != nil {
		return err
	}

	var notif any
	if snacPayloadIn.SendAs == 0 {
		notif = SNAC_0x01_0x10_OServiceEvilNotification{
			NewEvil: recipSess.GetWarning(),
			TLVUserInfo: TLVUserInfo{
				ScreenName:   sess.ScreenName,
				WarningLevel: recipSess.GetWarning(),
			},
		}
	} else {
		notif = SNAC_0x01_0x10_OServiceEvilNotificationAnon{
			NewEvil: recipSess.GetWarning(),
		}
	}

	sm.SendToScreenName(recipSess.ScreenName, XMessage{
		snacFrame: snacFrame{
			foodGroup: OSERVICE,
			subGroup:  OServiceEvilNotification,
		},
		snacOut: notif,
	})

	return NotifyArrival(recipSess, sm, fm)
}
