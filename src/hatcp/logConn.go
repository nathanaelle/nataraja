package hatcp

import	(
	"net"
)


type	LogConn struct {
	*net.TCPConn
	BytesRead	int64
	BytesWrote	int64
}


func (lc *LogConn) Read(b []byte) (n int, err error){
	n,err	= lc.TCPConn.Read(b)
	lc.BytesRead+=int64(n)
	return
}

// Write writes data to the connection.
// Write can be made to time out and return a Error with Timeout() == true
// after a fixed time limit; see SetDeadline and SetWriteDeadline.
func (lc *LogConn) Write(b []byte) (n int, err error){
	n,err	= lc.TCPConn.Write(b)
	lc.BytesWrote+=int64(n)
	return
}
