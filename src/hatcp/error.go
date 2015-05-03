package hatcp

import	(
	"fmt"
)


const	(
	Err_Unsupported_Proto		int = iota
)


type	HaTcpErr	struct {
	EType	int
	Proto	string
}


func (e HaTcpErr)Error()string  {
	err	:= ""
	switch e.EType {
		case	Err_Unsupported_Proto:
			err	= fmt.Sprintf("Unknown proto %s", e.Proto )
	}
	return	err
}


func unknown_proto(p string) error  {
	return HaTcpErr { Err_Unsupported_Proto, p }
}
