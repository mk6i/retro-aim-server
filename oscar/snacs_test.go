package oscar

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"
)

func TestMarshal(t *testing.T) {
	snac := SNAC_0x0E_0x03_ChatUsersJoined{
		Users: []TLVUserInfo{
			{
				ScreenName: "screenname1",
			},
			{
				ScreenName: "screenname2",
			},
		},
	}

	buf1 := &bytes.Buffer{}
	if err := Marshal(snac, buf1); err != nil {
		t.Fatalf("error: %s", err.Error())
	}
	fmt.Println(buf1)
}

func TestUnmarshal(t *testing.T) {
	snac1 := SNAC_0x04_0x07_ICBMChannelMsgToClient{
		Cookie:    [8]byte{1, 2, 3, 4, 5, 6, 7, 8},
		ChannelID: 4,
		TLVUserInfo: TLVUserInfo{
			ScreenName:   "myscreenname",
			WarningLevel: 100,
		},
		TLVRestBlock: TLVRestBlock{
			TLVList: TLVList{
				{
					TType: 0x0B,
					Val:   []byte{1, 2, 3, 4},
				},
			},
		},
	}

	buf := &bytes.Buffer{}
	if err := Marshal(snac1, buf); err != nil {
		t.Fatalf("error: %s", err.Error())
	}

	snac2 := SNAC_0x04_0x07_ICBMChannelMsgToClient{}
	if err := Unmarshal(&snac2, buf); err != nil {
		t.Fatalf("error: %s", err.Error())
	}

	if !reflect.DeepEqual(snac1, snac2) {
		fmt.Printf("%+v\n", snac1)
		fmt.Printf("%+v\n", snac2)
		t.Fatal("structs are not the same")
	}
}

type DummyContainerLen struct {
	Elems []struct {
		Name string `len_prefix:"uint8"`
	} `len_prefix:"uint8"`
}

func TestDummyContainerLen(t *testing.T) {

	x := DummyContainerLen{
		Elems: []struct {
			Name string `len_prefix:"uint8"`
		}{
			{"Mike"},
			{"John"},
			{"Jay"},
		},
	}

	buf := &bytes.Buffer{}
	if err := Marshal(x, buf); err != nil {
		t.Fatal(err.Error())
	}

	y := DummyContainerLen{}
	if err := Unmarshal(&y, buf); err != nil {
		t.Fatal(err.Error())
	}

	if !reflect.DeepEqual(x, y) {
		t.Fatal("structs are not the same")
	}
}

type DummyContainerRest struct {
	Elems []struct {
		Name string `len_prefix:"uint8"`
	}
}

func TestDummyContainerRest(t *testing.T) {

	x := DummyContainerRest{
		Elems: []struct {
			Name string `len_prefix:"uint8"`
		}{
			{"Mike"},
			{"John"},
			{"Jay"},
		},
	}

	buf := &bytes.Buffer{}
	if err := Marshal(x, buf); err != nil {
		t.Fatal(err.Error())
	}

	y := DummyContainerRest{}
	if err := Unmarshal(&y, buf); err != nil {
		t.Fatal(err.Error())
	}

	if !reflect.DeepEqual(x, y) {
		t.Fatal("structs are not the same")
	}
}

func TestSNAC_0x01_0x17_OServiceClientVersions(t *testing.T) {

	x := SNAC_0x01_0x17_OServiceClientVersions{
		Versions: []uint16{1, 2, 3},
	}

	buf := &bytes.Buffer{}
	if err := Marshal(x, buf); err != nil {
		t.Fatal(err.Error())
	}

	y := SNAC_0x01_0x17_OServiceClientVersions{}
	if err := Unmarshal(&y, buf); err != nil {
		t.Fatal(err.Error())
	}

	if !reflect.DeepEqual(x, y) {
		t.Fatal("structs are not the same")
	}
}
