package ios

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"

	goios "github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/forward"
	"github.com/mobile-next/mobilecli/utils"
)

type PortForwarder struct {
	udid           string
	listener       net.Listener
	cancelForward  context.CancelFunc
	forwardWorkers sync.WaitGroup
	forwardMutex   sync.Mutex
	srcPort        int
	dstPort        int
}

type proxyConnectionFunc func(context.Context, io.ReadWriteCloser, int, uint16) error

var startProxyConnection proxyConnectionFunc = forward.StartNewProxyConnection

func NewPortForwarder(udid string) *PortForwarder {
	return &PortForwarder{
		udid: udid,
	}
}

func (pf *PortForwarder) Forward(srcPort, dstPort int) error {
	pf.forwardMutex.Lock()
	defer pf.forwardMutex.Unlock()

	if pf.listener != nil {
		return fmt.Errorf("port forwarding is already running from %d to %d", pf.srcPort, pf.dstPort)
	}

	if srcPort < 1 || srcPort > 65535 {
		return fmt.Errorf("invalid source port %d: must be between 1 and 65535", srcPort)
	}

	if dstPort < 1 || dstPort > 65535 {
		return fmt.Errorf("invalid destination port %d: must be between 1 and 65535", dstPort)
	}

	pf.srcPort = srcPort
	pf.dstPort = dstPort

	device, err := goios.GetDevice(pf.udid)
	if err != nil {
		return fmt.Errorf("failed to get device %s: %w", pf.udid, err)
	}

	listener, err := listenLoopback(srcPort)
	if err != nil {
		return fmt.Errorf("failed to create loopback port forwarder: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	pf.listener = listener
	pf.cancelForward = cancel
	pf.forwardWorkers.Add(1)
	go pf.acceptConnections(ctx, device.DeviceID, uint16(dstPort), listener)
	utils.Verbose("Loopback port forwarding started from 127.0.0.1:%d to device port %d", srcPort, dstPort)

	return nil
}

func (pf *PortForwarder) Stop() error {
	pf.forwardMutex.Lock()
	defer pf.forwardMutex.Unlock()

	if pf.listener == nil {
		return fmt.Errorf("no port forwarding running")
	}

	pf.cancelForward()
	err := pf.listener.Close()
	if err != nil {
		utils.Verbose("Error stopping port forwarding %d->%d: %v", pf.srcPort, pf.dstPort, err)
	}
	pf.forwardWorkers.Wait()

	utils.Verbose("Stopping port forwarding %d->%d", pf.srcPort, pf.dstPort)
	pf.listener = nil
	pf.cancelForward = nil
	pf.srcPort = 0
	pf.dstPort = 0

	return err
}

func (pf *PortForwarder) IsRunning() bool {
	pf.forwardMutex.Lock()
	defer pf.forwardMutex.Unlock()

	return pf.listener != nil
}

func (pf *PortForwarder) acceptConnections(ctx context.Context, deviceID int, dstPort uint16, listener net.Listener) {
	defer pf.forwardWorkers.Done()

	for {
		clientConn, err := listener.Accept()
		if err != nil {
			if ctx.Err() != nil || isClosedNetworkError(err) {
				return
			}
			utils.Verbose("Error accepting forwarded connection: %v", err)
			continue
		}

		pf.forwardWorkers.Add(1)
		go func() {
			defer pf.forwardWorkers.Done()
			if err := startProxyConnection(ctx, clientConn, deviceID, dstPort); err != nil && ctx.Err() == nil {
				utils.Verbose("Forwarded connection ended: %v", err)
			}
		}()
	}
}

func isClosedNetworkError(err error) bool {
	return errors.Is(err, net.ErrClosed)
}

func listenLoopback(port int) (net.Listener, error) {
	return net.ListenTCP("tcp4", &net.TCPAddr{
		IP:   net.IPv4(127, 0, 0, 1),
		Port: port,
	})
}

func (pf *PortForwarder) GetPorts() (srcPort, dstPort int) {
	pf.forwardMutex.Lock()
	defer pf.forwardMutex.Unlock()

	return pf.srcPort, pf.dstPort
}
