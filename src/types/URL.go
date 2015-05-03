package	types

import (
	"bytes"
	"net/url"
)




type	URL	url.URL

func (d *URL) UnmarshalTOML(data []byte) error  {
	dest, err := url.Parse(string(bytes.Trim(data,"\"")))
	if err == nil {
		*d = URL(*dest)
	}
	return err
}
