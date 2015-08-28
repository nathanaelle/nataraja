// +build darwin dragonfly freebsd netbsd openbsd

package	hatcp

import	(
	"syscall"
	"time"
	"os"
)



func ka_idle(fd int, d time.Duration) error {
	if d == 0 {
		return os.NewSyscallError("ka_idle", syscall.SetsockoptInt(fd, syscall.IPPROTO_TCP, syscall.TCP_KEEPIDLE, 0 ))
	}

	// cargo cult from src/net/tcpsockopt_unix.go
	d += (time.Second - time.Nanosecond)
	return os.NewSyscallError("ka_idle", syscall.SetsockoptInt(fd, syscall.IPPROTO_TCP, syscall.TCP_KEEPIDLE, int(d.Seconds()) ))
}

func ka_intvl(fd int, d time.Duration) error {
	if d == 0 {
		return os.NewSyscallError("ka_intv", syscall.SetsockoptInt(fd, syscall.IPPROTO_TCP, syscall.TCP_KEEPINTVL, 0 ))
	}

	// cargo cult from src/net/tcpsockopt_unix.go
	d += (time.Second - time.Nanosecond)
	return os.NewSyscallError("ka_intv", syscall.SetsockoptInt(fd, syscall.IPPROTO_TCP, syscall.TCP_KEEPINTVL, int(d.Seconds()) ))
}

func ka_count(fd int, n int) error {
	return os.NewSyscallError("ka_count", syscall.SetsockoptInt(fd, syscall.IPPROTO_TCP, syscall.TCP_KEEPCNT, n))
}

func so_nodelay(fd int, flag bool) error {
	return os.NewSyscallError("so_nodelay", syscall.SetsockoptInt(fd, syscall.IPPROTO_TCP, syscall.TCP_NODELAY, boolint(flag)))
}

func so_tcpcork(fd int, flag bool) error {
	return nil
}

func so_tcpnopush(fd int, flag bool) error {
	return os.NewSyscallError("so_tcpnopush", syscall.SetsockoptInt(fd, syscall.IPPROTO_TCP, syscall.TCP_NOPUSH, boolint(flag)))
}

func so_reuseport(fd int, flag bool) error {
	return os.NewSyscallError("so_reuseport", syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_REUSEPORT, boolint(flag)) )
}

func so_reuseaddr(fd int, flag bool) error {
	return os.NewSyscallError("so_reuseaddr", syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, boolint(flag)) )
}

func so_fastopen(fd int, flag bool) error {
//	return os.NewSyscallError("so_fastopen", syscall.SetsockoptInt(fd, syscall.IPPROTO_TCP, TCP_FASTOPEN, boolint(flag)) )
	return nil
}

func so_nonblock(fd int, flag bool) error {
	return os.NewSyscallError("so_nonblock", syscall.SetNonblock(fd, true))
}

func so_linger(fd int, d time.Duration) error {
	if d == 0 {
		return os.NewSyscallError("so_linger", syscall.SetsockoptLinger(fd, syscall.SOL_SOCKET, syscall.SO_LINGER, &syscall.Linger { 0, 0 } ))
	}

	// cargo cult from src/net/tcpsockopt_unix.go
	d	+= (time.Second - time.Nanosecond)
	l	:= syscall.Linger { 1, int32(d.Seconds()) }

	return os.NewSyscallError("so_linger", syscall.SetsockoptLinger(fd, syscall.SOL_SOCKET, syscall.SO_LINGER, &l ))
}

func so_listen(fd int,queue int) error {
	if queue <1 {
		return os.NewSyscallError("so_listen", syscall.Listen(fd, syscall.SOMAXCONN) )
	}

	return os.NewSyscallError("so_listen", syscall.Listen(fd, queue) )

}
