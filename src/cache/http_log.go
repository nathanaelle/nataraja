package	cache


import (
	"log"
	"fmt"
	"time"
	"net/http"
	"crypto/tls"
	"encoding/json"
)


type	Datalog	struct {
	Start		int64
	Duration	int64
	Status		int
	Owner		string
	Project		string
	Vhost		string
	Host		string
	TLS		bool
	TLSCipher	string
	TLSVer		string
	Proto		string
	Method		string
	Request		string
	RemoteAddr	string
	Referer		string
	UserAgent	string
	ContentType	string
	BodySize	int64
}

func NewLog(req *http.Request) *Datalog {
	dlog := &Datalog {
		Owner		: "-",
		Project		: "-",
		Vhost		: "-",
		TLSVer		: "-",
		TLS		: false,
		Host		: req.Host,
		Proto		: req.Proto,
		Method		: req.Method,
		Request		: req.URL.String(),
		RemoteAddr	: req.RemoteAddr,
		Referer		: req.Referer(),
		UserAgent	: req.UserAgent(),
		ContentType	: "-",
	}

	if req.TLS != nil {
		dlog.TLS	= true

		switch req.TLS.CipherSuite {
			case tls.TLS_RSA_WITH_RC4_128_SHA:			dlog.TLSCipher = "TLS_RSA_WITH_RC4_128_SHA"
		        case tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA:			dlog.TLSCipher = "TLS_RSA_WITH_3DES_EDE_CBC_SHA"
		        case tls.TLS_RSA_WITH_AES_128_CBC_SHA:			dlog.TLSCipher = "TLS_RSA_WITH_AES_128_CBC_SHA"
		        case tls.TLS_RSA_WITH_AES_256_CBC_SHA:			dlog.TLSCipher = "TLS_RSA_WITH_AES_256_CBC_SHA"
		        case tls.TLS_ECDHE_ECDSA_WITH_RC4_128_SHA:		dlog.TLSCipher = "TLS_ECDHE_ECDSA_WITH_RC4_128_SHA"
		        case tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA:		dlog.TLSCipher = "TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA"
		        case tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA:		dlog.TLSCipher = "TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA"
		        case tls.TLS_ECDHE_RSA_WITH_RC4_128_SHA:		dlog.TLSCipher = "TLS_ECDHE_RSA_WITH_RC4_128_SHA"
		        case tls.TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA:		dlog.TLSCipher = "TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA"
		        case tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA:		dlog.TLSCipher = "TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA"
		        case tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA:		dlog.TLSCipher = "TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA"
		        case tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256:		dlog.TLSCipher = "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"
		        case tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256:	dlog.TLSCipher = "TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256"
		        case tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384:		dlog.TLSCipher = "TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384"
		        case tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384:	dlog.TLSCipher = "TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384"
			default:						dlog.TLSCipher	= fmt.Sprintf("0x%x",req.TLS.CipherSuite)
		}

		switch req.TLS.Version {
			case tls.VersionSSL30:	dlog.TLSVer = "SSL 3"
			case tls.VersionTLS10:	dlog.TLSVer = "TLS 1.0"
			case tls.VersionTLS11:	dlog.TLSVer = "TLS 1.1"
			case tls.VersionTLS12:	dlog.TLSVer = "TLS 1.2"
			default:		dlog.TLSVer = fmt.Sprintf("0x%x",req.TLS.Version)
		}
	}

	return dlog
}


func (d Datalog)String() string  {
	raw,_	:= json.Marshal(d)
	return string(raw)
}

func LogHTTP(accesslog *log.Logger, start time.Time, datalog *Datalog)  {
	if datalog == nil {
		return
	}
	if datalog.Status < 0 {
		return
	}

	datalog.Start	= start.Unix()
	datalog.Duration= int64(time.Since(start)/time.Microsecond)
	accesslog.Printf("%s", datalog.String() )
}
