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
	"crypto/rsa"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"crypto/sha256"
	"encoding/base64"
	"golang.org/x/crypto/ocsp"
)


const	MIN_STAPLE_SIZE	= 10
var	issuers map[string]*x509.Certificate = make(map[string]*x509.Certificate,100)

type	TLSConf struct {
	Cert	types.Path
	Keys	[]types.Path
	cert	*x509.Certificate
	kp	tls.Certificate

	hpkp	[]string
}




func (cp *TLSConf)Certificate() tls.Certificate {
	return cp.kp
}


func (cp *TLSConf)IsEnabled() bool {
	if cp.Cert == "" || len(cp.Keys) == 0 {
		return false
	}

	if len(cp.kp.Certificate) > 0{
		return true
	}

	var err error
	keys	:= []crypto.PrivateKey {}

	crt,err		:= file2pem(cp.Cert.String())
	if err != nil {
		return false
	}

	cp.kp.Certificate	= append( cp.kp.Certificate, crt.Bytes)
	cp.cert,err		= x509.ParseCertificate(crt.Bytes)
	if err != nil {
		return false
	}

	for _,k := range cp.Keys {
		var	key	crypto.PrivateKey
		var	pk_der	[]byte

		p,err		:= file2pem(k.String())
		if err != nil {
			continue
		}

		switch  p.Type {
			case "PRIVATE KEY":
				key,err := x509.ParsePKCS8PrivateKey(p.Bytes)
				if err != nil {
					continue
				}
				switch key := key.(type) {
					case *rsa.PrivateKey:
						rsa_key	:= key
						pk_der,_= x509.MarshalPKIXPublicKey(&rsa_key.PublicKey)
						if pub,ok := cp.cert.PublicKey.(*rsa.PublicKey); ok {
							if pub.N.Cmp(rsa_key.PublicKey.N) == 0 && pub.E == rsa_key.PublicKey.E {
								cp.kp.PrivateKey = key
							}
						}


					case *ecdsa.PrivateKey:
						ec_key		:= key
						pk_der,_	= x509.MarshalPKIXPublicKey(&ec_key.PublicKey)
						if pub,ok := cp.cert.PublicKey.(*ecdsa.PublicKey); ok {
							if pub.X.Cmp(ec_key.PublicKey.X) == 0 && pub.Y.Cmp(ec_key.PublicKey.Y) == 0 {
								cp.kp.PrivateKey = key
							}
						}

					default:
						continue
				}


			case "RSA PRIVATE KEY":
				rsa_key,err := x509.ParsePKCS1PrivateKey(p.Bytes)
				if err != nil {
					continue
				}
				key	= rsa_key
				pk_der,_= x509.MarshalPKIXPublicKey(&rsa_key.PublicKey)
				if pub,ok := cp.cert.PublicKey.(*rsa.PublicKey); ok {
					if pub.N.Cmp(rsa_key.PublicKey.N) == 0 && pub.E == rsa_key.PublicKey.E {
						cp.kp.PrivateKey = key
					}
				}


			case "EC PRIVATE KEY":
				ec_key,err := x509.ParseECPrivateKey(p.Bytes)
				if err != nil {
					continue
				}
				key	= ec_key
				pk_der,_= x509.MarshalPKIXPublicKey(&ec_key.PublicKey)
				if pub,ok := cp.cert.PublicKey.(*ecdsa.PublicKey); ok {
					if pub.X.Cmp(ec_key.PublicKey.X) == 0 && pub.Y.Cmp(ec_key.PublicKey.Y) == 0 {
						cp.kp.PrivateKey = key
					}
				}

			default:
				continue
		}
		hash	:= sha256.New()
		hash.Write(pk_der)
		cp.hpkp = append(cp.hpkp, base64.StdEncoding.EncodeToString(hash.Sum(nil)) )
		keys	= append(keys, key)

	}


	stack	:= make([]x509.Certificate,0,5)
	stack = append(stack, *cp.cert)
	for len(stack)>0 {
		cert	:= stack[0]
		stack	=  stack[1:]
		for _,issuing := range cert.IssuingCertificateURL {
			switch issuer, ok := issuers[issuing]; ok {
				case true:
					cp.kp.Certificate = append(cp.kp.Certificate, issuer.Raw)
					stack = append(stack, *issuer)

				case false:
					issuer, err := load_issuer(issuing)
					if err == nil {
						cp.kp.Certificate = append(cp.kp.Certificate, issuer.Raw)
						issuers[issuing] = issuer
						stack = append(stack, *issuer)
					}
			}
		}
	}
	cp.OCSP()

	return len(cp.cert.Raw)>0
}


func (cp *TLSConf)IsEnabledFor(zone string) bool {
	return	cp.IsEnabled() && cp.cert.VerifyHostname(zone) == nil
}

func (cp *TLSConf)OCSP() (err error) {
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
	resp, err	:= http.Get(issuing)
	if err!= nil {
		return nil,err
	}
	defer	resp.Body.Close()

	issuer,err	:= ioutil.ReadAll(resp.Body)
	if err!= nil {
		return nil,err
	}

	cert,err	:= x509.ParseCertificate(issuer)
	if err!= nil {
		return nil,err
	}

	return cert, nil
}


func (cp *TLSConf)PKP() []string {
	return cp.hpkp
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
		defer rsp.Body.Close()
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
		defer rsp.Body.Close()
	}

	body,err:= ioutil.ReadAll(rsp.Body)
	if err!= nil {
		if needs_panic(err) {
			panic(err)
		}
		return []byte{}
	}
	return body
}





func file2pem(file string) (*pem.Block,error) {
	f,err	:= os.Open(file)
	if err != nil {
		return nil,err
	}

	buf,err	:= ioutil.ReadAll(f)
	f.Close()
	if err != nil {
		return nil,err
	}

	b,_	:= pem.Decode(buf)
	return b,nil
}
