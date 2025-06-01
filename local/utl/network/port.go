package network

import (
	"fmt"
	"net"
	"time"
)

// IsPortAvailable checks if a port is available for both TCP and UDP
func IsPortAvailable(port int) bool {
	return IsTCPPortAvailable(port) && IsUDPPortAvailable(port)
}

// IsTCPPortAvailable checks if a TCP port is available
func IsTCPPortAvailable(port int) bool {
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return false
	}
	listener.Close()
	return true
}

// IsUDPPortAvailable checks if a UDP port is available
func IsUDPPortAvailable(port int) bool {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{Port: port})
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// FindAvailablePort finds an available port starting from the given port
func FindAvailablePort(startPort int) (int, error) {
	maxPort := 65535
	for port := startPort; port <= maxPort; port++ {
		if IsPortAvailable(port) {
			return port, nil
		}
	}
	return 0, fmt.Errorf("no available ports found between %d and %d", startPort, maxPort)
}

// FindAvailablePortRange finds a range of consecutive available ports
func FindAvailablePortRange(startPort, count int) ([]int, error) {
	maxPort := 65535
	ports := make([]int, 0, count)
	currentPort := startPort

	for len(ports) < count && currentPort <= maxPort {
		// Check if we have enough consecutive ports available
		available := true
		for i := 0; i < count-len(ports); i++ {
			if !IsPortAvailable(currentPort + i) {
				available = false
				currentPort += i + 1
				break
			}
		}

		if available {
			for i := 0; i < count-len(ports); i++ {
				ports = append(ports, currentPort+i)
			}
		}
	}

	if len(ports) < count {
		return nil, fmt.Errorf("could not find %d consecutive available ports starting from %d", count, startPort)
	}

	return ports, nil
}

// WaitForPortAvailable waits for a port to become available with timeout
func WaitForPortAvailable(port int, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if IsPortAvailable(port) {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("timeout waiting for port %d to become available", port)
} 