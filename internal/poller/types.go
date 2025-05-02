package poller

import (
	"github.com/weeb-vip/gateway-proxy/internal/container"
	"github.com/weeb-vip/gateway-proxy/internal/keys"
	"time"
)

type keysPoller struct {
	container container.Container[[]keys.Key]
	fetcher   keys.Fetcher
}

type KeyManager interface {
	Fetch() error
	SetupBackgroundPolling(pollingDuration time.Duration)
	FindKeyByID(id string) (*keys.Key, error)
}
