package	sectypes

import (
	"../types"
	"crypto"
	"crypto/rsa"
	"crypto/ecdsa"
	"crypto/x509"
	"crypto/sha256"
	"encoding/base64"
)

type	(
	keytype	byte

	Key	struct {
		file	types.Path
		prv_k	crypto.PrivateKey
		typ_k	keytype
		pkp	string
		invalid	bool
	}
)

const	(
	T_RSA_PRV	keytype	= iota
	T_RSA_PUB
	T_ECDSA_PRV
	T_ECDSA_PUB
)




func (kt *Key) UnmarshalTOML(data []byte) error  {
	return (&kt.file).UnmarshalTOML(data)
}



func (kt *Key) IsEnabled() bool {
	if kt.file == "" || kt.invalid {
		return false
	}

	if len(kt.pkp) > 0{
		return true
	}

	var err error

	p,err		:= file2pem(kt.file.String())
	if err != nil {
		kt.invalid = true
		return false
	}

	switch  p.Type {
		case "PRIVATE KEY":
			prv_key,err := x509.ParsePKCS8PrivateKey(p.Bytes)
			if err != nil {
				kt.invalid = true
				return false
			}

			switch key := prv_key.(type) {
				case *rsa.PrivateKey:
					kt.prv_k	= key
					kt.pkp		= computePKP(&key.PublicKey)
					kt.typ_k	= T_RSA_PRV

				case *ecdsa.PrivateKey:
					kt.prv_k	= key
					kt.pkp		= computePKP(&key.PublicKey)
					kt.typ_k	= T_ECDSA_PRV

				default:
					kt.invalid = true
					return false
			}

		case "PUBLIC KEY":
			pub_key,err := x509.ParsePKIXPublicKey(p.Bytes)
			if err != nil {
				kt.invalid = true
				return false
			}

			switch pub_key.(type) {
				case *rsa.PublicKey:
					kt.pkp		= signPKP(p.Bytes)
					kt.typ_k	= T_RSA_PUB

				case *ecdsa.PublicKey:
					kt.pkp		= signPKP(p.Bytes)
					kt.typ_k	= T_ECDSA_PUB

				default:
					kt.invalid = true
					return false
			}

		case "RSA PRIVATE KEY":
			rsa_key,err := x509.ParsePKCS1PrivateKey(p.Bytes)
			if err != nil {
				kt.invalid = true
				return false
			}

			kt.prv_k	= rsa_key
			kt.pkp		= computePKP(&rsa_key.PublicKey)
			kt.typ_k	= T_RSA_PRV

		case "EC PRIVATE KEY":
			ec_key,err := x509.ParseECPrivateKey(p.Bytes)
			if err != nil {
				kt.invalid = true
				return false
			}

			kt.prv_k	= ec_key
			kt.pkp		= computePKP(&ec_key.PublicKey)
			kt.typ_k	= T_ECDSA_PRV

		default:
			kt.invalid = true
			return false
	}

	return true
}



func (kt *Key) InCertificate(cert *x509.Certificate) crypto.PrivateKey {
	if kt.invalid {
		return nil
	}

	switch kt.typ_k {
		case T_RSA_PUB,T_ECDSA_PUB:
			return nil

		case T_RSA_PRV:
			return certificate_use_this_rsa_key(cert, kt.prv_k.(*rsa.PrivateKey))

		case T_ECDSA_PRV:
			return certificate_use_this_ecdsa_key(cert, kt.prv_k.(*ecdsa.PrivateKey))

	}
	panic("unreachable")
}

func certificate_use_this_rsa_key( cert *x509.Certificate, rsa_key *rsa.PrivateKey) crypto.PrivateKey {
	if pub,ok := cert.PublicKey.(*rsa.PublicKey); ok {
		if pub.N.Cmp(rsa_key.PublicKey.N) == 0 && pub.E == rsa_key.PublicKey.E {
			return rsa_key
		}
	}
	return nil
}

func certificate_use_this_ecdsa_key( cert *x509.Certificate, ecdsa_key *ecdsa.PrivateKey) crypto.PrivateKey {
	if pub,ok := cert.PublicKey.(*ecdsa.PublicKey); ok {
		if pub.X.Cmp(ecdsa_key.PublicKey.X) == 0 && pub.Y.Cmp(ecdsa_key.PublicKey.Y) == 0 {
			return ecdsa_key
		}
	}
	return nil
}


func (kt *Key) PKP() string {
	return kt.pkp
}


func computePKP(pk interface{}) string {
	pk_der,err:= x509.MarshalPKIXPublicKey(pk)
	if err != nil {
		return ""
	}
	return signPKP(pk_der)
}

func signPKP(pk_der []byte) string {
	hash	:= sha256.Sum256(pk_der)
	return base64.StdEncoding.EncodeToString(hash[:])
}
