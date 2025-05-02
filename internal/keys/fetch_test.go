//go:build integration

package keys_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/weeb-vip/gateway-proxy/internal/keys"
	"testing"
)

// this file is excluded for now in CI, I'm running this as a validation step for myself
// later, we'll most likely introduce proper integration test

func TestNewKeyFetcher(t *testing.T) {
	fetcher := keys.NewFetcher("http://localhost:5000/graphql")
	k, err := fetcher.FetchKeys()
	assert.NoError(t, err)
	assert.NotEmpty(t, k)
}
