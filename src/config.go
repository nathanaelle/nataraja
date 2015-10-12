package	main

import (
	"log"
	"path"
	"sync"
	"time"
	"errors"
	"strings"
	"net/url"
	"net/http"
	"io/ioutil"
	"crypto/tls"


	"./vhost"
	"./cache"

	"gopkg.in/fsnotify.v1"

	syslog	"github.com/nathanaelle/syslog5424"
	types	"github.com/nathanaelle/useful.types"
)


type	(

	Config	struct {
		Id		string
		Listen		[]types.IpAddr
		Proxied		types.URL
		IncludeVhosts	types.Path
		Cache		*cache.Cache
		RefreshOCSP	types.Duration

		conflock	*sync.RWMutex
		syslog		*syslog.Syslog
		log		*log.Logger

		tlspairs	map[string]*vhost.TLSConf

		file_zones	map[string][]string
		file_vhost	map[string]*vhost.Vhost
		servable	map[string]vhost.Servable
	}

)


func NewConfig(file string, parser func(string,interface{}), sl *syslog.Syslog ) *Config {
	conf := &Config{
		//RefreshOCSP:				types.Duration(12*time.Hour),
		RefreshOCSP:				types.Duration(1*time.Hour),

		conflock:				new(sync.RWMutex),
		tlspairs:				make(map[string]*vhost.TLSConf),
		file_zones:				make(map[string][]string),
		file_vhost:				make(map[string]*vhost.Vhost),
		syslog:					sl,
		log:					sl.Channel(syslog.LOG_INFO).Logger(""),
		servable:				make( map[string]vhost.Servable ),
	}

	conf.log.Printf("loading config file %s", file)
	parser( file, conf )

	switch {
		case conf.RefreshOCSP < types.Duration(5*time.Minute):	conf.RefreshOCSP= types.Duration(5*time.Minute)
		case conf.RefreshOCSP > types.Duration(24*time.Hour):	conf.RefreshOCSP= types.Duration(24*time.Hour)
	}

	root_dir	:= string(conf.IncludeVhosts)
	files,err	:= ioutil.ReadDir(root_dir)
	exterminate(err)

	for _,file := range files {
		if !file.Mode().IsRegular() {
			continue
		}
		if strings.HasSuffix(file.Name(), ".vhost") {
			conf.AddVhost(path.Join(root_dir,file.Name()), parser, sl.SubSyslog("vhost").Channel(syslog.LOG_INFO).Msgid(file.Name()).Logger("") )
		}
	}

	conf.scan_OCSP( new(sync.WaitGroup), new(sync.Mutex) )

	return	conf
}


func (conf *Config) GetCertificate(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	conf.conflock.RLock()
	defer conf.conflock.RUnlock()

	sni := strings.TrimRight(strings.ToLower(clientHello.ServerName),".")
	if cert, ok := conf.tlspairs[sni] ; ok {
		return cert.Certificate(), nil
	}

	labels := strings.Split(sni, ".")
	for i := range labels {
		labels[i] = "*"
		sni := strings.Join(labels, ".")
		if cert, ok := conf.tlspairs[sni]; ok {
			return cert.Certificate(), nil
		}
	}

	return nil, errors.New("No Certificate for :"+sni)
}


func (c *Config) ConfUpdater(end <-chan bool,wg  *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	defer watcher.Close()

	err = watcher.Add( c.IncludeVhosts.String() )
	if err != nil {
		panic(err)
	}

	for {
		select {
			case event := <-watcher.Events:
				if event.Op&fsnotify.Write == fsnotify.Write {
					// here broadcast reload conf
				}

			case err := <-watcher.Errors:
				log.Println("error:", err)

			case <-end:
				c.log.Println("ConfUpdater: end")
				return

		}
	}
}


func (c *Config) OCSPUpdater(end <-chan bool,wg  *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	ticker	:= time.Tick(c.RefreshOCSP.Get().(time.Duration))
	for {
		select {
			case <-ticker:
				c.scan_OCSP(new(sync.WaitGroup),new(sync.Mutex))

			case <-end:
				c.log.Println("OCSPUpdater: end")
				return
		}
	}
}


func (c *Config) refresh_cert(cert *vhost.TLSConf, wg *sync.WaitGroup, lock *sync.Mutex) {
	wg.Add(1)
	defer wg.Done()

	cert.OCSP()

	lock.Lock()
	defer lock.Unlock()
	for _,sni := range cert.DNSNames() {
		c.tlspairs[sni] = cert
	}
}


func (c *Config) scan_OCSP(wg *sync.WaitGroup, lock *sync.Mutex) {
	c.conflock.Lock()
	defer c.conflock.Unlock()

	c.tlspairs = make(map[string]*vhost.TLSConf, len(c.tlspairs))
	for _,vhost := range c.file_vhost {
		for _,cert := range vhost.ServerPairs() {
			go c.refresh_cert(cert, wg, lock)
		}
	}

	wg.Wait()
}


func (conf *Config) AddVhost(file string, parser func(string,interface{}), logger *log.Logger) {
	conf.conflock.Lock()
	defer conf.conflock.Unlock()

	vhost		:= vhost.New(file, parser, logger )
	servables	:= vhost.Servables()
	zones		:= make([]string, 0, len(servables))

	for zone,desc := range servables {
		zones	= append(zones, zone)
		already_servable, ok := conf.servable[zone]
		switch ok {
			case false:
				conf.servable[zone] = desc

			case true:
				log.Panic("Already Servable : "+ zone + " for "+ already_servable.Owner)
		}
	}
	conf.file_zones[file]	= zones
	conf.file_vhost[file]	= vhost
}


func (c *Config) SearchServable(matches []string) (servable vhost.Servable, ok bool) {
	c.conflock.RLock()
	defer c.conflock.RUnlock()

	for _,possible_match := range matches {
		servable, ok = c.servable[possible_match]
		if ok {
			return
		}
	}
	return vhost.Servable {}, false
}


func (c *Config) Configure() func(*http.Request,*cache.Datalog) (http.Header,url.URL) {
	return func(req *http.Request, datalog *cache.Datalog) (http.Header,url.URL) {
		d := new(types.FQDN)
		d.Set(req.Host)
		servable,_	:= c.SearchServable( d.PathToRoot() )

		datalog.Owner	= servable.Owner
		datalog.Project	= servable.Project
		datalog.Vhost	= servable.Zone

		header := make(http.Header)
		header.Set("X-Frame-Options"		, servable.XFO	)
		header.Set("X-Content-Type-Options"	, servable.XCTO	)
		header.Set("X-Download-Options"		, servable.XDO	)
		header.Set("X-XSS-Protection"		, servable.XXSSP)
		//header.Set("Content-Security-Policy",servable.CSP)

		if req.TLS != nil {
			header.Set("Strict-Transport-Security",servable.HSTS)
			if servable.PKP != "" {
				header.Set("Public-Key-Pins",servable.PKP)
			}
		}

		default_proxy	:= url.URL(c.Proxied)
		candidat_proxy	:= url.URL(servable.Proxied)
		if candidat_proxy.Host == "" {
			return header, default_proxy
		}
		return header, candidat_proxy
	}
}
