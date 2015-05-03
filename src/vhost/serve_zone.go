package	vhost

import (
	"../types"
//	"log"
//	"net/http"
)


type	ServeZone	struct {
	Zones			[]types.FQDN
	Proxied			types.URL
	TLS			*CertPair

	StrictTransportSecurity	string
	XFrameOptions		string
	XContentTypeOptions	string
	XDownloadOptions	string
	XXSSProtection		string
	PublicKeyPins		string
	ContentSecurityPolicy	string
}



func (sv ServeZone)Servable(Owner,Project string) []Servable {
	ret	:= make([]Servable, 0, len(sv.Zones))

	for _,zone := range sv.Zones {
		switch sv.TLS != nil && sv.TLS.IsEnabledFor(string(zone)) {
			case true:
				ret = append(ret, Servable {
					Owner	: Owner,
					Project	: Project,
					Zone	: string(zone),
					TLS	: true,
					HSTS	: sv.HSTS(),
					XCTO	: sv.XCTO(),
					XDO	: sv.XDO(),
					XFO	: sv.XFO(),
					XXSSP	: sv.XXSSP(),
					PKP	: sv.PKP(string(zone)),
				})


			default:
				ret = append(ret, Servable {
					Owner	: Owner,
					Project	: Project,
					Zone	: string(zone),
					TLS	: false,
					HSTS	: "",
					XCTO	: sv.XCTO(),
					XDO	: sv.XDO(),
					XFO	: sv.XFO(),
					XXSSP	: sv.XXSSP(),
					PKP	: "",
				})

		}
	}

	return	ret

}






func (s ServeZone)HSTS() string  {
	if s.StrictTransportSecurity == "" {
		s.StrictTransportSecurity = "max-age=15552000"
	}
	return s.StrictTransportSecurity
}

func (s ServeZone)XFO() string  {
	if s.XFrameOptions == "" {
		s.XFrameOptions = "SAMEORIGIN"
	}
	return s.XFrameOptions
}

func (s ServeZone)XDO() string  {
	if s.XDownloadOptions == "" {
		s.XDownloadOptions = "noopen"
	}
	return s.XDownloadOptions
}

func (s ServeZone)XCTO() string  {
	if s.XContentTypeOptions == "" {
		s.XContentTypeOptions = "nosniff"
	}
	return s.XContentTypeOptions
}

func (s ServeZone)XXSSP() string  {
	if s.XXSSProtection == "" {
		s.XXSSProtection = "1; mode=block"
	}
	return s.XXSSProtection
}

func (s ServeZone)CSP() string  {
	if s.ContentSecurityPolicy == "" {
		s.ContentSecurityPolicy = "default-src 'self'"
	}
	return s.ContentSecurityPolicy
}


func (s ServeZone)PKP(zone string) string {
	if s.TLS.IsEnabledFor(string(zone)) {
		// 2 months
		return "pin-sha256=\""+ s.TLS.PKP() +"\";max-age=5184000"
	}

	return ""
}
