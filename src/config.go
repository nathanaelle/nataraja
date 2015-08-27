package	main

import (
	"log"
	"path"
	"sync"
	"time"
	"strings"
	"net/url"
	"net/http"
	"io/ioutil"
	"crypto/tls"

	"./types"
	"./vhost"
	"./syslog"
	"./cache"

	"gopkg.in/fsnotify.v1"
)


type	(

	Config	struct {
		Id		string
		Listen		[]types.IpAddr
		Proxied		types.URL
		IncludeVhosts	types.Path
		Cache		*cache.Cache


		conflock	*sync.RWMutex
		refreshOCSP	time.Duration
		tls_config	*tls.Config
		syslog		*syslog.Syslog
		log		*log.Logger

		file_zones	map[string][]string
		file_vhost	map[string]*vhost.Vhost
		servable	map[string]vhost.Servable
	}

)


func NewConfig(file string, parser func(string,interface{}), sl *syslog.Syslog ) *Config {
	conf		:=new(Config)
	conf.conflock	= new(sync.RWMutex)
	conf.tls_config	= new(tls.Config)
	conf.file_zones	= make(map[string][]string)
	conf.file_vhost = make(map[string]*vhost.Vhost)
	conf.syslog	= sl
	conf.log	= sl.Channel(syslog.LOG_INFO).Logger("")
	conf.servable	= make( map[string]vhost.Servable )
	conf.refreshOCSP= 6*time.Hour

	conf.tls_config.CipherSuites = []uint16{
		tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,

		tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
		tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,

		tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
		tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
	}

	conf.tls_config.PreferServerCipherSuites	= true
	conf.tls_config.CurvePreferences		= []tls.CurveID{ tls.CurveP521, tls.CurveP384, tls.CurveP256 }

	conf.log.Printf("loading config file %s", file)
	parser( file, conf )

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

	return	conf
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

	ticker	:= time.Tick(c.refreshOCSP)
	for {
		select {
			case <-ticker:
				c.scan_OCSP(new(sync.WaitGroup))

			case <-end:
				c.log.Println("OCSPUpdater: end")
				return
		}
	}
}



func (c *Config) refresh_cert(cert *vhost.TLSConf, wg  *sync.WaitGroup)  {
	wg.Add(1)
	defer wg.Done()

	c.log.Printf("OCSPUpdater: [%s]\n", cert.CommonName())
	err	:= cert.OCSP()
	if err != nil {
		c.log.Print("OCSPUpdater: [%s] %s\n",cert.CommonName(), err.Error())
	}
	c.tls_config.Certificates = append(c.tls_config.Certificates, cert.Certificate())
}



func (c *Config) scan_OCSP(wg  *sync.WaitGroup) {
	c.conflock.Lock()
	defer c.conflock.Unlock()

	c.tls_config.Certificates = make([]tls.Certificate, 0, len(c.file_vhost))

	for _,vhost := range c.file_vhost {
		for _,cert := range vhost.ServerPairs() {
			go c.refresh_cert(cert,wg)
		}
	}

	c.tls_config.BuildNameToCertificate()

	wg.Wait()
}


func (c *Config) TLS() (*tls.Config) {
	c.conflock.Lock()
	defer c.conflock.Unlock()

	c.tls_config.Certificates = make([]tls.Certificate, 0, len(c.file_vhost))

	for _,vhost := range c.file_vhost {
		for _,cert := range vhost.ServerPairs() {
			ct := cert.Certificate()
			c.log.Printf("%s %d %t\n", ct.Leaf.Subject.CommonName, len(ct.Certificate), ct.PrivateKey != nil )
			c.tls_config.Certificates = append(c.tls_config.Certificates, ct)
		}
	}

	c.tls_config.BuildNameToCertificate()

	return c.tls_config
}


func (conf *Config) AddVhost(file string, parser func(string,interface{}), logger *log.Logger) {
	conf.conflock.Lock()
	defer conf.conflock.Unlock()

	vhost		:= vhost.New(file, parser, logger )
	servables	:= vhost.Servables()
	zones		:= make([]string,0,len(servables))

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
		if req.TLS != nil {
			datalog.TLS = true
		}

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


func (c *Config) WAF() func(*http.Request) *cache.Status {
	return func(req *http.Request) *cache.Status {

		return nil
	}
}
