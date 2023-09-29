package oscar

type SnacError struct {
	Code uint16
	TLVRestBlock
}

type FlapFrame struct {
	StartMarker   uint8
	FrameType     uint8
	Sequence      uint16
	PayloadLength uint16
}

type SnacFrame struct {
	FoodGroup uint16
	SubGroup  uint16
	Flags     uint16
	RequestID uint32
}

type FlapSignonFrame struct {
	FlapVersion uint32
	TLVRestBlock
}
