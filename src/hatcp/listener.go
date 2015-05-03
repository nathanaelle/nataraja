package hatcp

import	(
	"net"
	"os"
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
