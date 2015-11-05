// +build ignore

package	cache

import	(
	"net"
	"sync"
	"time"
	"net/http"

	"github.com/boltdb/bolt"
	"gopkg.in/vmihailenco/msgpack.v2"
)


type	(

	BoltCache	struct {
		DBpath		string
		transport	http.RoundTripper
	}
)



func (perm_cache *BoltCache) Init() error {
	perm_cache.transport = &http.Transport{
	        Proxy: nil,
	        Dial: (&net.Dialer{
			DualStack:	true,
	                Timeout:	5*time.Minute,
	                KeepAlive:	10*time.Minute,
	        }).Dial,
	        TLSHandshakeTimeout:	2*time.Second,
		DisableCompression:	true,
		MaxIdleConnsPerHost:	100,
	}

	return nil
}


func (perm_cache *BoltCache) round_trip(req *http.Request) (*Entry,error){
	res, err := perm_cache.transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	return NewEntry(res.Header, res.StatusCode, res.ContentLength, res.Body ), nil
}



func (perm_cache *BoltCache) cache_round_trip(key string, req *http.Request) (*Entry,error){
	res, err := perm_cache.transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	ent := NewEntry(res.Header, res.StatusCode, res.ContentLength, res.Body )

	if !ent.Cachable() {
		perm_cache.cache[key] = CacheCont { 0, nil }
		return ent,nil
	}

	switch res.StatusCode {
	case	200:
		perm_cache.cache[key] = CacheCont { time.Now().Unix()+ent.CacheControl.MaxAge, ent }

	case	304:
		ent = perm_cache.cache[key].Entity
		perm_cache.cache[key] = CacheCont { time.Now().Unix()+ent.CacheControl.MaxAge, ent }
	}

	return ent,nil
}


func (perm_cache *BoltCache) Get(req *http.Request) (*Entry,error) {
	lock	:= new(sync.Mutex)

	if req.Method != "GET" && req.Method != "HEAD" {
		return perm_cache.round_trip( req )
	}

	my_url		:= req.URL
	my_url.RawQuery	= my_url.Query().Encode()
	cleaned_url	:= my_url.String()

	lock.Lock()
	ent, ok	:= perm_cache.cache[cleaned_url]

	if ok {
		if ent.Entity == nil {
			lock.Unlock()
			return perm_cache.round_trip( req )
		}

		if time.Now().Unix() < ent.Expire {
			lock.Unlock()
			return ent.Entity, nil
		}
	}

	defer lock.Unlock()
	return perm_cache.cache_round_trip( cleaned_url, req )
}
