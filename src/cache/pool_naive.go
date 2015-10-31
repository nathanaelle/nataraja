package	cache

import	(
	"net"
	"sync"
	"time"
	"net/http"
	"sync/atomic"
)


type	(

	NaiveMemory	struct {
		sync.Mutex
		cache		atomic.Value
		transport	http.RoundTripper
	}

)


const	exp_not_cachable	int64	= 3600*4
const	dlc			int64	= 3600*12


func (volatile_cache *NaiveMemory) Init() error {
	volatile_cache.transport = &http.Transport{
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

	volatile_cache.cache.Store(make(map[string]CacheCont,1000))

	return nil
}


func (volatile_cache *NaiveMemory) round_trip(req *http.Request) (*Entry,error){
	res, err := volatile_cache.transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	return NewEntry(res.Header, res.StatusCode, res.ContentLength, res.Body ), nil
}



func (volatile_cache *NaiveMemory) cache_round_trip(key string, req *http.Request) (*Entry,error){
	res, err := volatile_cache.transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	ent := NewEntry(res.Header, res.StatusCode, res.ContentLength, res.Body )

	if !ent.Cachable() {
		volatile_cache.write( key, CacheCont { time.Now().Unix()+exp_not_cachable, nil } )
		return ent, nil
	}

	switch res.StatusCode {
	case	200,204,205,404,410,400:
		return volatile_cache.write( key, CacheCont { time.Now().Unix()+ent.CacheControl.MaxAge, ent } )

	case	304:
		cc,_	:= volatile_cache.read(key)
		ent	= cc.Entity
		return volatile_cache.write( key, CacheCont { time.Now().Unix()+ent.CacheControl.MaxAge, ent } )

	default:
		volatile_cache.write( key, CacheCont { time.Now().Unix()+exp_not_cachable, nil } )
		return ent, nil
	}
}


func (volatile_cache *NaiveMemory) read(key string) (CacheCont,bool)  {
	m	:= volatile_cache.cache.Load().(map[string]CacheCont)
	ent, ok	:= m[key]

	return ent,ok
}


func (volatile_cache *NaiveMemory) write(key string, cc CacheCont) (*Entry,error) {
	volatile_cache.Lock()
	defer volatile_cache.Unlock()

	now	:= time.Now().Unix()
	o_m	:= volatile_cache.cache.Load().(map[string]CacheCont)
	n_m	:= make(map[string]CacheCont,len(o_m))

	for k, v := range o_m {
		if (now - v.Expire) < dlc {
			n_m[k] = v
		}
	}
	n_m[key] = cc
	volatile_cache.cache.Store(n_m)

	return cc.Entity,nil
}


func (volatile_cache *NaiveMemory) Get(req *http.Request) (*Entry,error) {

	if req.Method != "GET" && req.Method != "HEAD" {
		return volatile_cache.round_trip( req )
	}

	my_url		:= req.URL
	my_url.RawQuery	= my_url.Query().Encode()
	cleaned_url	:= my_url.String()

	ent, ok	:= volatile_cache.read(cleaned_url)

	if ok {
		if ent.Entity == nil {
			return volatile_cache.round_trip( req )
		}

		if time.Now().Unix() < ent.Expire {
			return ent.Entity, nil
		}
	}

	return volatile_cache.cache_round_trip( cleaned_url, req )
}
