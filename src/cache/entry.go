package cache


import	(
	"net/http"
	"strings"
	"strconv"
	"time"
	"bytes"
	"io"
	"errors"
)


type	(

	ReadSeekCloser interface {
		io.Reader
		io.Closer
		io.Seeker
	}



	Entry	struct {
		Status		int
		ContentType	string
		LastModified	time.Time
		Etag		string
		CacheControl	CacheControl
		Header		http.Header
		BodyLen		int64
		Body		ReadSeekCloser
	}


	Buffer struct {
		*bytes.Reader
	}


	MimicReadSeekCloser struct {
		io.ReadCloser
	}

)


func (*Buffer)Close() error  {
	return nil
}


func (s *MimicReadSeekCloser) Seek(offset int64, whence int) (int64, error) {
	if whence != 0 {
		return 0,errors.New("Mimic can only SEEK_SET")
	}
	t_off := offset
	for t_off > 0 {
		t := make([]byte,t_off)
		size, err := s.Read(t)
		t_off = t_off - int64(size)
		if err == io.EOF {
			return t_off,nil
		}
		if err != nil {
			return t_off,err
		}
	}

	return offset,nil
}






func NewEntry(header http.Header,status int,ContentLength int64, body io.ReadCloser) *Entry {
	now	:= time.Now()

	entry	:= &Entry{
		ContentType:	header.Get("Content-Type"),
		Etag:		header.Get("Etag"),
		LastModified:	httpDate2Time( header.Get("Last-Modified"), now ),
		CacheControl:	NewCacheControl(header.Get("Cache-Control")),
		Status:		status,
	}

	delete(header, "Content-Length")
	entry.Header	= header

	// Drop Expires and inject if needed as Cache-Control
	if expires, ok := header["Expires"] ; ok {
		delete(header, "Expires")
		exp	:= int(httpDate2Time( expires[0], now ).Sub(now)/time.Second)
		switch {
			// invalid expires has no effect so nothing need to be done
			case exp == 0:

			// found an expires
			case exp > 0:
				if entry.CacheControl.MaxAge <= 0 {
					entry.CacheControl.MaxAge = exp
				}

			default:
				if entry.CacheControl.MaxAge <= 0 && !entry.CacheControl.NoCache {
					entry.CacheControl.NoStore = true
				}
		}
	}


	switch status {
		// user can custom these codes
		case 200,201,202,204,205, 203,226:

		// user can custom these codes
		case 401,403,404:

		default:
			body.Close()
			entry.BodyLen = 0
			entry.Body = nil

			return entry
	}



	switch ContentLength {
		case 0:
			body.Close()
			entry.BodyLen	= 0
			entry.Body	= nil
			return entry

		case -1:
			buff	:= new(bytes.Buffer)
			blen,err:= io.Copy(buff, body)
			entry.Body	= &Buffer{ bytes.NewReader( buff.Bytes() ) }
			entry.BodyLen	= blen
			body.Close()
			if err != nil {
				entry.Body.Close()
				entry.BodyLen	= 0
				entry.Body	= nil
				entry.Status	= http.StatusServiceUnavailable
				return entry
			}
			return entry

	}

	entry.BodyLen	= ContentLength
	switch rsc,ok	:= body.(ReadSeekCloser); ok {
		case true:	entry.Body = rsc
		case false:	entry.Body = &MimicReadSeekCloser{ body }
	}

	return entry
}



func (e Entry)Close() {
	if e.Body != nil {
		e.Body.Close()
	}
}








type	CacheControl	struct {
	MustRevalidate	bool
	NoCache		bool
	NoStore		bool
	NoTransform	bool
	Public		bool
	Private		bool
	ProxyRevalidate	bool
	MaxAge		int
	SMaxAge		int
	Unknown		[]string
	Raw		string
}




func NewCacheControl(rawcc string) CacheControl {
	cc := CacheControl {
		MustRevalidate:		false,
		NoCache:		false,
		NoStore:		false,
		NoTransform:		false,
		Public:			true,
		Private:		false,
		ProxyRevalidate:	false,
		MaxAge:			0,
		SMaxAge:		0,
		Unknown:		[]string{},
		Raw:			rawcc,
	}

	s_cc := strings.Split(rawcc, ",")

	for _, token := range s_cc {
		switch token {
			case "no-cache":		cc.NoCache		= true
			case "no-store":		cc.NoStore		= true
			case "must-revalidate":		cc.MustRevalidate	= true
			case "proxy-revalidate":	cc.ProxyRevalidate	= true

			case "public":
				cc.Public	= true
				cc.Private	= false
			case "private":
				cc.Public	= false
				cc.Private	= true

			default:
				if len(token)>8 && token[0:7] == "max-age" {
					ma,err := strconv.Atoi(token[8:])
					if err == nil {
						continue
					}
					cc.MaxAge = ma
				}
				if len(token)>9 && token[0:8] == "s-maxage" {
					sma,err := strconv.Atoi(token[9:])
					if err == nil {
						continue
					}
					cc.SMaxAge = sma
				}
				cc.Unknown = append( cc.Unknown, token )
		}
	}

	return cc
}
