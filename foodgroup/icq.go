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

	icqChunk := wire.ICQChunk{}
	if err := wire.UnmarshalICQ(&icqChunk, bytes.NewBuffer(md)); err != nil {
		return err
	}
	buf := bytes.NewBuffer(icqChunk.Body)
	icqMD := wire.ICQMetadata{}
	if err := wire.UnmarshalICQ(&icqMD, buf); err != nil {
		return err
	}

	switch icqMD.ReqType {
	case wire.ICQReqTypeOfflineMsg:
		fmt.Println("hello")
	case wire.ICQReqTypeDeleteMsg:
		fmt.Println("hello")
	case wire.ICQReqTypeInfo:
		switch icqMD.ReqSubType {
		case 0x04D0:

			userInfo := ReqUserInfo{}
			if err := binary.Read(buf, binary.LittleEndian, &userInfo); err != nil {
				return nil
			}

			// send SNAC(15,03)/07DA/00C8
			seq := uint16(2)
			snac, err := GetICQUserInfo(userInfo.SearchUIN, seq)
			if err != nil {
				return fmt.Errorf("get ICQ user info: %w", err)
			}
			sess.RelayMessage(snac)

			// send SNAC(15,03)/07DA/00DC
			seq++
			snac, err = GetICQMoreUserInfo(userInfo.SearchUIN, seq)
			if err != nil {
				return fmt.Errorf("get ICQ user more: %w", err)
			}
			sess.RelayMessage(snac)

			// send SNAC(15,03)/07DA/00EB
			seq++
			snac, err = GetICQInfoEmailMore(userInfo.SearchUIN, seq)
			if err != nil {
				return fmt.Errorf("get ICQ email more: %w", err)
			}
			sess.RelayMessage(snac)

			// send SNAC(15,03)/07DA/010E
			seq++
			snac, err = GetICQHomepageCat(userInfo.SearchUIN, seq)
			if err != nil {
				return fmt.Errorf("get ICQ homepage cat: %w", err)
			}
			sess.RelayMessage(snac)

			// send SNAC(15,03)/07DA/00D2
			seq++
			snac, err = GetICQMetaWorkUserInfo(userInfo.SearchUIN, seq)
			if err != nil {
				return fmt.Errorf("get ICQ work user info: %w", err)
			}
			sess.RelayMessage(snac)

			// send SNAC(15,03)/07DA/00E6
			seq++
			snac, err = GetICQUserNotes(userInfo.SearchUIN, seq)
			if err != nil {
				return fmt.Errorf("get ICQ user notes: %w", err)
			}
			sess.RelayMessage(snac)

			// send SNAC(15,03)/07DA/00F0
			seq++
			snac, err = GetICQUserInterests(userInfo.SearchUIN, seq)
			if err != nil {
				return fmt.Errorf("get ICQ user interests: %w", err)
			}
			sess.RelayMessage(snac)

			// send SNAC(15,03)/07DA/00FA
			seq++
			snac, err = GetICQMetaAffiliationsUserInfo(userInfo.SearchUIN, seq)
			if err != nil {
				return fmt.Errorf("get ICQ affiliations user: %w", err)
			}
			sess.RelayMessage(snac)

			// SNAC 1

			//subTypes := []uint16{
			//	0x00FA,
			//	0x00DC,
			//	0x00EB,
			//	0x010E,
			//	0x00D2,
			//	0x00E6,
			//	0x00F0,
			//	0x00C8,
			//}

			//seq := uint16(1)
			//for i, subType := range subTypes {
			//	seq++
			//	snac, err := getSNAC(0x07DA, subType, userInfo.SearchUIN, seq)
			//	if err != nil {
			//		return err
			//	}
			//	//seq++
			//	snac.Frame.Flags = setFirstBit(1)
			//	if i == len(subTypes)-1 {
			//		snac.Frame.Flags = 0
			//	}
			//	sess.RelayMessage(snac)
			//	//if i == 2 {
			//	//	break
			//	//}
			//	//s.messageRelayer.RelayToScreenName(ctx, sess.IdentScreenName(), snac)
			//}

			fmt.Println("hello")
		}
		fmt.Println("hello")
	}

	return nil
}

func setFirstBit(n uint16) uint16 {
	return n | 1
}

type ReqUserInfo struct {
	SearchUIN uint32
}

