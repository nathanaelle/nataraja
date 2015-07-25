package	cache

import	(
	"net/http"
)



type	(
	//InternalIP	*types.IpAddr
	//Peer		[]types.IpAddr
	//Path		types.Path
	//MemSize		types.StoreSize

	Pool	interface {
		Init()			error
		Get(*http.Request)	(*Entry,error)
	}


	PassThru	struct {
		transport	http.RoundTripper
	}
)




func (pt *PassThru) Init() error {
	pt.transport = http.DefaultTransport
	return nil
}


func (pt *PassThru) Get(req *http.Request) (*Entry,error) {
	res, err := pt.transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	return NewEntry(res.Header, res.StatusCode, res.ContentLength, res.Body ), nil
}
