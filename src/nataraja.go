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
	"./types"
	"./syslog"
	"./cache"

	"github.com/naoina/toml"
	"github.com/bradfitz/http2"

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
	conflock	*sync.RWMutex
	syslog		*syslog.Syslog
	slog		*syslog.Syslog
	config		*Config
	server		*http.Server
	end		chan bool
	wg		*sync.WaitGroup

}

func SummonNataraja() (nat *Nataraja) {
	nat		= new(Nataraja)
	nat.conflock	= new(sync.RWMutex)
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
		case true:	nat.syslog,_ =	syslog.New( os.Stderr, priority, APP_NAME )
		case false:	nat.syslog,_ =	syslog.New( syslog.Local(), priority, APP_NAME )
	}

	nat.config = NewConfig( conf_path.String(), parser, nat.syslog.SubSyslog("config") )
}


func (nat *Nataraja) GenerateServer() {
	nat.slog = nat.syslog.SubSyslog(nat.config.Id)

	mycache := &cache.Cache {
		AccessLog	: nat.slog.Channel(syslog.LOG_NOTICE).Logger(""),
		Prefilter	: nat.config.Routing(nat.conflock.RLocker()),
		WAF		: nat.config.WAF(),
		ErrorLog	: nat.slog.SubSyslog("proxy").Channel(syslog.LOG_WARNING).Logger("WARNING: "),
	}

	mycache.Init(&cache.PassThru {})

	nat.server = &http.Server {
		Handler			: mycache,
		ReadTimeout		: 10 * time.Minute,
		WriteTimeout		: 10 * time.Minute,
//		ConnState		: sessionLogger(),
		ErrorLog		: nat.slog.SubSyslog("connexion").Channel(syslog.LOG_INFO).Logger("INFO: "),
		TLSConfig		: nat.config.TLS(),
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

	go nat.config.ConfUpdater(nat.conflock,nat.end,nat.wg)
	go nat.config.OCSPUpdater(nat.conflock,nat.end,nat.wg)
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
