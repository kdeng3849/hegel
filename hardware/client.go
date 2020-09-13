package hardware

import (
	"context"
	"encoding/json"
	"flag"

	cacherClient "github.com/packethost/cacher/client"
	"github.com/packethost/cacher/protos/cacher"
	"github.com/packethost/pkg/env"
	"github.com/pkg/errors"
	tinkClient "github.com/tinkerbell/tink/client"
	tink "github.com/tinkerbell/tink/protos/hardware"
	"github.com/tinkerbell/tink/util"
	"google.golang.org/grpc"
)

var (
	dataModelVersion = env.Get("DATA_MODEL_VERSION")
	facility         = flag.String("facility", env.Get("HEGEL_FACILITY", "onprem"), "The facility we are running in (mostly to connect to cacher)")
)

// Client acts as the messenger between Hegel and Cacher/Tink
type Client interface {
	ByIP(ctx context.Context, id string, opts ...grpc.CallOption) (Hardware, error)
	Watch(ctx context.Context, id string, opts ...grpc.CallOption) (Watcher, error)
}

// Hardware is the interface for Cacher/Tink hardware types
type Hardware interface {
	Export() ([]byte, error)
	ID() (string, error)
}

// Watcher is the interface for Cacher/Tink watch client types
type Watcher interface {
	Recv() (Hardware, error)
}

type clientCacher struct {
	client cacher.CacherClient
}

type clientTinkerbell struct {
	client tink.HardwareServiceClient
}

type hardwareCacher struct {
	hardware *cacher.Hardware
}

type hardwareTinkerbell struct {
	hardware *tink.Hardware
}

type watcherCacher struct {
	client cacher.Cacher_WatchClient
}

type watcherTinkerbell struct {
	client tink.HardwareService_WatchClient
}

// New returns a new hardware Client, configured appropriately according to the mode (Cacher or Tink) Hegel is running in
func New() (Client, error) {
	var hg Client

	switch dataModelVersion {
	case "1":
		tc, err := tinkClient.TinkHardwareClient()
		if err != nil {
			return nil, errors.Wrap(err, "failed to create the tink client")
		}
		hg = clientTinkerbell{client: tc}
	default:
		cc, err := cacherClient.New(*facility)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create the cacher client")
		}
		hg = clientCacher{client: cc}
	}

	return hg, nil
}

// ByIP retrieves from Cacher the piece of hardware with the specified IP
func (hg clientCacher) ByIP(ctx context.Context, ip string, opts ...grpc.CallOption) (Hardware, error) {
	in := &cacher.GetRequest{
		IP: ip,
	}
	hw, err := hg.client.ByIP(ctx, in, opts...)
	if err != nil {
		return nil, err
	}
	return &hardwareCacher{hw}, nil
}

// Watch returns a Cacher watch client on the hardware with the specified ID
func (hg clientCacher) Watch(ctx context.Context, id string, opts ...grpc.CallOption) (Watcher, error) {
	in := &cacher.GetRequest{
		ID: id,
	}
	w, err := hg.client.Watch(ctx, in, opts...)
	if err != nil {
		return nil, err
	}
	return &watcherCacher{w}, nil
}

// ByIP retrieves from Tink the piece of hardware with the specified IP
func (hg clientTinkerbell) ByIP(ctx context.Context, ip string, opts ...grpc.CallOption) (Hardware, error) {
	in := &tink.GetRequest{
		Ip: ip,
	}
	hw, err := hg.client.ByIP(ctx, in, opts...)
	if err != nil {
		return nil, err
	}
	return &hardwareTinkerbell{hw}, nil
}

// Watch returns a Tink watch client on the hardware with the specified ID
func (hg clientTinkerbell) Watch(ctx context.Context, id string, opts ...grpc.CallOption) (Watcher, error) {
	in := &tink.GetRequest{
		Id: id,
	}
	w, err := hg.client.Watch(ctx, in, opts...)
	if err != nil {
		return nil, err
	}
	return &watcherTinkerbell{w}, nil
}

// Export formats the piece of hardware to be returned in responses to clients
func (hw *hardwareCacher) Export() ([]byte, error) {
	exported := &exportedHardwareCacher{}

	err := json.Unmarshal([]byte(hw.hardware.JSON), exported)
	if err != nil {
		return nil, err
	}
	return json.Marshal(exported)
}

// ID returns the hardware ID
func (hw *hardwareCacher) ID() (string, error) {
	hwJSON := make(map[string]interface{})
	err := json.Unmarshal([]byte(hw.hardware.JSON), &hwJSON)
	if err != nil {
		return "", err
	}

	hwID := hwJSON["id"]
	id := hwID.(string)

	return id, err
}

// Export formats the piece of hardware to be returned in responses to clients
func (hw *hardwareTinkerbell) Export() ([]byte, error) {
	return json.Marshal(util.HardwareWrapper{Hardware: hw.hardware})
}

// ID returns the hardware ID
func (hw *hardwareTinkerbell) ID() (string, error) {
	return hw.hardware.Id, nil
}

// Recv receives a piece of hardware from the Cacher watch client
func (w *watcherCacher) Recv() (Hardware, error) {
	hw, err := w.client.Recv()
	if err != nil {
		return nil, err
	}
	return &hardwareCacher{hw}, nil
}

// Recv receives a piece of hardware from the Tink watch client
func (w *watcherTinkerbell) Recv() (Hardware, error) {
	hw, err := w.client.Recv()
	if err != nil {
		return nil, err
	}
	return &hardwareTinkerbell{hw}, nil
}