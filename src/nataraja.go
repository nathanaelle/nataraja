package	main

import (
	"os"
	"log"
	"net"
	"flag"
	"sync"
	"time"
	"runtime"
	"syscall"
	"os/signal"
	"net/http"
	"io/ioutil"
	"crypto/tls"

	"./hatcp"
	"./cache"

	// "gopkg.in/fsnotify.v1"
	"github.com/naoina/toml"
	"github.com/bradfitz/http2"

	syslog "github.com/nathanaelle/syslog5424"
	types "github.com/nathanaelle/useful.types"
)

const	APP_NAME	string		= "nataraja"
const	DEFAULT_CONF	types.Path	= "/etc/nataraja/config.toml"
const	DEFAULT_PRIO	syslog.Priority	= (syslog.LOG_DAEMON|syslog.LOG_WARNING)

func parser(file string, data interface{}) {
	f,_	:= os.Open(file)
	defer f.Close()

	buf,_	:= ioutil.ReadAll(f)
	err	:= toml.Unmarshal(buf, data)
	exterminate(err)
}


type Nataraja struct {
	syslog		*syslog.Syslog
	slog		*syslog.Syslog
	config		*Config
	server		*http.Server
	end		chan bool
	wg		*sync.WaitGroup
	cache		*cache.Cache

}

func SummonNataraja() (nat *Nataraja) {
	nat		= new(Nataraja)
	nat.wg		= new(sync.WaitGroup)
	return
}

func (nat *Nataraja)ReadFlags()  {
	conf_path	:= new(types.Path)
	priority	:= new(syslog.Priority)

	*priority	 = DEFAULT_PRIO
	*conf_path	 = DEFAULT_CONF

	var	numcpu	= flag.Int("cpu", 1, "maximum number of logical CPU that can be executed simultaneously")
	var	stderr	= flag.Bool("stderr", false, "send message to stderr instead of syslog")

	flag.Var(conf_path, "conf", "path to the conf" )
	flag.Var(priority , "priority", "log priority in syslog format facility.severity" )

	flag.Parse()

	switch {
		case *numcpu >runtime.NumCPU():	runtime.GOMAXPROCS(runtime.NumCPU())
		case *numcpu <1:		runtime.GOMAXPROCS(1)
		default:			runtime.GOMAXPROCS(*numcpu)
	}

	switch *stderr {
		case true:
			conn	:= syslog.Dial( "stdio", "stderr", syslog.T_LFENDED, 100 )
			if conn == nil {
				panic("no log!")
			}
			nat.syslog,_ =	syslog.New( conn, *priority, APP_NAME )

		case false:
			conn	:= syslog.Dial( "local", "", syslog.T_LFENDED, 100 )
			if conn == nil {
				panic("no log!")
			}
			nat.syslog,_ =	syslog.New( conn, *priority, APP_NAME )
	}

	nat.config = NewConfig(conf_path.String(), parser, nat.syslog.SubSyslog("config") )
}


func (nat *Nataraja) GenerateServer() {
	nat.slog	= nat.syslog.SubSyslog(nat.config.Id)
	nat.cache	= &cache.Cache {
		AccessLog	: nat.slog.Channel(syslog.LOG_NOTICE).Logger(""),
		Configure	: nat.config.Configure(),
		WAF		: nat.config.WAF(),
		ErrorLog	: nat.slog.SubSyslog("proxy").Channel(syslog.LOG_WARNING).Logger("WARNING: "),
	}

	nat.cache.Init(&cache.PassThru {})

	nat.server = &http.Server {
		Handler:			nat,
		ReadTimeout:			10 * time.Minute,
		WriteTimeout:			10 * time.Minute,
//		ConnState:			sessionLogger(),
		ErrorLog:			nat.slog.SubSyslog("connexion").Channel(syslog.LOG_INFO).Logger("INFO: "),
		TLSConfig:			&tls.Config{
			GetCertificate:			nat.config.GetCertificate,
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
}


func (nat *Nataraja) SignalHandler() {
	nat.end		 = make(chan bool)
	signalChannel	:= make(chan os.Signal)

	//signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	signal.Notify(signalChannel, syscall.SIGTERM, syscall.SIGHUP)

	go func() {
		switch <-signalChannel {
			case os.Interrupt:
				close(nat.end)

			case syscall.SIGTERM:
				close(nat.end)

			case syscall.SIGHUP:
		}
	}()

	go nat.config.ConfUpdater(nat.end,nat.wg)
	go nat.config.OCSPUpdater(nat.end,nat.wg)
}


func (nat *Nataraja) Run() {

	crit	:= nat.slog.Channel(syslog.LOG_CRIT)
	for _,ip:= range nat.config.Listen {
		go serveSock(nat.end, nat.wg, crit.Msgid("http" ).Logger("Fatal: "), nat.server, nat.forgeHTTP_(ip))
		go serveSock(nat.end, nat.wg, crit.Msgid("https").Logger("Fatal: "), nat.server, nat.forgeHTTPS(ip))
	}

	nat.wg.Wait()
	log.Println("DeadEnd")
}






func serveSock(end <-chan bool, nat_wg *sync.WaitGroup, slog *log.Logger, server *http.Server, sock net.Listener) {
	nat_wg.Add(1)
	defer nat_wg.Done()

	go func(){
		for {
			err := server.Serve(sock)
			if err == nil {
				log.Println("serveSock: WOOT")
				break
			}
			slog.Println(err)
		}
	}()

	<-end
	log.Println("serveSock: end")
}


func (nat *Nataraja) forgeHTTP_(ip types.IpAddr) (net.Listener) {
	addr,err:= ip.ToTCPAddr( "http" )
	exterminate(err)
	sock,err:= hatcp.Listen( "tcp", addr )
	exterminate(err)

	return sock
}


func (nat *Nataraja) forgeHTTPS(ip types.IpAddr) (net.Listener) {
	addr,err:= ip.ToTCPAddr( "https" )
	exterminate(err)
	tcp,err	:= hatcp.Listen( "tcp", addr )
	exterminate(err)
	sock	:= tls.NewListener( tcp, nat.server.TLSConfig )

	return sock
}



func (nat *Nataraja)ServeHTTP(rw http.ResponseWriter, req *http.Request){
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

	servable,ok	:= nat.config.SearchServable( d.PathToRoot() )
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

	acclog.Status = -1
	nat.cache.ServeHTTP(rw,req)
}
