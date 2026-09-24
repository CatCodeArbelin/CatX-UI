package analytics

import (
	"context"
	"fmt"

	"github.com/mhsanaei/3x-ui/v3/internal/xray"
)

// XrayRuntimeSource is implemented by the existing Xray API/runtime. The
// adapter reads it and does not persist or maintain a second counter/state map.
type XrayRuntimeSource interface {
	GetOnlineUsers() ([]xray.OnlineUser, error)
	GetTraffic() ([]*xray.Traffic, []*xray.ClientTraffic, error)
}

type RuntimeSnapshot struct {
	OnlineUsers   []xray.OnlineUser
	Traffic       []*xray.Traffic
	ClientTraffic []*xray.ClientTraffic
}

type XrayRuntimeAdapter struct{ source XrayRuntimeSource }

func NewXrayRuntimeAdapter(source XrayRuntimeSource) *XrayRuntimeAdapter {
	return &XrayRuntimeAdapter{source: source}
}

func (a *XrayRuntimeAdapter) Snapshot(ctx context.Context) (RuntimeSnapshot, error) {
	if a == nil || a.source == nil {
		return RuntimeSnapshot{}, fmt.Errorf("xray runtime source is nil")
	}
	select {
	case <-ctx.Done():
		return RuntimeSnapshot{}, ctx.Err()
	default:
	}
	online, err := a.source.GetOnlineUsers()
	if err != nil {
		return RuntimeSnapshot{}, err
	}
	traffic, clients, err := a.source.GetTraffic()
	if err != nil {
		return RuntimeSnapshot{}, err
	}
	return RuntimeSnapshot{OnlineUsers: online, Traffic: traffic, ClientTraffic: clients}, nil
}
