package oscar

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFLAPFrame_ReadBody(t *testing.T) {
	flap := FLAPFrame{
		PayloadLength: 4,
	}
	bufIn := bytes.NewBuffer([]byte{0, 1, 2, 3, 4, 5})
	buf, err := flap.ReadBody(bufIn)
	assert.NoError(t, err)
	assert.Equal(t, []byte{0, 1, 2, 3}, buf.Bytes())
}

func TestFLAPFrame_ReadBodyError(t *testing.T) {
	flap := FLAPFrame{
		PayloadLength: 4,
	}
	bufIn := &bytes.Buffer{}
	_, err := flap.ReadBody(bufIn)
	assert.ErrorIs(t, err, io.EOF)
}
