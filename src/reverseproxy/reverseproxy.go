// minimal rewrite cut 'n paste form net/http/httputil/reverseproxy.go
// need a full rewrite later (don't like the mutex here)
package reverseproxy


import (
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)





// ReverseProxy is an HTTP Handler that takes an incoming request and
// sends it to another server, proxying the response back to the
// client.
type	ReverseProxy	struct {
	Target url.URL

	// The transport used to perform proxy requests.
	// If nil, http.DefaultTransport is used.
	Transport http.RoundTripper

	// FlushInterval specifies the flush interval
	// to flush to the client while copying the
	// response body.
	// If zero, no periodic flushing is done.
	FlushInterval time.Duration

	// ErrorLog specifies an optional logger for errors
	// that occur when attempting to proxy the request.
	// If nil, logging goes to os.Stderr via the log package's
	// standard logger.
	ErrorLog	*log.Logger
	AccessLog	*log.Logger

	Prefilter	func(*http.Request,*Datalog)	(*Status,http.Header)
	WAF		func(*http.Request)		*Status

}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
		case aslash && bslash:
			return a + b[1:]
		case !aslash && !bslash:
			return a + "/" + b
	}
	return a + b
}


func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

// Hop-by-hop headers. These are removed when sent to the backend.
// http://www.w3.org/Protocols/rfc2616/rfc2616-sec13.html
var hopHeaders = []string{
	"Connection",
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"Te", // canonicalized version of "TE"
	"Trailers",
	"Transfer-Encoding",
	"Upgrade",
}

type requestCanceler interface {
	CancelRequest(*http.Request)
}




// Remove hop-by-hop headers to the backend.  Especially
// important is "Connection" because we want a persistent
// connection, regardless of what the client sent to us.  This
// is modifying the same underlying map from req (shallow
// copied above) so we only copy it if necessary.
func copy_cleaned_headers(dst,src *http.Request) {
	copiedHeaders := false
	for _, h := range hopHeaders {
		if dst.Header.Get(h) != "" {
			if !copiedHeaders {
				dst.Header = make(http.Header)
				copyHeader(dst.Header, src.Header)
				copiedHeaders = true
			}
			dst.Header.Del(h)
		}
	}
}


func configure_outgoing_request(req	*http.Request, target url.URL) *http.Request {
	outreq := new(http.Request)
	*outreq = *req // includes shallow copies of maps, but okay

	outreq.Proto		= "HTTP/1.1"
	outreq.ProtoMajor	= 1
	outreq.ProtoMinor	= 1
	outreq.Close		= false
	outreq.URL.Scheme	= target.Scheme
	outreq.URL.Host		= target.Host
	outreq.URL.Path		= singleJoiningSlash(target.Path, outreq.URL.Path)

	if target.RawQuery == "" || outreq.URL.RawQuery == "" {
		outreq.URL.RawQuery = target.RawQuery + outreq.URL.RawQuery
	} else {
		outreq.URL.RawQuery = target.RawQuery + "&" + outreq.URL.RawQuery
	}

	copy_cleaned_headers(outreq, req)

	if clientIP, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
		// If we aren't the first proxy retain prior
		// X-Forwarded-For information as a comma+space
		// separated list and fold multiple headers into one.
		if prior, ok := outreq.Header["X-Forwarded-For"]; ok {
			clientIP = strings.Join(prior, ", ") + ", " + clientIP
		}
		outreq.Header.Set("X-Forwarded-For", clientIP)
	}

	return outreq
}


func (p *ReverseProxy) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	datalog	:= &Datalog {
		Owner		: "-",
		Project		: "-",
		Vhost		: "-",
		Host		: req.Host,
		TLS		: false,
		Proto		: req.Proto,
		Method		: req.Method,
		Request		: req.URL.String(),
		RemoteAddr	: req.RemoteAddr,
		Referer		: req.Referer(),
		UserAgent	: req.UserAgent(),
		ContentType	: "-",
	}

	defer http_log(p.AccessLog, time.Now(), datalog )

	ret_status, injectHeaders := p.Prefilter( req, datalog )
	if ret_status != nil {
		ret_status.PrematureExit(rw, datalog )
		return
	}

	ret_status = p.WAF( req )
	if ret_status != nil {
		ret_status.PrematureExit(rw, datalog )
		return
	}

	transport := p.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}

	outreq := configure_outgoing_request(req, p.Target)

	if closeNotifier, ok := rw.(http.CloseNotifier); ok {
		if requestCanceler, ok := transport.(requestCanceler); ok {
			reqDone := make(chan struct{})
			defer close(reqDone)

			clientGone := closeNotifier.CloseNotify()

			outreq.Body = struct {
				io.Reader
				io.Closer
			}{
				Reader: &runOnFirstRead{
					Reader: outreq.Body,
					fn: func() {
						go func() {
							select {
							case <-clientGone:
								requestCanceler.CancelRequest(outreq)
							case <-reqDone:
							}
						}()
					},
				},
				Closer: outreq.Body,
			}
		}
	}


	res, err := transport.RoundTrip(outreq)
	if err != nil {
		p.ErrorLog.Printf("http: proxy error: %v", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer res.Body.Close()

	for _, h := range hopHeaders {
		res.Header.Del(h)
	}
	res.Header.Del("Server")
	res.Header.Del("X-Powered-By")


	copyHeader(rw.Header(), res.Header)
	copyHeader(rw.Header(), injectHeaders)

	datalog.ContentType = rw.Header().Get("Content-Type")

	datalog.Status = res.StatusCode
	rw.WriteHeader(res.StatusCode)
	p.copyResponse(rw, res.Body)
}


func (p *ReverseProxy) copyResponse(dst io.Writer, src io.Reader) (size int64) {
	if p.FlushInterval != 0 {
		if wf, ok := dst.(writeFlusher); ok {
			mlw := &maxLatencyWriter{
				dst:     wf,
				latency: p.FlushInterval,
				done:    make(chan bool),
			}
			go mlw.flushLoop()
			defer mlw.stop()
			dst = mlw
		}
	}

	size,_ = io.Copy(dst, src)
	return
}
