package poller_test

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/weeb-vip/gateway-proxy/internal/keys"
	"github.com/weeb-vip/gateway-proxy/internal/poller"
	"testing"
)

type mockFetcher struct {
	k       []keys.Key
	e       error
	counter int
}

func (r *mockFetcher) FetchKeys() ([]keys.Key, error) {
	r.counter = r.counter + 1
	return r.k, r.e
}

func TestKeys(t *testing.T) {
	t.Run("returns error while initializing if the initial fetching fails", func(t *testing.T) {
		p, err := poller.Keys(&mockFetcher{k: nil, e: errors.New("some error")})
		assert.Error(t, err)
		assert.Nil(t, p)
	})
	t.Run("if there was no error on initial loading, it initializes and the initial data can be queried", func(t *testing.T) {
		p, err := poller.Keys(&mockFetcher{k: []keys.Key{{ID: "id1", Body: "body1"}, {ID: "id2", Body: "body2"}}, e: nil})
		assert.NoError(t, err)
		key, err := p.FindKeyByID("id2")
		assert.NoError(t, err)
		assert.Equal(t, "id2", key.ID)
	})
	t.Run("once initialized, querying for a key that doesn't exist, returns in error after calling FetchKeys for second time", func(t *testing.T) {
		successMockFetcher := &mockFetcher{k: []keys.Key{{ID: "id1", Body: "body1"}, {ID: "id2", Body: "body2"}}, e: nil}
		p, err := poller.Keys(successMockFetcher)
		assert.NoError(t, err)
		key, err := p.FindKeyByID("id3")
		assert.Error(t, err)
		assert.Nil(t, key)
		assert.Equal(t, 2, successMockFetcher.counter)
	})
}
