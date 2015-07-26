package	vhost

import (
	"../types"
)


type	Servable	struct {
	Owner		string
	Project		string
	Zone		string
	Proxied		types.URL

	Redirect	string

	TLS		bool
	HSTS		string
	XFO		string
	XCTO		string
	XDO		string
	XXSSP		string
	PKP		string
	CSP		string
}
