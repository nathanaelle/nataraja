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

//	"gopkg.in/fsnotify.v1"

	"./types"
	"./vhost"
	"./syslog"
	rp	"./reverseproxy"

)


type	Config	struct {
	Id		string
	Listen		[]types.IpAddr
	Proxied		types.URL
	IncludeVhosts	types.Path


	refreshOCSP	time.Duration
	tls_config	*tls.Config
	serverpairs	[]*vhost.CertPair
	syslog		*syslog.Syslog
	log		*log.Logger

	file_zones	map[string][]string
	servable	map[string]vhost.Servable
}


func NewConfig(file string, parser func(string,interface{}), sl *syslog.Syslog ) *Config {
	conf		:= new( Config )
	conf.tls_config	 = new(tls.Config)
	conf.file_zones	 = make(map[string][]string)
	conf.serverpairs = make( []*vhost.CertPair, 0, 1 )
	conf.syslog	 = sl
	conf.log	 = sl.Channel(syslog.LOG_INFO).Logger("")
	conf.servable	 = make( map[string]vhost.Servable )
	conf.refreshOCSP = 1*time.Hour

	conf.tls_config.CipherSuites = []uint16{
	//	tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
	//	tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,

		tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
		tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,

		tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
		tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
	}

	conf.tls_config.PreferServerCipherSuites	= true

	conf.log.Printf("loading config file %s", file)
	parser( file, conf )

	root_dir	:= string(conf.IncludeVhosts)
	files,err	:= ioutil.ReadDir(root_dir)
	exterminate(err)

	for _,file := range files {
		switch {
			case file.IsDir():
				is_vhost	:= false
				vhost_dir,err	:= ioutil.ReadDir(path.Join(root_dir,file.Name()))
				exterminate(err)
				for _,vhost_file := range vhost_dir {
					if vhost_file.Mode().IsRegular() && vhost_file.Name() == "config.vhost" {
						is_vhost = true
					}
				}

				if is_vhost {
					conf.AddVhost(path.Join(root_dir,file.Name(),"config.vhost"), parser, sl.SubSyslog("vhost").Channel(syslog.LOG_INFO).Msgid(file.Name()).Logger("") )
				}

			case file.Mode().IsRegular():
				if strings.HasSuffix(file.Name(), ".vhost") {
					conf.AddVhost(path.Join(root_dir,file.Name()), parser, sl.SubSyslog("vhost").Channel(syslog.LOG_INFO).Msgid(file.Name()).Logger("") )
				}
		}
	}

	return	conf
}


func (c *Config) ConfUpdater(write_conf sync.Locker,end <-chan bool,wg  *sync.WaitGroup) {
	wg.Add(1)
	for {
		select {
			case <-end:
				c.log.Println("ConfUpdater: end")
				return
		}
	}
	defer wg.Done()
}


func (c *Config) OCSPUpdater(write_conf sync.Locker,end <-chan bool,wg  *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	ticker	:= time.Tick(c.refreshOCSP)
	for {
		select {
			case <-ticker:
				c.scan_OCSP(write_conf)

			case <-end:
				c.log.Println("OCSPUpdater: end")
				return
		}
	}
}


func (c *Config) scan_OCSP(write_conf sync.Locker) {
	write_conf.Lock()
	defer write_conf.Unlock()

	for _,cert := range c.serverpairs {
		cert.OCSP()
	}
}






func (c *Config) TLS() (*tls.Config) {
	if len(c.serverpairs)==0 {
		c.log.Println("No TLS conf detected")
		return nil
	}

	c.tls_config.Certificates = make([]tls.Certificate, 0, len(c.serverpairs))

	for _,v := range c.serverpairs {
		if v != nil && v.IsEnabled() {
			c.tls_config.Certificates = append(c.tls_config.Certificates, v.Certificate())
		}
	}

	c.tls_config.BuildNameToCertificate()
	return c.tls_config
}







func (conf *Config) AddVhost(file string, parser func(string,interface{}), logger *log.Logger) {
	v		:= vhost.New(file, parser, logger )
	servables	:= v.Servables()
	zones		:= make([]string,0,len(servables))
	conf.serverpairs = append(conf.serverpairs, v.ServerPairs()... )

	for zone,desc := range v.Servables() {
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
}


func (c *Config) found_servable(matches []string) (servable vhost.Servable, ok bool) {
	for _,possible_match := range matches {
		servable, ok = c.servable[possible_match]
		if ok {
			return
		}
	}
	return vhost.Servable {}, false
}

func (c *Config) ToProxy() url.URL {
	return url.URL(c.Proxied)
}


func (c *Config) Routing(read_conf sync.Locker) func(*http.Request,*rp.Datalog) (*rp.Status,http.Header) {
	return func(req *http.Request, datalog *rp.Datalog) (*rp.Status, http.Header) {
		read_conf.Lock()
		defer read_conf.Unlock()

		header := make(http.Header)

		if req.Host == "" {
			return &rp.Status { http.StatusBadRequest, "No [Host:]" }, header
		}

		if req.TLS != nil {
			datalog.TLS = true
			if req.TLS.ServerName == "" {
				return &rp.Status { http.StatusBadRequest, "no tls servername" }, header
			}

			if req.TLS.ServerName != req.Host {
				return &rp.Status { http.StatusBadRequest, "tls server name mismatch [Host:]" }, header
			}
		}

		d	:= new(types.FQDN)
		if d.Set(req.Host) != nil {
			return &rp.Status { http.StatusBadRequest, "invalid [Host:]" }, header
		}

		servable, ok	:= c.found_servable( d.PathToRoot() )

		if !ok {
			return &rp.Status { http.StatusBadRequest, "unknown [Host:]" }, header
		}

		datalog.Owner	= servable.Owner
		datalog.Project	= servable.Project
		datalog.Vhost	= servable.Zone

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

		if servable.Redirect != "" {
			t := *(req.URL)
			switch req.TLS {
				case	nil:	t.Scheme="http"
				default:	t.Scheme="https"
			}
			t.Host	= servable.Redirect
			return &rp.Status { http.StatusMovedPermanently,  t.String() }, header
		}

		if req.TLS == nil && servable.TLS {
			t := *(req.URL)
			t.Scheme= "https"
			t.Host	= req.Host
			return &rp.Status { http.StatusMovedPermanently,  t.String() }, header
		}

		return nil,header
	}
}


func (c *Config) WAF() func(*http.Request) *rp.Status {
	return func(req *http.Request) *rp.Status {

		return nil
	}
}
