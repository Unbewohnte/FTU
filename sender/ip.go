package sender

import (
	"fmt"
	"io"
	"net"
	"net/http"
)

// gets a local ip. Borrowed from StackOverflow, thank you, whoever I brought it from
func GetLocalIP() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String(), nil
}

// gets a remote ip. Borrowed from StackOverflow, thank you, whoever I brought it from
func GetRemoteIP() (string, error) {
	resp, err := http.Get("https://api.ipify.org?format=text")
	if err != nil {
		return "", fmt.Errorf("could not make a request to get your remote IP: %s", err)
	}
	defer resp.Body.Close()
	ip, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("could not read a response: %s", err)
	}
	return string(ip), nil
}
