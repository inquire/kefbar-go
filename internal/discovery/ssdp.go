package discovery

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

// SSDP constants.
const (
	ssdpMulticastAddr = "239.255.255.250:1900"
)

// DiscoverViaSSDP attempts to find a KEF speaker using SSDP multicast.
func DiscoverViaSSDP(ctx context.Context, timeout time.Duration) (string, error) {
	multicastAddr, err := net.ResolveUDPAddr("udp4", ssdpMulticastAddr)
	if err != nil {
		return "", err
	}

	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	type result struct {
		ip  string
		err error
	}

	resultChan := make(chan result, 1)
	var wg sync.WaitGroup

	// Try each interface
	for _, iface := range interfaces {
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}

		wg.Add(1)
		go func(iface net.Interface) {
			defer wg.Done()

			conn, err := net.ListenMulticastUDP("udp4", &iface, multicastAddr)
			if err != nil {
				return
			}
			defer func() { _ = conn.Close() }()

			deadline := time.Now().Add(timeout)
			_ = conn.SetReadDeadline(deadline)

			// Send M-SEARCH requests
			searchRequests := []string{
				buildMSearchRequest("upnp:rootdevice"),
				buildMSearchRequest("urn:schemas-upnp-org:device:MediaRenderer:1"),
				buildMSearchRequest("ssdp:all"),
			}

			for _, req := range searchRequests {
				select {
				case <-ctx.Done():
					return
				default:
					_, _ = conn.WriteToUDP([]byte(req), multicastAddr)
				}
			}

			// Read responses
			buffer := make([]byte, 4096)
			for {
				select {
				case <-ctx.Done():
					return
				default:
					_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
					n, addr, err := conn.ReadFromUDP(buffer)
					if err != nil {
						if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
							if time.Now().After(deadline) {
								return
							}
							continue
						}
						continue
					}

					response := strings.ToUpper(string(buffer[:n]))
					if isKEFDevice(response) {
						select {
						case resultChan <- result{ip: addr.IP.String(), err: nil}:
						case <-ctx.Done():
						}
						return
					}
				}
			}
		}(iface)
	}

	// Wait for first result or timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case res := <-resultChan:
		return res.ip, res.err
	case <-ctx.Done():
		return "", ctx.Err()
	case <-time.After(timeout):
		return "", fmt.Errorf("SSDP discovery timeout")
	case <-done:
		return "", fmt.Errorf("SSDP discovery failed - no KEF device found")
	}
}

// buildMSearchRequest creates an SSDP M-SEARCH request.
func buildMSearchRequest(searchTarget string) string {
	return "M-SEARCH * HTTP/1.1\r\n" +
		"HOST: 239.255.255.250:1900\r\n" +
		"MAN: \"ssdp:discover\"\r\n" +
		"ST: " + searchTarget + "\r\n" +
		"MX: 3\r\n" +
		"\r\n"
}

// isKEFDevice checks if the SSDP response indicates a KEF device.
func isKEFDevice(response string) bool {
	return strings.Contains(response, "KEF") ||
		strings.Contains(response, "LSX") ||
		strings.Contains(response, "LS50")
}
