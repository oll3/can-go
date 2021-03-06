package candebug

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.einride.tech/can"
	"go.einride.tech/can/pkg/descriptor"
	"go.einride.tech/can/pkg/generated"
)

func TestServeMessagesHTTP_Single(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ServeMessagesHTTP(w, r, []generated.Message{
			&testMessage{
				frame:      can.Frame{ID: 100, Length: 1},
				descriptor: newDriverHeartbeatDescriptor(),
			},
		})
	}))
	res, err := http.Get(ts.URL)
	require.NoError(t, err)
	response, err := ioutil.ReadAll(res.Body)
	require.NoError(t, err)
	require.NoError(t, res.Body.Close())
	const expected = `
DriverHeartbeat
===============
ID: 100 (0x64)
Sender: DRIVER
SendType: Cyclic
CycleTime: 100ms
DelayTime: 2s
===============
Command: 0 (0x0) None
`
	require.Equal(t, strings.TrimSpace(expected), string(response))
}

func TestServeMessagesHTTP_Multi(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ServeMessagesHTTP(w, r, []generated.Message{
			&testMessage{
				frame:      can.Frame{ID: 100, Length: 1},
				descriptor: newDriverHeartbeatDescriptor(),
			},
			&testMessage{
				frame:      can.Frame{ID: 100, Length: 1, Data: can.Data{0x01}},
				descriptor: newDriverHeartbeatDescriptor(),
			},
		})
	}))
	res, err := http.Get(ts.URL)
	require.NoError(t, err)
	response, err := ioutil.ReadAll(res.Body)
	require.NoError(t, err)
	require.NoError(t, res.Body.Close())
	const expected = `
DriverHeartbeat
===============
ID: 100 (0x64)
Sender: DRIVER
SendType: Cyclic
CycleTime: 100ms
DelayTime: 2s
===============
Command: 0 (0x0) None


DriverHeartbeat
===============
ID: 100 (0x64)
Sender: DRIVER
SendType: Cyclic
CycleTime: 100ms
DelayTime: 2s
===============
Command: 1 (0x1) Sync
`
	require.Equal(t, strings.TrimSpace(expected), string(response))
}

type testMessage struct {
	frame      can.Frame
	descriptor *descriptor.Message
}

func (m *testMessage) Frame() can.Frame {
	return m.frame
}

func (m *testMessage) Descriptor() *descriptor.Message {
	return m.descriptor
}

func (m *testMessage) MarshalFrame() (can.Frame, error) {
	panic("should not be called")
}

func (testMessage) Reset() {
	panic("should not be called")
}

func (testMessage) String() string {
	panic("should not be called")
}

func (testMessage) UnmarshalFrame(can.Frame) error {
	panic("should not be called")
}

func newDriverHeartbeatDescriptor() *descriptor.Message {
	return &descriptor.Message{
		Name:        "DriverHeartbeat",
		SenderNode:  "DRIVER",
		ID:          100,
		Length:      1,
		Description: "Sync message used to synchronize the controllers",
		SendType:    descriptor.SendTypeCyclic,
		CycleTime:   100 * time.Millisecond,
		DelayTime:   2 * time.Second,
		Signals: []*descriptor.Signal{
			{
				Name:   "Command",
				Start:  0,
				Length: 8,
				Scale:  1,
				ValueDescriptions: []*descriptor.ValueDescription{
					{Value: 0, Description: "None"},
					{Value: 1, Description: "Sync"},
					{Value: 2, Description: "Reboot"},
				},
				ReceiverNodes: []string{"SENSOR", "MOTOR"},
			},
		},
	}
}