func GetICQUserInfo(uin uint32, seq uint16) (wire.SNACMessage, error) {
	buf := &bytes.Buffer{}
	userInfo := wire.ICQUserInfo{
		Success:      0x0A,
		Nickname:     "mike",
		FirstName:    "mike",
		LastName:     "mike",
		Email:        "mike@mike.com",
		HomeCity:     "New York",
		HomeState:    "New York",
		HomePhone:    "555-555-5555",
		HomeFax:      "555-555-5555",
		HomeAddress:  "555 Street",
		CellPhone:    "555-555-5555",
		ZipCode:      "11111",
		CountryCode:  1,
		GMTOffset:    2,
		AuthFlag:     0,
		WebAware:     1,
		DCPerms:      0,
		PublishEmail: 1,
	}
	if err := wire.MarshalICQ(userInfo, buf); err != nil {
		return wire.SNACMessage{}, err
	}

	md := wire.ICQMetadata{
		UIN:        uin,
		ReqType:    0x07DA,
		ReqSubType: 0x00C8,
		Seq:        seq,
	}
	if err := wire.MarshalICQ(md, buf); err != nil {
		return wire.SNACMessage{}, err
	}

	chunk := wire.ICQChunk{
		Body: buf.Bytes(),
	}
	buf2 := &bytes.Buffer{}
	if err := wire.MarshalICQ(chunk, buf2); err != nil {
		return wire.SNACMessage{}, err
	}

	return wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.ICQ,
			SubGroup:  wire.ICQDBReply,
			Flags:     1,
		},
		Body: wire.SNAC_0x0F_0x02_ICQDBReply{
			TLVRestBlock: wire.TLVRestBlock{
				TLVList: wire.TLVList{
					wire.NewTLV(0x01, buf2.Bytes()),
				},
			},
		},
	}, nil
}

func GetICQMoreUserInfo(uin uint32, seq uint16) (wire.SNACMessage, error) {
	buf := &bytes.Buffer{}
	userInfo := wire.ICQMoreUserInfo{
		Success: 0x0A,
	}
	if err := wire.MarshalICQ(userInfo, buf); err != nil {
		return wire.SNACMessage{}, err
	}

	md := wire.ICQMetadata{
		UIN:        uin,
		ReqType:    0x07DA,
		ReqSubType: 0x00DC,
		Seq:        seq,
	}
	if err := wire.MarshalICQ(md, buf); err != nil {
		return wire.SNACMessage{}, err
	}

	chunk := wire.ICQChunk{
		Body: buf.Bytes(),
	}
	buf2 := &bytes.Buffer{}
	if err := wire.MarshalICQ(chunk, buf2); err != nil {
		return wire.SNACMessage{}, err
	}

	return wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.ICQ,
			SubGroup:  wire.ICQDBReply,
			Flags:     1,
		},
		Body: wire.SNAC_0x0F_0x02_ICQDBReply{
			TLVRestBlock: wire.TLVRestBlock{
				TLVList: wire.TLVList{
					wire.NewTLV(0x01, buf2.Bytes()),
				},
			},
		},
	}, nil
}

func GetICQInfoEmailMore(uin uint32, seq uint16) (wire.SNACMessage, error) {
	buf := &bytes.Buffer{}
	userInfo := wire.ICQInfoEmailMore{
		Success: 0x0A,
	}
	if err := wire.MarshalICQ(userInfo, buf); err != nil {
		return wire.SNACMessage{}, err
	}

	md := wire.ICQMetadata{
		UIN:        uin,
		ReqType:    0x07DA,
		ReqSubType: 0x00EB,
		Seq:        seq,
	}
	if err := wire.MarshalICQ(md, buf); err != nil {
		return wire.SNACMessage{}, err
	}

	chunk := wire.ICQChunk{
		Body: buf.Bytes(),
	}
	buf2 := &bytes.Buffer{}
	if err := wire.MarshalICQ(chunk, buf2); err != nil {
		return wire.SNACMessage{}, err
	}

	return wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.ICQ,
			SubGroup:  wire.ICQDBReply,
			Flags:     1,
		},
		Body: wire.SNAC_0x0F_0x02_ICQDBReply{
			TLVRestBlock: wire.TLVRestBlock{
				TLVList: wire.TLVList{
					wire.NewTLV(0x01, buf2.Bytes()),
				},
			},
		},
	}, nil
}

