package	types

import (
	"bytes"
	"errors"
	"unicode"
)


/*
 *	type:		FQDN
 *	content:	full qualified domain name
 *	pitfall:	don't verify if the fqdn exists
 *			don't verify if the tld exists
 *			don't handle conversion from unicode to punycode
 */
//	fqdn validation
//	(?=^.{4,253}$)(^((?!-)[a-zA-Z0-9-]{1,63}(?<!-)\.)+[a-zA-Z]{2,63}\.?$)

type	FQDN	string

func (d *FQDN) Set(t_fqdn string) (err error) {
	if len(t_fqdn)< 1 || len(t_fqdn)>253 {
		err	= errors.New("invalid FQDN : "+t_fqdn )
		return
	}

	dot	:= false
	for _,char := range t_fqdn {
		switch	{
			case char == '\\' :
				dot	= false

			case char == '-' :
				if (dot) {
					return	errors.New("hyphen after is forbidden for FQDN ["+t_fqdn+"]")
				}
				dot	= false

			case char == '.' :
				if (dot) {
					return	errors.New("double dot is forbidden for FQDN ["+t_fqdn+"]")
				}
				dot	= true

			case unicode.IsNumber(char) || unicode.IsLetter(char):
				dot	= false

			default:
				return	errors.New("invalid char ["+string(char)+"] for FQDN ["+t_fqdn+"]")
		}
	}

	*d = FQDN(t_fqdn)

	return nil
}


func (d *FQDN) String() string {
	return string(*d)
}



func (d *FQDN) UnmarshalTOML(data []byte) (err error) {
	return d.Set(string(bytes.Trim(data,"\"")))
}



/*
 *	Explode a FQDN to a slice of strings
 */
func (d *FQDN) Split() []string {
	res	:= make([]string, 1)
	fqdn	:= string(*d)
	begin	:= 0

	quote	:= false
	for pos,char := range fqdn {
		switch char {
			case '\\':
				quote = !quote

			case '.':
				if !quote {
					if len(fqdn[begin:pos]) > 0 {
						res	= append(res, fqdn[begin:pos] )
					}
					begin	= pos+1
				}

			default:
				quote	= false
		}
	}

	if begin < len(fqdn) {
		res	= append(res, fqdn[begin:len(fqdn)] )
	}

	return res
}


/*
 *	Explode a FQDN to a slice of strings
 */
func (d *FQDN) PathToRoot() []string {
	res	:= make([]string, 1)
	fqdn	:= string(*d)
	end	:= len(fqdn)
	last	:= 0
	res[0]	= fqdn

	quote	:= false
	for pos,char := range fqdn {
		switch char {
			case '\\':
				quote = !quote

			case '.':
				if !quote {
					res	= append(res, fqdn[pos:end] )
					last	= pos
				}
				quote = false

			default:
				quote = false
		}
	}

	if last < end-1 {
		res	= append(res, "." )
	}

	return res
}


func (d *FQDN) ToPunny() (*FQDN) {
	return d
}



func (d *FQDN) FromPunny() (*FQDN) {
	return d
}
