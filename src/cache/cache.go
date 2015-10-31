package cache


import	(
	"io"
	"os"
	"fmt"
	"log"
	"net"
	"time"
	"net/url"
	"strings"
	"strconv"
	"net/http"
)


type	Cache	struct {
	// ErrorLog specifies an optional logger for errors
	// that occur when attempting to proxy the request.
	ErrorLog	*log.Logger

	// AccessLog specifies an optional logger for errors
	// that occur when attempting to proxy the request.
	// If nil, nothing is logged
	AccessLog	*log.Logger

	Configure	func(*http.Request,*Datalog)	(http.Header,url.URL)

	pool		Pool
}


func (cache *Cache) Init(pool Pool) {
	cache.pool	= pool
	cache.pool.Init()
}


func (cache *Cache) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	datalog	:= NewLog(req)

	defer LogHTTP(cache.AccessLog, time.Now(), datalog )

	injectHeaders, ProxyTarget := cache.Configure( req, datalog )

	outreq	:= configure_outgoing_request(req, ProxyTarget)
	entry,err := cache.pool.Get( outreq )
	defer entry.Close()

	if err != nil {
		cache.ErrorLog.Printf("http: proxy error: %v", err)
		InternalServerError("").PrematureExit(rw, datalog )
		return
	}

	if entry.Status >=500 {
		InternalServerError("").PrematureExit(rw, datalog )
		return
	}

	serveContent(rw, datalog, req.Header, injectHeaders, entry, req.Method == "HEAD")
}





func serveContent(rw http.ResponseWriter,datalog *Datalog, source_header http.Header, injectHeaders http.Header, data *Entry, nobody bool) {

	// Cache & Range & If-Range needed here
	if s_ims, ok := source_header["If-Modified-Since"]; ok {
		ims := httpDate2Time( s_ims[0], time.Unix(0,0) )
		if !ims.Before(data.LastModified) {
			NotModified().PrematureExit(rw, datalog )
			return
		}
	}

	if inm, ok := source_header["If-None-Match"]; ok {
		if inm[0] == data.Etag {
			NotModified().PrematureExit(rw, datalog )
			return
		}
	}

	body	:= data.Body()

	ranges := [][2]int64 {}
	if s_r, ok := source_header["Range"]; ok {
		ok,ranges = compute_ranges(s_r[0],data.BodyLen)
		if !ok {
			RangeNotSatisfiable("*/"+strconv.FormatInt(int64(data.BodyLen),10)).PrematureExit(rw, datalog )
			return
		}

		if len(ranges) == 1 {
			if _, err := body.Seek( ranges[0][0], os.SEEK_SET); err != nil {
				RangeNotSatisfiable("*/"+strconv.FormatInt(int64(data.BodyLen),10)).PrematureExit(rw, datalog )
				return
			}

			Header_in2out(rw.Header(), data.Header, injectHeaders )
			datalog.ContentType	= data.ContentType
			datalog.Status		= http.StatusPartialContent
			size			:=ranges[0][1]-ranges[0][0]
			datalog.BodySize	= size
			rw.Header().Set("Content-Length", strconv.FormatInt(size,10))
			rw.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d",ranges[0][0], ranges[0][1],data.BodyLen))

			rw.WriteHeader(http.StatusPartialContent)
			if nobody {
				return
			}

			io.CopyN(rw, body, size)
			return
		}

		// TODO : need to handle multipart Range
		RangeNotSatisfiable("*/"+strconv.FormatInt(int64(data.BodyLen),10)).PrematureExit(rw, datalog )
		return
	}

	Header_in2out(rw.Header(), data.Header, injectHeaders )
	datalog.Status		= data.Status
	datalog.ContentType	= data.ContentType
	datalog.BodySize	= data.BodyLen
	if data.BodyLen > 100*1024 {
		rw.Header().Set("Accept-Ranges", "bytes")
	}

	rw.Header().Set("Content-Length", strconv.FormatInt(data.BodyLen,10))
	rw.WriteHeader(data.Status)
	if nobody {
		return
	}

	io.CopyN(rw, body,data.BodyLen)
	return
}




