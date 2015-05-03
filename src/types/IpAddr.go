package	types

import (
	"bytes"
	"errors"
	"net"
)


type	IpAddr	net.IP

func (d *IpAddr) UnmarshalTOML(data []byte) error  {

	dest := IpAddr(net.ParseIP(string(bytes.Trim(data,"\""))))

	if dest == nil {
		return errors.New("invalid IpAddr : "+string(data))
	}
	*d = dest

	return nil
}

func (d *IpAddr) ToTCPAddr(port string) (*net.TCPAddr, error)   {
	ip := net.IP(*d)
	if ip.To4() == nil {
		return net.ResolveTCPAddr( "tcp", "["+ip.String()+"]:"+port )
	}
	return net.ResolveTCPAddr( "tcp", ip.String()+":"+port )
}
