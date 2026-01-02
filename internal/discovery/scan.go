package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"
)

// DiscoverViaNetworkScan scans the local network for KEF speakers.
func DiscoverViaNetworkScan(ctx context.Context, timeout time.Duration) (string, error) {
	localIPs, err := getLocalIPs()
	if err != nil {
		return "", err
	}

	if len(localIPs) == 0 {
		return "", fmt.Errorf("no local network interfaces found")
	}

	scanCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	type result struct {
		ip  string
		err error
	}

	resultChan := make(chan result, 1)
	var wg sync.WaitGroup

	client := &http.Client{
		Timeout: 1 * time.Second,
	}

	// Scan each local network
	for _, localIP := range localIPs {
		ip := localIP.To4()
		if ip == nil {
			continue
		}

		networkPrefix := fmt.Sprintf("%d.%d.%d", ip[0], ip[1], ip[2])

		// Scan IPs 1-254
		for i := 1; i <= 254; i++ {
			select {
			case <-scanCtx.Done():
				goto waitForResults
			default:
			}

			testIP := fmt.Sprintf("%s.%d", networkPrefix, i)
			wg.Add(1)

			go func(ipAddr string) {
				defer wg.Done()

				select {
				case <-scanCtx.Done():
					return
				default:
				}

				if isKEFSpeaker(scanCtx, client, ipAddr) {
					select {
					case resultChan <- result{ip: ipAddr, err: nil}:
					case <-scanCtx.Done():
					}
				}
			}(testIP)
		}
	}

waitForResults:
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case res := <-resultChan:
		return res.ip, res.err
	case <-scanCtx.Done():
		if scanCtx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("network scan timeout")
		}
		return "", scanCtx.Err()
	case <-done:
		return "", fmt.Errorf("speaker not found on network")
	}
}

// getLocalIPs returns all local IPv4 addresses.
func getLocalIPs() ([]net.IP, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("failed to get interfaces: %v", err)
	}

	var localIPs []net.IP
	for _, iface := range interfaces {
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			if ipNet, ok := addr.(*net.IPNet); ok && ipNet.IP.To4() != nil {
				localIPs = append(localIPs, ipNet.IP)
			}
		}
	}

	return localIPs, nil
}

// isKEFSpeaker checks if the given IP hosts a KEF speaker.
func isKEFSpeaker(ctx context.Context, client *http.Client, ipAddr string) bool {
	apiURL := fmt.Sprintf("http://%s:80/api/getData?path=settings:/deviceName&roles=value", ipAddr)

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return false
	}

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != 200 {
		return false
	}

	var data []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return false
	}

	if len(data) > 0 {
		if _, ok := data[0]["string_"]; ok {
			return true
		}
	}

	return false
}

