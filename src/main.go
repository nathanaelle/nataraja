package main

import (
	"os"
	"log"
	"net"
	"sync"
	"syscall"
	"reflect"
	"net/http"
	"io/ioutil"
	"os/signal"
	"crypto/tls"

	"github.com/naoina/toml"

	hatcp	"github.com/nathanaelle/pasnet"
	types	"github.com/nathanaelle/useful.types"
)


func main() {
	nataraja:= SummonNataraja()
	nataraja.ReadConfiguration()
	nataraja.LoadVirtualHosts()
	nataraja.GenerateServer()
	nataraja.SignalHandler()

	nataraja.Run()
}


//
// Helpers
//

func exterminate(err error)  {
	var s reflect.Value

	if err == nil {
		return
	}

	s_t	:= reflect.ValueOf(err)

	for  s_t.Kind() == reflect.Ptr {
		s_t = s_t.Elem()
	}

	switch s_t.Kind() {
		case reflect.Interface:	s = s_t.Elem()
		default:		s = s_t
	}

	typeOfT := s.Type()
	pkg	:= typeOfT.PkgPath() + "/" + typeOfT.Name()

	log.Printf("\n------------------------------------\nKind : %d %d\n%s\n\n", s_t.Kind(), s.Kind(), err.Error())

	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		if f.CanInterface() {
			log.Printf("%s %d: %s %s = %v\n", pkg, i, typeOfT.Field(i).Name, f.Type(), f.Interface())
		} else {
			log.Printf("%s %d: %s %s = %s\n", pkg, i, typeOfT.Field(i).Name, f.Type(), f.String())
		}
	}

	os.Exit(500)

	//syscall.Kill(syscall.Getpid(),syscall.SIGTERM)
}


func parse_toml_file(file string, data interface{}) {
	f,_	:= os.Open(file)
	defer f.Close()

	buf,_	:= ioutil.ReadAll(f)
	err	:= toml.Unmarshal(buf, data)
	exterminate(err)
}


func SignalCatcher() (<-chan bool,<-chan bool)  {
	end		:= make(chan bool)
	update		:= make(chan bool)

	go func() {
		signalChannel	:= make(chan os.Signal)

		//signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
		signal.Notify(signalChannel, syscall.SIGTERM, syscall.SIGHUP)

		defer close(signalChannel)
		defer close(update)
		defer close(end)

		for sig := range signalChannel {
			switch sig {
			case os.Interrupt, syscall.SIGTERM:
				return

			case syscall.SIGHUP:
				update <- true
			}
		}
	}()

	return end,update
}


func forgeHTTP_(ip types.IpAddr) (net.Listener) {
	addr,err:= ip.ToTCPAddr( "http" )
	exterminate(err)
	sock,err:= hatcp.Listen( "tcp", addr )
	exterminate(err)

	return sock
}


func forgeHTTPS(ip types.IpAddr, tlsconf *tls.Config) (net.Listener) {
	addr,err:= ip.ToTCPAddr( "https" )
	exterminate(err)
	tcp,err	:= hatcp.Listen( "tcp", addr )
	exterminate(err)
	sock	:= tls.NewListener( tcp, tlsconf )

	return sock
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
