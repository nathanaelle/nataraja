package	main

import (
	"log"
	"flag"
	"path"
	"sync"
	"time"
	"errors"
	"runtime"
	"strings"
	"net/url"
	"net/http"
	"io/ioutil"
	"crypto/tls"

	"strconv"
	"crypto/sha256"
	"crypto/hmac"


	"./vhost"
	"./cache"

	"github.com/go-fsnotify/fsnotify"
	"golang.org/x/net/http2"

	syslog	"github.com/nathanaelle/syslog5424"
	types	"github.com/nathanaelle/useful.types"
)


const	APP_NAME	string		= "nataraja"
const	DEFAULT_CONF	types.Path	= "/etc/nataraja/config.toml"
const	DEFAULT_PRIO	syslog.Priority	= (syslog.LOG_DAEMON|syslog.LOG_WARNING)
const	DEFAULT_DEVLOG	types.Path	= ""

const	derive_list_size	int	= 12
const	ticket_lifetime		int64	= 1800

type Nataraja struct {
	Id			string
	Listen			[]types.IpAddr
	Proxied			types.URL
	DevLog			types.Path
	IncludeVhosts		types.Path
	RefreshOCSP		types.Duration
	TicketMasterSecret	string

	wg			*sync.WaitGroup
	conflock		*sync.RWMutex
	end			<-chan bool
	ticket_id		int64

	syslog			*syslog.Syslog
	log			*log.Logger

	server			*http.Server
	cache			*cache.Cache

	tlspairs		map[string]*vhost.TLSConf
	file_zones		map[string][]string
	file_vhost		map[string]*vhost.Vhost
	servable		map[string]vhost.Servable
}


func SummonNataraja() (*Nataraja) {
	nat := &Nataraja {
		RefreshOCSP:		types.Duration(1*time.Hour),
		TicketMasterSecret:	rand_b64_string(),

		wg:			new(sync.WaitGroup),
		conflock:		new(sync.RWMutex),

		tlspairs:		make(map[string]*vhost.TLSConf),
		file_zones:		make(map[string][]string),
		file_vhost:		make(map[string]*vhost.Vhost),
		servable:		make(map[string]vhost.Servable ),
	}

	return nat
}

func (nat *Nataraja) ReadConfiguration()  {
	conf_path	:= new(types.Path)
	priority	:= new(syslog.Priority)

	*priority	 = DEFAULT_PRIO
	*conf_path	 = DEFAULT_CONF

	var	numcpu	= flag.Int("cpu", 1, "maximum number of logical CPU that can be executed simultaneously")
	var	stderr	= flag.Bool("stderr", false, "send message to stderr instead of syslog")

	flag.Var(conf_path	, "conf", "path to the conf" )
	flag.Var(priority	, "priority", "log priority in syslog format facility.severity" )

	flag.Parse()

	parse_toml_file( conf_path.String(), nat )

	switch {
		case *numcpu >runtime.NumCPU():	runtime.GOMAXPROCS(runtime.NumCPU())
		case *numcpu <1:		runtime.GOMAXPROCS(1)
		default:			runtime.GOMAXPROCS(*numcpu)
	}

	switch {
		case nat.RefreshOCSP < types.Duration(5*time.Minute):	nat.RefreshOCSP= types.Duration(5*time.Minute)
		case nat.RefreshOCSP > types.Duration(24*time.Hour):	nat.RefreshOCSP= types.Duration(24*time.Hour)
	}

	switch *stderr {
	case true:
		co,err	:= (syslog.Dialer{
			QueueLen:	100,
			FlushDelay:	100*time.Millisecond,
		}).Dial( "stdio", "stderr", new(syslog.T_LFENDED) )

		if err != nil {
			log.Fatal(err)
		}
		nat.syslog,_ =	syslog.New( co, *priority, APP_NAME )

	case false:
		co,err	:= (syslog.Dialer{
			QueueLen:	100,
			FlushDelay:	100*time.Millisecond,
		}).Dial( "local", nat.DevLog.String(), new(syslog.T_ZEROENDED) )

		if err != nil {
			log.Fatal(err)
		}
		nat.syslog,_ =	syslog.New( co, *priority, APP_NAME )

	}

	nat.log = nat.syslog.Channel(syslog.LOG_INFO).Logger("")


	if nat.Id != "" {
		nat.syslog	= nat.syslog.SubSyslog(nat.Id)
	}
}

func (nat *Nataraja) LoadVirtualHosts() {
	root_dir	:= string(nat.IncludeVhosts)
	files,err	:= ioutil.ReadDir(root_dir)
	exterminate(err)

	for _,file := range files {
		if !file.Mode().IsRegular() {
			continue
		}
		filename := file.Name()
		if strings.HasSuffix(filename, ".vhost") {
			nat.AddVhost(path.Join(root_dir,filename), parse_toml_file, nat.syslog.SubSyslog("vhost").Channel(syslog.LOG_INFO).Msgid(filename).Logger("") )
		}
	}

	nat.scan_OCSP( new(sync.WaitGroup), new(sync.Mutex) )
}


