package	vhost

import (
	"crypto"
	"crypto/tls"
	sectype	"../security.types"
)

type	TLSConf struct {
	Cert	*sectype.Cert
	Keys	[]*sectype.Key

	prv_k	crypto.PrivateKey

	invalid	bool
	valid	bool
}


func (cp *TLSConf)Certificate() tls.Certificate {
	cert		:= cp.Cert.Certificate()
	cert.PrivateKey	= cp.prv_k
	return cert
}


func (cp TLSConf)CommonName() string {
	return cp.Cert.CommonName()
}


func (cp *TLSConf)IsEnabled() bool {
	if cp.invalid {
		return false
	}

	if cp.valid {
		return true
	}

	if cp.Cert == nil || !cp.Cert.IsEnabled() || len(cp.Keys) == 0 {
		cp.invalid = true
		return false
	}

	cert := cp.Cert.Certificate().Leaf
	for _,k := range cp.Keys {
		if !k.IsEnabled() {
			continue
		}

		pk := k.InCertificate(cert)
		if pk != nil {
			cp.prv_k = pk
		}
	}

	if cp.prv_k == nil {
		cp.invalid = true
		return false
	}

	cp.valid = true
	return true
}

func (cp *TLSConf)IsEnabledFor(zone string) bool {
	return	cp.IsEnabled() && cp.Cert.IsEnabledFor(zone)
}

func (cp *TLSConf)OCSP() error {
	if !cp.IsEnabled() {
		return nil
	}

	return cp.Cert.RefreshOCSP()
}

func (cp *TLSConf)PKP() []string {
	pkp	:= make([]string,0,len(cp.Keys))

	for _,k := range cp.Keys {
		if !k.IsEnabled() {
			continue
		}
		p := k.PKP()
		if p != ""{
			pkp = append(pkp, p)
		}
	}
	return pkp
}
