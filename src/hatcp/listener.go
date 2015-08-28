package hatcp

import	(
	"os"
	"net"
	"time"
	"strconv"
	"strings"
	"syscall"
)

const	newfile_prefix	string = "prefix_newfile"

func boolint(b bool) int {
	switch b {
		case true:	return 1
		default:	return 0
	}
}


type HATCPListener struct {
	net.Listener
}


func Listen(proto string, laddr *net.TCPAddr) (ln *HATCPListener, err error)  {
	fd, err	:= system_HaTcpListener(proto, laddr)

	if err != nil {
		return nil, err
	}

	file	:= os.NewFile(uintptr(fd), strings.Join([]string { newfile_prefix, strconv.Itoa(fd),strconv.Itoa(os.Getpid()) }, "_" ) )

	l, err := net.FileListener(file);
	if  err != nil {
		syscall.Close(fd)
		return nil, err
	}
	ln	= &HATCPListener { l }


	if err = file.Close(); err != nil {
		syscall.Close(fd)
		return nil, err
	}

	return
}


func system_HaTcpListener(net string, laddr *net.TCPAddr) (fd int, err error) {
	switch {
		case laddr.IP.To4() != nil:
			var addr	[4]byte
			copy(addr[:], laddr.IP[len(laddr.IP)-4:len(laddr.IP)] )

			return  generic_HaTcpListener(
				func() (int,error){
					return syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, syscall.IPPROTO_TCP)
				},
				func(fd int) error{
					return syscall.Bind(fd, &syscall.SockaddrInet4{Port: laddr.Port, Addr: addr })
				})


		case laddr.IP.To16() != nil:
			var addr	[16]byte
			copy(addr[:], laddr.IP )

			return generic_HaTcpListener(
				func() (int,error){
					return syscall.Socket(syscall.AF_INET6, syscall.SOCK_STREAM, syscall.IPPROTO_TCP)
				},
				func(fd int) error{
					return syscall.Bind(fd, &syscall.SockaddrInet6{Port: laddr.Port, Addr: addr })
				})

		default:
			return	-1, unknown_proto(net)
	}
}


func generic_HaTcpListener( generic_create func() (int,error), generic_bind func(int) error) (fd int, err error){
	if fd, err = generic_create(); err != nil {
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
		{ bullet_nil(generic_bind)	, nil		},
		{ bullet_int(so_listen)		, -1		},
		{ bullet_bool(so_nodelay)	, true		},
		{ bullet_bool(so_tcpcork)	, false		},
		{ bullet_bool(so_tcpnopush)	, false		},
		{ bullet_bool(so_nonblock)	, true		},
	}, fd )

	if err != nil {
		syscall.Close(fd)
		return -1, err
	}

	return
}
