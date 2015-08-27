package	sectypes

import (
	"os"
	"io/ioutil"
	"encoding/pem"
)




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
