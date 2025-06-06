package utils

import "net"

// from https://gist.github.com/sevkin/96bdae9274465b2d09191384f86ef39d
func GetFreePort() (freePort int, err error) {
	var a *net.TCPAddr
	if a, err = net.ResolveTCPAddr("tcp", "localhost:0"); err == nil {
		var l *net.TCPListener
		if l, err = net.ListenTCP("tcp", a); err == nil {
			defer func() { _ = l.Close() }()
			return l.Addr().(*net.TCPAddr).Port, nil
		}
	}
	return
}

