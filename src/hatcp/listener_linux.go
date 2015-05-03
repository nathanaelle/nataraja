package	hatcp

import	(
	"net"
	"syscall"
	"time"
	"os"
)

const	SO_REUSEPORT	= 15
const	TCP_FASTOPEN	= 23



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
	return os.NewSyscallError("so_tcpcork", syscall.SetsockoptInt(fd, syscall.IPPROTO_TCP, syscall.TCP_CORK, boolint(flag)))
}

func so_reuseport(fd int, flag bool) error {
	return os.NewSyscallError("so_reuseport", syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, SO_REUSEPORT, boolint(flag)) )
}

func so_reuseaddr(fd int, flag bool) error {
	return os.NewSyscallError("so_reuseaddr", syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, boolint(flag)) )
}

func so_fastopen(fd int, flag bool) error {
	return os.NewSyscallError("so_fastopen", syscall.SetsockoptInt(fd, syscall.IPPROTO_TCP, TCP_FASTOPEN, boolint(flag)) )
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



func system_HaTcpListener(net string, laddr *net.TCPAddr) (fd int, err error) {
	switch {
		case laddr.IP.To4() != nil:
			var addr	[4]byte
			copy(addr[:], laddr.IP[len(laddr.IP)-4:len(laddr.IP)] )

			if fd, err = syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, syscall.IPPROTO_TCP); err != nil {
				return -1, err
			}

			err = gatling_run( gatling{
				{ bullet_bool(so_reuseaddr)	, true		},
				{ bullet_bool(so_reuseport)	, true		},
				{ bullet_bool(so_fastopen)	, true		},
				{ bullet_duration(ka_idle)	, 10*time.Second},
				{ bullet_duration(ka_intvl)	, 5*time.Second	},
				{ bullet_int(ka_count)		, 10		},
				{ bullet_duration(so_linger)	, 3*time.Second	},
			}, fd )

			if err != nil {
				syscall.Close(fd)
				return -1, err
			}

			if err = syscall.Bind(fd, &syscall.SockaddrInet4{Port: laddr.Port, Addr: addr }); err != nil {
				syscall.Close(fd)
				return -1, err
			}

			err = gatling_run( gatling{
				{ bullet_int(so_listen)		, -1		},
				{ bullet_bool(so_nodelay)	, true		},
				{ bullet_bool(so_tcpcork)	, false		},
				{ bullet_bool(so_nonblock)	, true		},
			}, fd )

			if err != nil {
				syscall.Close(fd)
				return -1, err
			}

			return


		case laddr.IP.To16() != nil:
			var addr	[16]byte
			copy(addr[:], laddr.IP )
			if fd, err = syscall.Socket(syscall.AF_INET6, syscall.SOCK_STREAM, syscall.IPPROTO_TCP); err != nil {
				return -1, err
			}

			err = gatling_run( gatling{
				{ bullet_bool(so_reuseaddr)	, true		},
				{ bullet_bool(so_reuseport)	, true		},
				{ bullet_bool(so_fastopen)	, true		},
				{ bullet_duration(ka_idle)	, 10*time.Second},
				{ bullet_duration(ka_intvl)	, 5*time.Second	},
				{ bullet_int(ka_count)		, 10		},
				{ bullet_duration(so_linger)	, 3*time.Second	},
			}, fd )

			if err != nil {
				syscall.Close(fd)
				return -1, err
			}

			if err = syscall.Bind(fd, &syscall.SockaddrInet6{Port: laddr.Port, Addr: addr }); err != nil {
				syscall.Close(fd)
				return -1, err
			}

			err = gatling_run( gatling{
				{ bullet_int(so_listen)		, -1		},
				{ bullet_bool(so_nodelay)	, true		},
				{ bullet_bool(so_tcpcork)	, false		},
				{ bullet_bool(so_nonblock)	, true		},
			}, fd )

			if err != nil {
				syscall.Close(fd)
				return -1, err
			}

			return

		default:
			return	-1, unknown_proto(net)
	}
}
