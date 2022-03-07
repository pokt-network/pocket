package utils

import (
	"errors"
	"net"
)

var (
	Unspecified error = errors.New("misc: could not resolve ip address. Either ip is loopback or unspecified")
	NilIp       error = errors.New("misc: could not stringify ip. IP is nil")
)

func ResolveIP(ip net.IP) (net.IP, error) {
	if ip.IsLoopback() || ip.IsUnspecified() {
		return nil, Unspecified
	}

	return ip, nil
}

func StringIP(ip net.IP) (string, error) {
	ip, err := ResolveIP(ip)

	if err != nil {
		return "", err
	}

	str := ip.String()

	if str == "<nil>" {
		return "", NilIp
	}

	return str, nil
}

func ResolveAddress(address string) (string, error) {
	addr, err := net.ResolveTCPAddr("tcp", address)

	if err != nil {
		return "", err
	}

	ip, err := ResolveIP(addr.IP)
	if err != nil {
		return "", err
	}

	addr.IP = ip

	return addr.String(), nil
}
