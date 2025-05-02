package poller

import (
	"errors"
	"time"

	"github.com/weeb-vip/gateway-proxy/internal/container"
	"github.com/weeb-vip/gateway-proxy/internal/generics"
	"github.com/weeb-vip/gateway-proxy/internal/keys"
)

func (k keysPoller) Fetch() error {
	result, err := k.fetcher.FetchKeys()
	if err != nil {
		return err
	}

	k.container.ReplaceWith(result)

	return nil
}

func (k keysPoller) SetupBackgroundPolling(pollDuration time.Duration) {
	go func() {
		for {
			time.Sleep(pollDuration)
			_ = k.Fetch() // we can safely ignore the error
		}
	}()
}

func (k keysPoller) FindKeyByID(id string) (*keys.Key, error) {
	// try to find
	// if not found, fetch once
	// try to find again
	// if not found, it's an error

	key := k.findKeyByID(id)
	if key != nil {
		return key, nil
	}
	err := k.Fetch()
	if err != nil {
		return nil, err
	}
	key = k.findKeyByID(id)
	if key == nil {
		return nil, errors.New("key couldn't be found")
	}

	return key, nil
}

func (k keysPoller) findKeyByID(id string) *keys.Key {
	return generics.First(k.container.GetLatest(), func(item keys.Key) bool {
		return item.ID == id
	})
}

func Keys(fetcher keys.Fetcher) (KeyManager, error) {
	key, err := fetcher.FetchKeys()
	if err != nil {
		return nil, err
	}
	c := container.New(key)
	return keysPoller{
		container: c,
		fetcher:   fetcher,
	}, nil
}
