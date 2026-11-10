package tcpportscan

import "net"

type job struct {
	host string
	ip   net.IP
	port uint16
}
