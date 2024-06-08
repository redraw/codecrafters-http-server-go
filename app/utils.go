package main

import (
	"fmt"
	"net"
	"time"
)

func checkPortOpen(host string, port int, timeout time.Duration) bool {
	address := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func waitForPortOpen(host string, port int, interval time.Duration, timeout time.Duration) bool {
	timeoutChan := time.After(timeout)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutChan:
			return false
		case <-ticker.C:
			if checkPortOpen(host, port, interval) {
				return true
			}
		}
	}
}
