package container_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/weeb-vip/gateway-proxy/internal/container"
	"testing"
)

func TestNew(t *testing.T) {
	t.Run("container can replace item and get latest item at any point and return correct data", func(t *testing.T) {
		c := container.New("1")
		assert.Equal(t, "1", c.GetLatest())
		c.ReplaceWith("2")
		assert.Equal(t, "2", c.GetLatest())
	})
}
