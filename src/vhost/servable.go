package	vhost

import (
)


type	Servable	struct {
	Owner		string
	Project		string
	Zone		string

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
