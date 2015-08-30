package	vhost

import (
	sectype	"../security.types"
	"log"
	"strings"

	types "github.com/nathanaelle/useful.types"
)


type	ServeZone	struct {
	Zones			[]types.FQDN
	Proxied			types.URL
	Cert			*sectype.Cert
	Keys			[]*sectype.Key

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
	if sv.TLS != nil {
		log.Println("Serve.TLS is deprecated, please upgrade your conf")
	}

	ret	:= make([]Servable, 0, len(sv.Zones))
	tls	:= sv.TLSConf()

	for _,zone := range sv.Zones {
		ret = append(ret, sv.get_Servable(tls, Owner,Project, zone.String()))
	}

	return	ret

}


func (sv ServeZone)get_Servable(tls *TLSConf, Owner,Project,zone string) Servable {
	switch tls != nil && tls.IsEnabledFor(zone) {
		case true:
			return Servable {
				Owner	: Owner,
				Project	: Project,
				Proxied	: sv.Proxied,
				Zone	: zone,
				TLS	: true,
				HSTS	: sv.HSTS(tls, zone),
				XCTO	: sv.XCTO(),
				XDO	: sv.XDO(),
				XFO	: sv.XFO(),
				XXSSP	: sv.XXSSP(),
				PKP	: sv.PKP(tls, zone),
			}


		default:
			return Servable {
				Owner	: Owner,
				Project	: Project,
				Proxied	: sv.Proxied,
				Zone	: zone,
				TLS	: false,
				HSTS	: "",
				XCTO	: sv.XCTO(),
				XDO	: sv.XDO(),
				XFO	: sv.XFO(),
				XXSSP	: sv.XXSSP(),
				PKP	: "",
			}
	}

}



func (sv ServeZone) TLSConf() *TLSConf {
	tls	:= &TLSConf {
		Cert:	sv.Cert,
		Keys:	sv.Keys,
	}

	if tls.IsEnabled() {
		return tls
	}

	return nil
}




func (s ServeZone)HSTS(tls *TLSConf, zone string) string  {
	if !tls.IsEnabledFor(zone) {
		return ""
	}

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


func (s ServeZone)PKP(tls *TLSConf, zone string) string {
	if !tls.IsEnabledFor(zone) {
		return ""
	}

	ret := make([]string,0,5)
	for _,pkp := range tls.PKP() {
		ret = append(ret,"pin-sha256=\""+pkp+"\"")
	}

	// 2 months
	ret = append( ret, "max-age=5184000")
	return strings.Join(ret,";")
}
