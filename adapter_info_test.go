package zstack

import (
	"context"
	. "github.com/shimmeringbee/unpi"
	unpiTest "github.com/shimmeringbee/unpi/testing"
	"github.com/shimmeringbee/zigbee"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func Test_GetAdapterIEEEAddress(t *testing.T) {
	t.Run("gets the IEEE address", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		unpiMock := unpiTest.NewMockAdapter()
		zstack := New(unpiMock)
		defer unpiMock.Stop()

		c := unpiMock.On(SREQ, SAPI, SAPIZBGetDeviceInfoID).Return(Frame{
			MessageType: SRSP,
			Subsystem:   SAPI,
			CommandID:   SAPIZBGetDeviceInfoReplyID,
			Payload:     []byte{0x01, 0x0f, 0x0e, 0x0d, 0x0c, 0x0b, 0x0a, 0x09, 0x08},
		})

		address, err := zstack.GetAdapterIEEEAddress(ctx)
		assert.NoError(t, err)
		assert.Equal(t, zigbee.IEEEAddress(0x08090a0b0c0d0e0f), address)

		assert.Equal(t, uint8(0x01), c.CapturedCalls[0].Frame.Payload[0])

		unpiMock.AssertCalls(t)
	})
}

func Test_GetAdapterNetworkAddress(t *testing.T) {
	t.Run("gets the network address", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		unpiMock := unpiTest.NewMockAdapter()
		zstack := New(unpiMock)
		defer unpiMock.Stop()

		c := unpiMock.On(SREQ, SAPI, SAPIZBGetDeviceInfoID).Return(Frame{
			MessageType: SRSP,
			Subsystem:   SAPI,
			CommandID:   SAPIZBGetDeviceInfoReplyID,
			Payload:     []byte{0x02, 0x09, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		})

		address, err := zstack.GetAdapterNetworkAddress(ctx)
		assert.NoError(t, err)
		assert.Equal(t, zigbee.NetworkAddress(0x0809), address)

		assert.Equal(t, uint8(0x02), c.CapturedCalls[0].Frame.Payload[0])

		unpiMock.AssertCalls(t)
	})
}
