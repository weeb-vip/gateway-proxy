package generics_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/weeb-vip/gateway-proxy/internal/generics"
	"testing"
)

func TestFirst(t *testing.T) {
	t.Run("returns pointer to the first found item", func(t *testing.T) {
		search := generics.First([]int{1, 2, 3, 4, 5}, func(item int) bool {
			return item == 2
		})
		assert.Equal(t, 2, *search)
	})
	t.Run("returns nil when not found", func(t *testing.T) {
		search := generics.First([]int{1, 2, 3, 4, 5}, func(item int) bool {
			return item == 10
		})
		assert.Nil(t, search)
	})
}