func (nat *Nataraja) AddVhost(file string, parser func(string,interface{}), logger *log.Logger) {
	nat.conflock.Lock()
	defer nat.conflock.Unlock()

	vhost		:= vhost.New(file, parser, logger )
	servables	:= vhost.Servables()
	zones		:= make([]string, 0, len(servables))

	for zone,desc := range servables {
		zones	= append(zones, zone)
		already_servable, ok := nat.servable[zone]
		switch ok {
			case false:
				nat.servable[zone] = desc

			case true:
				log.Panic("Already Servable : "+ zone + " for "+ already_servable.Owner)
		}
	}
	nat.file_zones[file]	= zones
	nat.file_vhost[file]	= vhost
}


func (nat *Nataraja) GenerateServer() {
	nat.cache	= &cache.Cache {
		AccessLog	: nat.syslog.Channel(syslog.LOG_NOTICE).Logger(""),
		Configure	: nat.Configure_pre_route_request(),
		ErrorLog	: nat.syslog.SubSyslog("proxy").Channel(syslog.LOG_WARNING).Logger("WARNING: "),
	}

	nat.cache.Init(&cache.NaiveMemory {})

	nat.server = &http.Server {
		Handler:			nat,
		ReadTimeout:			10 * time.Minute,
		WriteTimeout:			10 * time.Minute,
//		ConnState:			sessionLogger(),
		ErrorLog:			nat.syslog.SubSyslog("connexion").Channel(syslog.LOG_INFO).Logger("INFO: "),
		TLSConfig:			&tls.Config{
			GetCertificate:			nat.GetCertificate,
			PreferServerCipherSuites:	true,
			CurvePreferences:		[]tls.CurveID{
								tls.CurveP521,
								tls.CurveP384,
								tls.CurveP256,
							},
			CipherSuites:			[]uint16{
								tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
								tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
								tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
								tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,

								tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
								tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
								tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
								tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
							},
						},
	}

	http2.ConfigureServer( nat.server, &http2.Server {} )

	nat.derive_ticket()
}


func (nat *Nataraja) SignalHandler() {
	nat.end, _	= SignalCatcher()

	go nat.ConfUpdater(nat.end,nat.wg)
	go nat.OCSPUpdater(nat.end,nat.wg)
	go nat.TicketUpdater(nat.end,nat.wg)
}


func (nat *Nataraja) Run() {
	crit	:= nat.syslog.Channel(syslog.LOG_CRIT)
	for _,ip:= range nat.Listen {
		go serveSock(nat.end, nat.wg, crit.Msgid("http" ).Logger("Fatal: "), nat.server, forgeHTTP_(ip))
		go serveSock(nat.end, nat.wg, crit.Msgid("https").Logger("Fatal: "), nat.server, forgeHTTPS(ip, nat.server.TLSConfig))
	}

	nat.wg.Wait()
	log.Println("DeadEnd")
}


func (nat *Nataraja) ConfUpdater(end <-chan bool,wg  *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	watcher, err := fsnotify.NewWatcher()
	exterminate(err)

	defer watcher.Close()

	err = watcher.Add( nat.IncludeVhosts.String() )
	exterminate(err)

	for {
		select {
			case event := <-watcher.Events:
				if event.Op&fsnotify.Write == fsnotify.Write {
					// here broadcast reload conf
				}

			case err := <-watcher.Errors:
				log.Println("error:", err)

			case <-end:
				return
		}
	}
}


func (nat *Nataraja) TicketUpdater(end <-chan bool,wg  *sync.WaitGroup) {
	wg.Add(1)

	time.AfterFunc( time.Duration(ticket_lifetime-(time.Now().Unix()%ticket_lifetime)) * time.Second, func(){
			go func(){
				defer wg.Done()

				ticker	:= time.Tick( time.Duration(ticket_lifetime) * time.Second)
				nat.derive_ticket()
				for {
					select {
					case <-ticker:
						nat.derive_ticket()

					case <-end:
						return
					}
				}
			}()
		}  )
}


func (nat *Nataraja) derive_ticket() {
	if nat.ticket_id == 0 {
		nat.ticket_id	= time.Now().Unix()/ticket_lifetime
	} else {
		nat.ticket_id	+= 1
	}

	h	:= hmac.New(sha256.New, []byte(nat.TicketMasterSecret))

	list	:= make([][32]byte,derive_list_size,derive_list_size)
	for i,_ := range list {
		h.Reset()
		switch	i {
		case	0:
			// the current key
			h.Write([]byte(strconv.FormatInt((nat.ticket_id)*ticket_lifetime*int64(time.Second),10)))

		case	1:
			// the next key hidden as old key
			//
			// the main idea is to cope with the risk of desync in case of multiple instances
			h.Write([]byte(strconv.FormatInt((nat.ticket_id+1)*ticket_lifetime*int64(time.Second),10)))

		default:
			// the old keys
			h.Write([]byte(strconv.FormatInt((nat.ticket_id-int64(i+1))*ticket_lifetime*int64(time.Second),10)))
		}
		copy(list[i][:],h.Sum(nil)[0:32])
	}

	nat.server.TLSConfig.SetSessionTicketKeys(list)
}


