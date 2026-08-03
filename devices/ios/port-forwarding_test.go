package ios

import (
	"context"
	"io"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPortForwarderListensOnLoopbackOnly(t *testing.T) {
	listener, err := listenLoopback(0)
	require.NoError(t, err)
	t.Cleanup(func() { _ = listener.Close() })

	address := listener.Addr().(*net.TCPAddr)
	require.True(t, address.IP.IsLoopback())
	require.Equal(t, "127.0.0.1", address.IP.String())
}

func TestPortForwarderRejectsZeroPortsBeforeDeviceLookup(t *testing.T) {
	forwarder := NewPortForwarder("not-a-device")

	require.EqualError(t, forwarder.Forward(0, 8100), "invalid source port 0: must be between 1 and 65535")
	require.EqualError(t, forwarder.Forward(8100, 0), "invalid destination port 0: must be between 1 and 65535")
}

func TestAcceptConnectionsUsesSelectedDeviceAndPort(t *testing.T) {
	listener, err := listenLoopback(0)
	require.NoError(t, err)

	type proxyCall struct {
		deviceID int
		dstPort  uint16
	}
	called := make(chan proxyCall, 1)
	originalStartProxyConnection := startProxyConnection
	startProxyConnection = func(_ context.Context, connection io.ReadWriteCloser, deviceID int, dstPort uint16) error {
		defer connection.Close()
		called <- proxyCall{deviceID: deviceID, dstPort: dstPort}
		return nil
	}
	t.Cleanup(func() { startProxyConnection = originalStartProxyConnection })

	ctx, cancel := context.WithCancel(context.Background())
	forwarder := &PortForwarder{}
	forwarder.forwardWorkers.Add(1)
	go forwarder.acceptConnections(ctx, 42, 8100, listener)

	connection, err := net.Dial("tcp4", listener.Addr().String())
	require.NoError(t, err)
	require.NoError(t, connection.Close())

	select {
	case call := <-called:
		require.Equal(t, 42, call.deviceID)
		require.Equal(t, uint16(8100), call.dstPort)
	case <-time.After(time.Second):
		t.Fatal("forwarder did not accept the loopback connection")
	}

	cancel()
	require.NoError(t, listener.Close())
	forwarder.forwardWorkers.Wait()
}
