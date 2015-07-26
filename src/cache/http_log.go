package	cache


import (
	"log"
	"time"
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
	Proto		string
	Method		string
	Request		string
	RemoteAddr	string
	Referer		string
	UserAgent	string
	ContentType	string
	BodySize	int64
}


func (d *Datalog)String() string  {
	raw,_	:= json.Marshal(*d)
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
