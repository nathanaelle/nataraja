package	vhost

import (
	"../types"
//	"log"
//	"net/http"
	"strings"
)


type	ServeZone	struct {
	Zones			[]types.FQDN
	Proxied			types.URL
	Cert			types.Path
	Keys			[]types.Path
	TLS			*TLSConf

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
					Proxied	: sv.Proxied,
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
					Proxied	: sv.Proxied,
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
		ret := make([]string,0,5)
		for _,pkp := range s.TLS.PKP() {
			ret = append(ret,"pin-sha256=\""+pkp+"\"")
		}

		// 2 months
		ret = append( ret, "max-age=5184000")
		return strings.Join(ret,";")
	}

	return ""
}
