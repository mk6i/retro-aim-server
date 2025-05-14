package handler

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/mk6i/retro-aim-server/wire"
)

func TestStatsHandler_ReportEvents(t *testing.T) {
	input := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.Stats,
			SubGroup:  wire.StatsReportEvents,
		},
		Body: wire.SNAC_0x0B_0x03_StatsReportEvents{},
	}
	output := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.Stats,
			SubGroup:  wire.StatsReportAck,
		},
		Body: wire.SNAC_0x0B_0x04_StatsReportAck{},
	}

	svc := newMockStatsService(t)
	svc.EXPECT().
		ReportEvents(mock.Anything, input.Frame, input.Body).
		Return(output)

	h := NewStatsHandler(slog.Default(), svc)

	ss := newMockResponseWriter(t)
	ss.EXPECT().
		SendSNAC(output.Frame, output.Body).
		Return(nil)

	buf := &bytes.Buffer{}
	assert.NoError(t, wire.MarshalBE(input.Body, buf))

	assert.NoError(t, h.ReportEvents(nil, nil, input.Frame, buf, ss))
}