func GetICQHomepageCat(uin uint32, seq uint16) (wire.SNACMessage, error) {
	buf := &bytes.Buffer{}
	userInfo := wire.ICQHomepageCat{
		Success: 0x0A,
	}
	if err := wire.MarshalICQ(userInfo, buf); err != nil {
		return wire.SNACMessage{}, err
	}

	md := wire.ICQMetadata{
		UIN:        uin,
		ReqType:    0x07DA,
		ReqSubType: 0x010E,
		Seq:        seq,
	}
	if err := wire.MarshalICQ(md, buf); err != nil {
		return wire.SNACMessage{}, err
	}

	chunk := wire.ICQChunk{
		Body: buf.Bytes(),
	}
	buf2 := &bytes.Buffer{}
	if err := wire.MarshalICQ(chunk, buf2); err != nil {
		return wire.SNACMessage{}, err
	}

	return wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.ICQ,
			SubGroup:  wire.ICQDBReply,
			Flags:     1,
		},
		Body: wire.SNAC_0x0F_0x02_ICQDBReply{
			TLVRestBlock: wire.TLVRestBlock{
				TLVList: wire.TLVList{
					wire.NewTLV(0x01, buf2.Bytes()),
				},
			},
		},
	}, nil
}

func GetICQMetaWorkUserInfo(uin uint32, seq uint16) (wire.SNACMessage, error) {
	buf := &bytes.Buffer{}
	userInfo := wire.ICQMetaWorkUserInfo{
		Success: 0x0A,
	}
	if err := wire.MarshalICQ(userInfo, buf); err != nil {
		return wire.SNACMessage{}, err
	}

	md := wire.ICQMetadata{
		UIN:        uin,
		ReqType:    0x07DA,
		ReqSubType: 0x00D2,
		Seq:        seq,
	}
	if err := wire.MarshalICQ(md, buf); err != nil {
		return wire.SNACMessage{}, err
	}

	chunk := wire.ICQChunk{
		Body: buf.Bytes(),
	}
	buf2 := &bytes.Buffer{}
	if err := wire.MarshalICQ(chunk, buf2); err != nil {
		return wire.SNACMessage{}, err
	}

	return wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.ICQ,
			SubGroup:  wire.ICQDBReply,
			Flags:     1,
		},
		Body: wire.SNAC_0x0F_0x02_ICQDBReply{
			TLVRestBlock: wire.TLVRestBlock{
				TLVList: wire.TLVList{
					wire.NewTLV(0x01, buf2.Bytes()),
				},
			},
		},
	}, nil
}

func GetICQUserNotes(uin uint32, seq uint16) (wire.SNACMessage, error) {
	buf := &bytes.Buffer{}
	userInfo := wire.ICQUserNotes{
		Success: 0x0A,
	}
	if err := wire.MarshalICQ(userInfo, buf); err != nil {
		return wire.SNACMessage{}, err
	}

	md := wire.ICQMetadata{
		UIN:        uin,
		ReqType:    0x07DA,
		ReqSubType: 0x00E6,
		Seq:        seq,
	}
	if err := wire.MarshalICQ(md, buf); err != nil {
		return wire.SNACMessage{}, err
	}

	chunk := wire.ICQChunk{
		Body: buf.Bytes(),
	}
	buf2 := &bytes.Buffer{}
	if err := wire.MarshalICQ(chunk, buf2); err != nil {
		return wire.SNACMessage{}, err
	}

	return wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.ICQ,
			SubGroup:  wire.ICQDBReply,
			Flags:     1,
		},
		Body: wire.SNAC_0x0F_0x02_ICQDBReply{
			TLVRestBlock: wire.TLVRestBlock{
				TLVList: wire.TLVList{
					wire.NewTLV(0x01, buf2.Bytes()),
				},
			},
		},
	}, nil
}

