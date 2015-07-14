package	vhost

import (
	"../types"
	"os"
	"io"
	"net"
	"bytes"
	"crypto"
	"strconv"
	"net/url"
	"net/http"
	"io/ioutil"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"crypto/sha256"
	"encoding/base64"
	"golang.org/x/crypto/ocsp"
)


const	MIN_STAPLE_SIZE	= 10
var issuers map[string][]byte = make(map[string][]byte,5)

type CertPair struct {
	Cert	types.Path
	Key	types.Path
	cert	*x509.Certificate
	kp	tls.Certificate
}



func (cp *CertPair)Certificate() tls.Certificate {
	return cp.kp
}


func (cp *CertPair)IsEnabled() bool {
	var err error

	stack	:= make([]x509.Certificate,0,5)

	if cp.Cert != "" && cp.Key != "" {
		//log.Println("+ "+cp.Cert.String())
		//debug.PrintStack()
		if(len(cp.kp.Certificate) ==0){
			crt		:= load_file(cp.Cert.String())
			key		:= load_file(cp.Key.String())
			cp.cert,err	= x509.ParseCertificate(load_pem(crt))
			if err != nil {
				return false
			}
			cp.kp,err	= tls.X509KeyPair(crt,key)
			if err != nil {
				return false
			}
			stack = append(stack, *cp.cert)

			for len(stack)>0 {
				cert	:= stack[0]
				stack	=  stack[1:]
				for _,issuing := range cert.IssuingCertificateURL {
					switch issuer, ok := issuers[issuing]; ok {
						case true:
							cp.kp.Certificate = append(cp.kp.Certificate, issuer)

						case false:
							issuer, err := load_issuer(issuing)
							if err == nil {
								cp.kp.Certificate = append(cp.kp.Certificate, issuer.Raw)
								issuers[issuing] = issuer.Raw
								stack = append(stack, *issuer)
							}
					}
				}

			}
			cp.OCSP()
		}
		return len(cp.cert.Raw)>0
	}
	return false
}


func (cp *CertPair)IsEnabledFor(zone string) bool {
	return	cp.IsEnabled() && cp.cert.VerifyHostname(zone) == nil
}

func (cp *CertPair)OCSP() (err error) {
	if cp.IsEnabled() && len(cp.kp.Certificate)>1 {
		for _,ocsp_server := range cp.cert.OCSPServer {
			for _,issuing := range cp.cert.IssuingCertificateURL {
				//log.Println("OCSP : ["+ocsp_server+"] ["+issuing+"]")
				issuer, err := load_issuer(issuing)
				if err != nil {
					return err
				}

				request,err := ocsp.CreateRequest(cp.cert, issuer, &ocsp.RequestOptions { crypto.SHA1 })
				if err !=nil {
					return err
				}

				staple := get_or_post_OCSP(ocsp_server,"application/ocsp-request",request)
				if len(staple) <MIN_STAPLE_SIZE {
					return nil
				}

				_,err = ocsp.ParseResponse(staple, issuer )
				//log.Printf("\n%+v\n", struct{
				//		ProducedAt, ThisUpdate, NextUpdate string
				//	}{ resp.ProducedAt.Format(time.RFC3339), resp.ThisUpdate.Format(time.RFC3339), resp.NextUpdate.Format(time.RFC3339) } )
				if err == nil {
					cp.kp.OCSPStaple = staple
					return nil
				}
			}
		}
	}
	return err
}


func load_issuer(issuing string) (*x509.Certificate, error) {
	issuer, ok := issuers[issuing]

	if !ok {
		resp, err	:= http.Get(issuing)
		defer	resp.Body.Close()
		if err!= nil {
			return nil,err
		}

		issuer,err	= ioutil.ReadAll(resp.Body)
		if err!= nil {
			return nil,err
		}
		issuers[issuing] = issuer
	}

	cert,err	:= x509.ParseCertificate(issuer)
	if err!= nil {
		return nil,err
	}

	return cert, nil
}


func (cp *CertPair)PKP() string {
	if !cp.IsEnabled() {
		return ""
	}

	der,_	:= x509.MarshalPKIXPublicKey(cp.cert.PublicKey)
	v	:= sha256.Sum256(der) //cert.RawSubjectPublicKeyInfo)
	return base64.StdEncoding.EncodeToString(v[:])
}



func needs_panic(err error) bool {
	if err == nil	{
		return false
	}

	if err == io.EOF {
		return false
	}

	switch err.(type) {
		case net.Error:
			t_err := err.(net.Error)
			return !(t_err.Timeout() || t_err.Temporary())

		case *url.Error:
			t_err := err.(*url.Error).Err.(net.Error)
			return !(t_err.Timeout() || t_err.Temporary())
	}

	return true
}






func get_or_post_OCSP(url string, mime string, data []byte) []byte {
	var	err	error
	var	rsp	*http.Response

	get_url := url + "/"+base64.URLEncoding.EncodeToString(data)

	if len(get_url)<255 {
		rsp,err	= http.Get(get_url)

		if err!= nil {
			if needs_panic(err) {
				panic(err)
			}
			return []byte{}
		}
		//log.Printf("\n-G----\n%s\n%+v %+v\n----\n\n", get_url, rsp.Status, rsp.Header)
	}

	need_post	:= false
	switch {
		case len(get_url)>=255:
			need_post	= true

		case err != nil:
			need_post	= true

		default:
			v,_ := strconv.Atoi(rsp.Header.Get("Content-Length"))
			need_post	= v <= MIN_STAPLE_SIZE
	}

	if need_post {
		rsp,err	= http.Post(url, mime, bytes.NewReader(data))
		//log.Printf("\n-P----\n%+v %+v\n----\n\n", rsp.Status, rsp.Header)
		if err!= nil {
			if needs_panic(err) {
				panic(err)
			}
			return []byte{}
		}
	}
	defer rsp.Body.Close()

	body,err:= ioutil.ReadAll(rsp.Body)
	if err!= nil {
		if needs_panic(err) {
			panic(err)
		}
		return []byte{}
	}
	return body
}





func load_file(file string) []byte {
	f,_	:= os.Open(file)
	buf,_	:= ioutil.ReadAll(f)
	f.Close()
	return buf
}



func load_pem(file []byte) []byte {
	b,_	:= pem.Decode(file)
	return b.Bytes
}
