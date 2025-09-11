package utils

import (
	"net"
)

func IsValidIPAddress(ip string) bool {
	ipAddress := net.ParseIP(ip)
	return ipAddress != nil
}