func GetICQUserInterests(uin uint32, seq uint16) (wire.SNACMessage, error) {
	buf := &bytes.Buffer{}
	userInfo := wire.ICQUserInterests{
		Success: 0x0A,
		Interests: make([]struct {
			Code    uint16
			Keyword string `len_prefix:"uint16"`
		}, 4),
	}
	if err := wire.MarshalICQ(userInfo, buf); err != nil {
		return wire.SNACMessage{}, err
	}

	md := wire.ICQMetadata{
		UIN:        uin,
		ReqType:    0x07DA,
		ReqSubType: 0x00F0,
		Seq:        seq,
	}
	if err := wire.MarshalICQ(md, buf); err != nil {
		return wire.SNACMessage{}, err
	}

	chunk := wire.ICQChunk{
		Body: buf.Bytes(),
	}
	buf2 := &bytes.Buffer{}
	if err := wire.MarshalICQ(chunk, buf2); err != nil {
		return wire.SNACMessage{}, err
	}

	return wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.ICQ,
			SubGroup:  wire.ICQDBReply,
			Flags:     1,
		},
		Body: wire.SNAC_0x0F_0x02_ICQDBReply{
			TLVRestBlock: wire.TLVRestBlock{
				TLVList: wire.TLVList{
					wire.NewTLV(0x01, buf2.Bytes()),
				},
			},
		},
	}, nil
}

func GetICQMetaAffiliationsUserInfo(uin uint32, seq uint16) (wire.SNACMessage, error) {
	buf := &bytes.Buffer{}
	userInfo := wire.ICQMetaAffiliationsUserInfo{
		Success: 0x0A,
		PastAffiliations: make([]struct {
			Code    uint16
			Keyword string `len_prefix:"uint16"`
		}, 3),
		Affiliations: make([]struct {
			Code    uint16
			Keyword string `len_prefix:"uint16"`
		}, 3),
	}
	if err := wire.MarshalICQ(userInfo, buf); err != nil {
		return wire.SNACMessage{}, err
	}

	md := wire.ICQMetadata{
		UIN:        uin,
		ReqType:    0x07DA,
		ReqSubType: 0x00FA,
		Seq:        seq,
	}
	if err := wire.MarshalICQ(md, buf); err != nil {
		return wire.SNACMessage{}, err
	}

	chunk := wire.ICQChunk{
		Body: buf.Bytes(),
	}
	buf2 := &bytes.Buffer{}
	if err := wire.MarshalICQ(chunk, buf2); err != nil {
		return wire.SNACMessage{}, err
	}

	return wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.ICQ,
			SubGroup:  wire.ICQDBReply,
		},
		Body: wire.SNAC_0x0F_0x02_ICQDBReply{
			TLVRestBlock: wire.TLVRestBlock{
				TLVList: wire.TLVList{
					wire.NewTLV(0x01, buf2.Bytes()),
				},
			},
		},
	}, nil
}

//func getSNAC(reqType uint16, reqSubType uint16, uin uint32, seq uint16) (wire.SNACMessage, error) {
//	snac := wire.SNAC_0x0F_0x02_ICQDBReply{}
//
//	md := wire.ICQMetadata{
//		ChunkSize: uint16(4) + // request owner uin
//			uint16(2) + // data type: META_DATA
//			uint16(2) + // 	request sequence number
//			uint16(2) + // 	data subtype: META_BASIC_USERINFO
//			uint16(1), //  	success byte
//		UIN:        uin,
//		ReqType:    reqType,
//		Seq:        seq,
//		ReqSubType: reqSubType,
//	}
//
//	buf := &bytes.Buffer{}
//	ok := uint8(0x0)
//	if err := binary.Write(buf, binary.LittleEndian, md.ChunkSize); err != nil {
//		return wire.SNACMessage{}, err
//	}
//	if err := binary.Write(buf, binary.LittleEndian, md.UIN); err != nil {
//		return wire.SNACMessage{}, err
//	}
//	if err := binary.Write(buf, binary.LittleEndian, md.ReqType); err != nil {
//		return wire.SNACMessage{}, err
//	}
//	if err := binary.Write(buf, binary.LittleEndian, md.Seq); err != nil {
//		return wire.SNACMessage{}, err
//	}
//	if err := binary.Write(buf, binary.LittleEndian, md.ReqSubType); err != nil {
//		return wire.SNACMessage{}, err
//	}
//	if err := binary.Write(buf, binary.LittleEndian, ok); err != nil {
//		return wire.SNACMessage{}, err
//	}
//
//	snac.Append(wire.NewTLV(0x01, buf.Bytes()))
//	return wire.SNACMessage{
//		Frame: wire.SNACFrame{
//			FoodGroup: wire.ICQ,
//			SubGroup:  wire.ICQDBReply,
//		},
//		Body: snac,
//	}, nil
//}
