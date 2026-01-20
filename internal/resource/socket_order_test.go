package resource

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSocketManager_GetFiles_Order_Deterministic(t *testing.T) {
	sm := NewSocketManager()
	defer sm.Close()

	addrs := []string{
		"127.0.0.1:9095",
		"127.0.0.1:9091",
		"127.0.0.1:9093",
		"127.0.0.1:9092",
		"127.0.0.1:9094",
	}

	for _, addr := range addrs {
		_, err := sm.EnsureListener(addr)
		require.NoError(t, err)
	}

	var orders [][]uintptr
	for i := 0; i < 100; i++ {
		files := sm.GetFiles()
		require.Equal(t, len(addrs), len(files))

		fds := make([]uintptr, len(files))
		for j, f := range files {
			fds[j] = f.Fd()
		}
		orders = append(orders, fds)
	}

	// Check if all orders are the same
	for i := 1; i < len(orders); i++ {
		assert.Equal(t, orders[0], orders[i], "GetFiles() order should be deterministic")
	}
}
