package foodgroup

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"
)

// NewICQService creates an instance of ICQService.
func NewICQService(messageRelayer MessageRelayer) ICQService {
	return ICQService{
		messageRelayer: messageRelayer,
	}
}

// ICQService provides functionality for the ICQ (PD) food group.
// The PD food group manages settings for permit/deny (allow/block) for
// pre-feedbag (sever-side buddy list) AIM clients. Right now it's stubbed out
// to support pidgin. Eventually this food group will be fully implemented in
// order to support client blocking in AIM <= 3.0.
type ICQService struct {
	messageRelayer MessageRelayer
}

func (s ICQService) DBQuery(ctx context.Context, sess *state.Session, frame wire.SNACFrame, body wire.SNAC_0x0F_0x02_ICQDBQuery) error {
	md, ok := body.Slice(0x01)
	if !ok {
		return errors.New("invalid ICQ frame")
	}

	icqMD := wire.ICQMetadata{}
	buf := bytes.NewBuffer(md)

	if err := binary.Read(buf, binary.LittleEndian, &icqMD.ChunkSize); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.LittleEndian, &icqMD.UIN); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.LittleEndian, &icqMD.ReqType); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.LittleEndian, &icqMD.Seq); err != nil {
		return err
	}

	switch icqMD.ReqType {
	case wire.ICQReqTypeOfflineMsg:
		fmt.Println("hello")
	case wire.ICQReqTypeDeleteMsg:
		fmt.Println("hello")
	case wire.ICQReqTypeInfo:
		if err := binary.Read(buf, binary.LittleEndian, &icqMD.ReqSubType); err != nil {
			return err
		}
		switch icqMD.ReqSubType {
		case 0x04D0:

			userInfo := ReqUserInfo{}
			if err := binary.Read(buf, binary.LittleEndian, &userInfo); err != nil {
				return nil
			}

			subTypes := []uint16{
				0x00C8,
				0x00DC,
				0x00EB,
				0x010E,
				0x00D2,
				0x00E6,
				0x00F0,
				0x00FA,
			}

			seq := uint16(1)
			for _, subType := range subTypes {
				snac, err := getSNAC(0x07DA, subType, userInfo.SearchUIN, seq)
				if err != nil {
					return err
				}
				s.messageRelayer.RelayToScreenName(ctx, sess.IdentScreenName(), snac)
				seq++
			}

			fmt.Println("hello")
		}
		fmt.Println("hello")
	}

	return nil
}

type ReqUserInfo struct {
	SearchUIN uint32
}

func getSNAC(reqType uint16, reqSubType uint16, uin uint32, seq uint16) (wire.SNACMessage, error) {
	snac := wire.SNAC_0x0F_0x02_ICQDBReply{}

	md := wire.ICQMetadata{
		ChunkSize: uint16(2) + // data chunk size (TLV.Length-2)
			uint16(4) + // request owner uin
			uint16(2) + // data type: META_DATA
			uint16(2) + // 	request sequence number
			uint16(2) + // 	data subtype: META_BASIC_USERINFO
			uint16(1), //  	success byte
		UIN:        uin,
		ReqType:    reqType,
		Seq:        seq,
		ReqSubType: reqSubType,
	}

	buf := &bytes.Buffer{}
	err := wire.Marshal(md, buf)
	if err != nil {
		return wire.SNACMessage{}, err
	}
	ok := uint8(0)
	if err := binary.Write(buf, binary.LittleEndian, ok); err != nil {
		return wire.SNACMessage{}, err
	}

	snac.Append(wire.NewTLV(0x01, buf.Bytes()))
	return wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.ICQ,
			SubGroup:  wire.ICQDBReply,
		},
		Body: snac,
	}, nil
}
