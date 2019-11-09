package zstack

import (
	"bytes"
	"errors"
	"github.com/shimmeringbee/unpi"
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
)

func TestZnp(t *testing.T) {
	t.Run("async outgoing request writes bytes", func(t *testing.T) {
		writer := bytes.Buffer{}
		reader := EmptyReader{
			End: make(chan bool),
		}
		defer reader.Done()

		z := New(&reader, &writer)
		defer z.Stop()

		f := unpi.Frame{
			MessageType: unpi.AREQ,
			Subsystem:   unpi.ZDO,
			CommandID:   1,
			Payload:     []byte{0x78},
		}

		err := z.AsyncRequest(f)
		assert.NoError(t, err)

		expectedFrame := f.Marshall()
		actualFrame := writer.Bytes()

		assert.Equal(t, expectedFrame, actualFrame)
	})

	t.Run("async outgoing request with non async request errors", func(t *testing.T) {
		writer := bytes.Buffer{}
		reader := EmptyReader{
			End: make(chan bool),
		}
		defer reader.Done()

		z := New(&reader, &writer)
		defer z.Stop()

		f := unpi.Frame{
			MessageType: unpi.SREQ,
			Subsystem:   unpi.ZDO,
			CommandID:   1,
			Payload:     []byte{0x78},
		}

		err := z.AsyncRequest(f)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, FrameNotAsynchronous))
	})

	t.Run("async outgoing request passes error back to caller", func(t *testing.T) {
		expectedError := errors.New("error")

		writer := ControllableReaderWriter{
			Writer: func(p []byte) (n int, err error) {
				return 0, expectedError
			},
		}
		reader := EmptyReader{
			End: make(chan bool),
		}
		defer reader.Done()

		z := New(&reader, &writer)
		defer z.Stop()

		f := unpi.Frame{
			MessageType: unpi.AREQ,
			Subsystem:   unpi.ZDO,
			CommandID:   1,
			Payload:     []byte{0x78},
		}

		actualError := z.AsyncRequest(f)
		assert.Error(t, actualError)
		assert.Equal(t, expectedError, actualError)
	})

	t.Run("receive frames from unpi", func(t *testing.T) {
		reader := bytes.Buffer{}
		writer := bytes.Buffer{}

		expectedFrameOne := unpi.Frame{
			MessageType: 0,
			Subsystem:   unpi.ZDO,
			CommandID:   1,
			Payload:     []byte{0x78},
		}

		expectedFrameTwo := unpi.Frame{
			MessageType: 0,
			Subsystem:   unpi.SYS,
			CommandID:   2,
			Payload:     []byte{},
		}

		reader.Write(expectedFrameOne.Marshall())
		reader.Write(expectedFrameTwo.Marshall())

		z := New(&reader, &writer)
		defer z.Stop()

		frame, err := z.Receive()
		assert.NoError(t, err)
		assert.Equal(t, expectedFrameOne, frame)

		frame, err = z.Receive()
		assert.NoError(t, err)
		assert.Equal(t, expectedFrameTwo, frame)
	})

	t.Run("requesting a sync send with a non sync frame errors", func(t *testing.T) {
		reader := bytes.Buffer{}
		writer := bytes.Buffer{}

		z := New(&reader, &writer)
		defer z.Stop()

		f := unpi.Frame{
			MessageType: unpi.AREQ,
			Subsystem:   unpi.ZDO,
			CommandID:   1,
			Payload:     []byte{0x78},
		}

		_, err := z.SyncRequest(f)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, FrameNotSynchronous))
	})

	t.Run("sync requests are sent to unpi and reply is read", func(t *testing.T) {
		responseFrame := unpi.Frame{
			MessageType: unpi.SRSP,
			Subsystem:   unpi.ZDO,
			CommandID:   1,
			Payload:     []byte{},
		}
		responseBytes := responseFrame.Marshall()

		beenWrittenBuffer := bytes.Buffer{}
		r, w := io.Pipe()

		device := ControllableReaderWriter{
			Writer: func(p []byte) (n int, err error) {
				beenWrittenBuffer.Write(p)
				go func() { w.Write(responseBytes) }()
				return len(p), nil
			},
			Reader: func(p []byte) (n int, err error) {
				return r.Read(p)
			},
		}

		z := New(&device, &device)
		defer z.Stop()

		f := unpi.Frame{
			MessageType: unpi.SREQ,
			Subsystem:   unpi.ZDO,
			CommandID:   1,
			Payload:     []byte{0x78},
		}

		actualResponseFrame, err := z.SyncRequest(f)
		assert.NoError(t, err)

		expectedFrame := f.Marshall()
		actualFrame := beenWrittenBuffer.Bytes()

		assert.Equal(t, expectedFrame, actualFrame)
		assert.Equal(t, responseFrame, actualResponseFrame)
	})

	t.Run("sync outgoing request passes error during write back to caller", func(t *testing.T) {
		expectedError := errors.New("error")

		reader := bytes.Buffer{}
		writer := ControllableReaderWriter{
			Writer: func(p []byte) (n int, err error) {
				return 0, expectedError
			},
		}

		z := New(&reader, &writer)
		defer z.Stop()

		f := unpi.Frame{
			MessageType: unpi.SREQ,
			Subsystem:   unpi.ZDO,
			CommandID:   1,
			Payload:     []byte{0x78},
		}

		_, actualError := z.SyncRequest(f)
		assert.Error(t, actualError)
		assert.Equal(t, expectedError, actualError)
	})
}

type EmptyReader struct {
	End chan bool
}

func (e *EmptyReader) Done() {
	e.End <- true
}

func (e *EmptyReader) Read(p []byte) (n int, err error) {
	<-e.End

	return 0, io.EOF
}

type ControllableReaderWriter struct {
	Reader func(p []byte) (n int, err error)
	Writer func(p []byte) (n int, err error)
}

func (c *ControllableReaderWriter) Read(p []byte) (n int, err error) {
	return c.Reader(p)
}

func (c *ControllableReaderWriter) Write(p []byte) (n int, err error) {
	return c.Writer(p)
}
