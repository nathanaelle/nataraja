package cache


import	(
	"io"
	"fmt"
	"time"
	"bytes"
	"errors"
	"strings"
	"strconv"
	"net/http"
)


type	(

	ReadSeekCloser interface {
		io.Reader
		io.Closer
		io.Seeker
	}

	CacheCont	struct {
		Expire		int64
		Purge		int64
		Entity		*Entry
	}


	Entry	struct {
		Status		int
		ContentType	string
		LastModified	time.Time
		Etag		string
		CacheControl	CacheControl
		Header		http.Header
		BodyLen		int64
		body		ReadSeekCloser
		BodyBytes	[]byte
	}


	Buffer struct {
		*bytes.Reader
	}


	MimicReadSeekCloser struct {
		io.ReadCloser
	}

)


func (*Buffer) Close() error  {
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
		Etag:		header.Get("ETag"),
		LastModified:	httpDate2Time( header.Get("Last-Modified"), now ),
		CacheControl:	NewCacheControl(header.Get("Cache-Control")),
		Status:		status,
	}

	header.Del("Pragma")
	header.Del("Content-Length")
	header.Del("Cache-Control")

	// Drop Expires and inject if needed as Cache-Control
	if expires, ok := header["Expires"] ; ok {
		delete(header, "Expires")
		exp	:= int64(httpDate2Time( expires[0], now ).Sub(now)/time.Second)
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

	if header.Get("Set-Cookie") != "" {
		entry.CacheControl.Private = true
		entry.CacheControl.Public = false
		entry.CacheControl.NoCache = true
	}

	if header.Get("Vary") != "" {
		entry.CacheControl.Private = true
		entry.CacheControl.Public = false
		entry.CacheControl.NoCache = true
	}


	header.Set("Cache-Control", entry.CacheControl.String() )
	entry.Header	= header


	switch status {
		// user can custom these codes
		case 200,201,202,203,226:

		// user can custom these codes
		case 401,403,404:

		default:
			body.Close()
			entry.BodyLen = 0
			entry.BodyBytes = []byte{}

			return entry
	}



	switch ContentLength {
		case 0:
			body.Close()
			entry.BodyLen	= 0
			entry.BodyBytes = []byte{}
			return entry

		case -1:
			buff	:= new(bytes.Buffer)
			blen,err:= io.Copy(buff, body)
			entry.BodyBytes = buff.Bytes()
			entry.BodyLen	= blen
			body.Close()
			if err != nil {
				entry.BodyLen	= 0
				entry.BodyBytes = []byte{}
				entry.Status	= http.StatusServiceUnavailable
				return entry
			}
			return entry
	}
	entry.BodyLen	= ContentLength
	if !entry.Cachable() {
		switch rsc,ok	:= body.(ReadSeekCloser); ok {
			case true:	entry.body = rsc
			case false:	entry.body = &MimicReadSeekCloser{ body }
		}

		return entry
	}

	buff	:= bytes.NewBuffer(make([]byte, 0, ContentLength))
	blen,err:= io.CopyN(buff, body, ContentLength)
	entry.BodyBytes = buff.Bytes()
	body.Close()
	if err != nil || blen != ContentLength {
		entry.BodyLen	= 0
		entry.BodyBytes = []byte{}
		entry.Status	= http.StatusServiceUnavailable
	}

	return entry
}


func (e Entry) Body() ReadSeekCloser {
	if e.body != nil {
		return e.body
	}

	return &Buffer{ bytes.NewReader( e.BodyBytes ) }
}

func (e Entry) Close() error {
	if e.body != nil {
		return e.body.Close()
	}
	return nil
}

func (e Entry) Cachable() bool {
	return e.CacheControl.Cachable()
}







type	CacheControl	struct {
	MustRevalidate	bool		`msgpack:",omitempty"`
	NoCache		bool		`msgpack:",omitempty"`
	NoStore		bool		`msgpack:",omitempty"`
	NoTransform	bool		`msgpack:",omitempty"`
	Public		bool		`msgpack:",omitempty"`
	Private		bool		`msgpack:",omitempty"`
	ProxyRevalidate	bool		`msgpack:",omitempty"`
	MaxAge		int64		`msgpack:",omitempty"`
	SMaxAge		int64		`msgpack:",omitempty"`
	Unknown		[]string	`msgpack:",omitempty"`
}

func (cc CacheControl) Cachable() bool {
	return cc.Public && !cc.Private && !cc.NoCache && !cc.NoStore && !cc.MustRevalidate && cc.MaxAge > 0
}


func (cc CacheControl) String() string {
	list	:= make([]string,0,9+len(cc.Unknown))

	if cc.Private {
		list = append(list, "private")
	}

	if cc.Public {
		list = append(list, "public")
	}

	if cc.NoCache {
		list = append(list, "no-cache")
	}

	if cc.NoStore {
		list = append(list, "no-store")
	}

	if cc.MustRevalidate {
		list = append(list, "must-revalidate")
	}

	if cc.ProxyRevalidate {
		list = append(list, "proxy-revalidate")
	}

	if cc.MaxAge >= 0 {
		list = append(list, fmt.Sprintf("max-age=%d",cc.MaxAge))
	}

	if cc.SMaxAge > 0 {
		list = append(list, fmt.Sprintf("s-maxage=%d",cc.SMaxAge))
	}

	if len(cc.Unknown) >0 {
		list = append(list, cc.Unknown...)
	}

	return strings.Join(list,",")
}


func NewCacheControl(rawcc string) CacheControl {
	cc := CacheControl {
		MustRevalidate:		false,
		NoCache:		false,
		NoStore:		false,
		NoTransform:		false,
		Public:			false,
		Private:		true,
		ProxyRevalidate:	false,
		MaxAge:			0,
		SMaxAge:		0,
		Unknown:		[]string{},
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
					if err != nil {
						continue
					}
					cc.MaxAge = int64(ma)
				}
				if len(token)>9 && token[0:8] == "s-maxage" {
					sma,err := strconv.Atoi(token[9:])
					if err != nil {
						continue
					}
					cc.SMaxAge = int64(sma)
				}
				cc.Unknown = append( cc.Unknown, token )
		}
	}

	return cc
}
