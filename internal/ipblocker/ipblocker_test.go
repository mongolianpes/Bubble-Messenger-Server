package ipblocker

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestIPBlocker(t *testing.T) {
	ipblocker := New(2, time.Second)

	user1IP := "1.1.1.1"
	user2Ip := "2.2.2.2"
	ipblocker.RegisterStrike(user1IP)

	require.Equal(t, false, ipblocker.IsBlocked(user1IP))

	ipblocker.RegisterStrike(user1IP)

	require.Equal(t, true, ipblocker.IsBlocked(user1IP))
	require.Equal(t, false, ipblocker.IsBlocked(user2Ip))
}
