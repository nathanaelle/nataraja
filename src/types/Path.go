package	types

import (
	"bytes"
	"os"
	"fmt"
)




type	Path	string

func (d *Path) UnmarshalTOML(data []byte) error  {
	return d.Set(string(bytes.Trim(data,"\"")))
}


func (d *Path) Set(data string) (err error) {
	_, err = os.Stat(data)
	if err == nil {
		*d = Path(data)
	}
	return
}


func (d *Path) String() string {
	return string(*d)
}




type	PathList	[]Path


func (d *PathList) Set(data string) (err error) {
	_, err = os.Stat(data)
	if err == nil {
		*d = append( *d, Path(data) )
	}
	return
}


func (d *PathList) String() string {
	return fmt.Sprintf("%s", *d)
}