func compute_ranges(s_r string,data_len int64) (ok bool,ranges [][2]int64) {
	sum_range := int64(0)

	if !strings.HasPrefix(s_r, "bytes=") {
		return false,[][2]int64 {}
	}

	s_ranges := strings.Split(s_r[6:], ",")
	for _,s_range := range s_ranges {
		start_end := strings.Split(s_range, "-")
		if len(start_end) != 2 {
			return false,[][2]int64 {}
		}

		switch {
			case start_end[0] == "":
				if start_end[1]=="" {
					return false,[][2]int64 {}
				}
				r_end, err := strconv.ParseInt(start_end[1], 10, 64 )
				if err != nil {
					return false,[][2]int64 {}
				}
				if r_end > data_len {
					return false,[][2]int64 {}
				}
				ranges = append( ranges, [2]int64{ data_len-r_end, data_len })
				sum_range = sum_range+r_end


			case start_end[1] == "":
				r_start, err := strconv.ParseInt(start_end[0], 10, 64 )
				if err != nil {
					return false,[][2]int64 {}
				}
				if r_start > data_len {
					return false,[][2]int64 {}
				}
				ranges = append( ranges, [2]int64{ r_start, data_len })
				sum_range = sum_range+data_len-r_start


			default:
				r_start, err := strconv.ParseInt(start_end[0], 10, 64 )
				if err != nil {
					return false,[][2]int64 {}
				}

				r_end, err := strconv.ParseInt(start_end[1], 10, 64 )
				if err != nil {
					return false,[][2]int64 {}
				}

				if r_start > data_len || r_start >= r_end || r_end > data_len {
					return false,[][2]int64 {}
				}
				ranges = append( ranges, [2]int64{ r_start, r_end })
				sum_range = sum_range+r_end-r_start
		}
	}

	if sum_range > data_len {
		return false,[][2]int64 {}
	}

	if sum_range == 0 {
		return false,[][2]int64 {}
	}

	return	true,ranges
}


func Header_out2in(src http.Header) http.Header {
	dst := make(http.Header,len(src))

	for k, vv := range src {
		switch	k {
		//
		case "Connection", "Keep-Alive", "Proxy-Authenticate", "Proxy-Authorization", "Te", "Trailers", "Transfer-Encoding", "Upgrade":

		// Nataraja's job
		case "Range", "Cache-Control","Accept-Encoding":

		// really old and cargo culted header
		case "Pragma":

		default:
			for _, v := range vv {
				dst.Add(k, v)
			}
		}
	}
	return dst
}


func Header_in2out(dst http.Header, headers ...http.Header) {
	size := 0
	for _,h := range headers {
		size = size + len(h)
	}

	for _,h := range headers {
		for k, vv := range h {
			switch	k {
			// Some Security by Obscurity
			case	"Server", "X-Powered-By":

			// really old and cargo culted header
			case "Pragma":

			// Nataraja's job - enforce the last only
			case	"Accept-Ranges", "Public-Key-Pins", "StrictTransportSecurity","X-XSS-Protection":
				for _, v := range vv {
					dst.Set(k, v)
				}

			// enforce only if not already set
			case	"X-Frame-Options","X-Content-Type-Options","X-Download-Options","Content-Security-Policy":
				if _,ok := dst[k]; !ok {
					for _, v := range vv {
						dst.Set(k, v)
					}
				}

			default:
				for _, v := range vv {
					dst.Add(k, v)
				}
			}
		}
	}
}


func configure_outgoing_request(req	*http.Request, target url.URL) *http.Request {
	outreq	:= new(http.Request)
	*outreq	= *req

	outreq.Proto		= "HTTP/1.1"
	outreq.ProtoMajor	= 1
	outreq.ProtoMinor	= 1
	outreq.Close		= false
	outreq.Header		= Header_out2in(req.Header)

	outreq.URL.Scheme	= target.Scheme
	outreq.URL.Host		= target.Host

	aslash		:= strings.HasSuffix(target.Path, "/")
	bslash		:= strings.HasPrefix(outreq.URL.Path, "/")
	switch {
		case  aslash&&  bslash:	outreq.URL.Path = target.Path + outreq.URL.Path[1:]
		case !aslash&& !bslash:	outreq.URL.Path = target.Path + "/" + outreq.URL.Path
		default:		outreq.URL.Path = target.Path + outreq.URL.Path
	}

	switch target.RawQuery == "" || outreq.URL.RawQuery == "" {
		case	true:	outreq.URL.RawQuery = target.RawQuery + outreq.URL.RawQuery
		case	false:	outreq.URL.RawQuery = target.RawQuery + "&" + outreq.URL.RawQuery
	}

	// If we aren't the first proxy retain prior
	// X-Forwarded-For information as a comma+space
	// separated list and fold multiple headers into one.
	if clientIP, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
		if prior, ok := outreq.Header["X-Forwarded-For"]; ok {
			clientIP = strings.Join(prior, ", ") + ", " + clientIP
		}
		outreq.Header.Set("X-Forwarded-For", clientIP)
	}

	return outreq
}