func (nat *Nataraja) OCSPUpdater(end <-chan bool,wg  *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	ticker	:= time.Tick(nat.RefreshOCSP.Get().(time.Duration))
	for {
		select {
		case <-ticker:
			nat.scan_OCSP(new(sync.WaitGroup),new(sync.Mutex))

		case <-end:
			return
		}
	}
}

func (nat *Nataraja) refresh_cert(cert *vhost.TLSConf, wg *sync.WaitGroup, lock *sync.Mutex) {
	wg.Add(1)
	defer wg.Done()

	cert.OCSP()

	lock.Lock()
	defer lock.Unlock()
	for _,sni := range cert.DNSNames() {
		nat.tlspairs[sni] = cert
	}
}


func (nat *Nataraja) scan_OCSP(wg *sync.WaitGroup, lock *sync.Mutex) {
	nat.conflock.Lock()
	defer nat.conflock.Unlock()

	nat.tlspairs = make(map[string]*vhost.TLSConf, len(nat.tlspairs))
	for _,vhost := range nat.file_vhost {
		for _,cert := range vhost.ServerPairs() {
			go nat.refresh_cert(cert, wg, lock)
		}
	}

	wg.Wait()
}


func (nat *Nataraja) SearchServable(matches []string) (servable vhost.Servable, ok bool) {
	nat.conflock.RLock()
	defer nat.conflock.RUnlock()

	for _,possible_match := range matches {
		servable, ok = nat.servable[possible_match]
		if ok {
			return
		}
	}
	return vhost.Servable {}, false
}


func (nat *Nataraja) Configure_pre_route_request() func(*http.Request,*cache.Datalog) (http.Header,url.URL) {
	return func(req *http.Request, datalog *cache.Datalog) (http.Header,url.URL) {
		d := new(types.FQDN)
		d.Set(req.Host)
		servable,_	:= nat.SearchServable( d.PathToRoot() )

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

		default_proxy	:= url.URL(nat.Proxied)
		candidat_proxy	:= url.URL(servable.Proxied)
		if candidat_proxy.Host == "" {
			return header, default_proxy
		}
		return header, candidat_proxy
	}
}


func (nat *Nataraja) GetCertificate(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	nat.conflock.RLock()
	defer nat.conflock.RUnlock()

	sni := strings.TrimRight(strings.ToLower(clientHello.ServerName),".")
	if cert, ok := nat.tlspairs[sni] ; ok {
		return cert.Certificate(), nil
	}

	labels := strings.Split(sni, ".")
	for i := range labels {
		labels[i] = "*"
		sni := strings.Join(labels, ".")
		if cert, ok := nat.tlspairs[sni]; ok {
			return cert.Certificate(), nil
		}
	}

	return nil, errors.New("No Certificate for :"+sni)
}


func (nat *Nataraja) ServeHTTP(rw http.ResponseWriter, req *http.Request){
	acclog	:= cache.NewLog(req)
	defer cache.LogHTTP(nat.cache.AccessLog, time.Now(), acclog )

	if req.ProtoMajor < 1 {
		cache.BadRequest("Obsolete Pre 1.0 Protocol").PrematureExit(rw,acclog)
		return
	}

	if req.ProtoMajor == 1 && req.ProtoMinor < 1 {
		cache.BadRequest("Obsolete 1.0 Protocol").PrematureExit(rw,acclog)
		return
	}

	if req.Host == "" {
		cache.BadRequest("No [Host:]").PrematureExit(rw,acclog)
		return
	}

	if req.TLS != nil {
		if req.TLS.ServerName == "" {
			cache.BadRequest("no tls servername").PrematureExit(rw,acclog)
			return
		}
	}

	d	:= new(types.FQDN)
	if d.Set(req.Host) != nil {
		cache.BadRequest("invalid [Host:]").PrematureExit(rw,acclog)
		return
	}

	servable,ok	:= nat.SearchServable( d.PathToRoot() )
	if !ok {
		cache.BadRequest("unknown [Host:]").PrematureExit(rw,acclog)
		return
	}

	// This is an anoying bug...
	// emilinate the situation at first connection
	// can't cope with it if resumed
	if req.TLS != nil && !req.TLS.DidResume && req.TLS.ServerName != req.Host {
//		cache.BadRequest("tls server name mismatch [Host:]" + req.TLS.ServerName + " : " + req.Host).PrematureExit(rw,acclog)
//		return
	}

	if servable.Redirect != "" {
		t := *(req.URL)
		switch req.TLS {
			case	nil:
				t.Scheme="http"
			default:
				t.Scheme="https"
				if servable.TLS {
					rw.Header().Set("Strict-Transport-Security",servable.HSTS)
				}
		}
		t.Host	= servable.Redirect
		cache.MovedPermanently(t.String()).PrematureExit(rw,acclog)
		return
	}

	if req.TLS == nil && servable.TLS {
		t := *(req.URL)
		t.Scheme= "https"
		t.Host	= req.Host
		cache.MovedPermanently(t.String()).PrematureExit(rw,acclog)
		return
	}


	// WAF HERE


	acclog.Status = -1
	nat.cache.ServeHTTP(rw,req)
}
