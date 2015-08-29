package	vhost

import (
	"log"
//	"../types"
)


type Vhost struct {

	Internal struct {
		Owner	string
		Project	string
	}

	serverpairs	[]*TLSConf
	logger		*log.Logger

	Serve		[]ServeZone
	Redirect	[]RedirectZone
}

func New(file string, parser func(string,interface{}), logger *log.Logger ) *Vhost {
	logger.Printf("loading vhost file %s", file)
	vhost	:= new(Vhost)
	parser(file,vhost)

	vhost.logger		= logger
	vhost.serverpairs	= make([]*TLSConf,0,len(vhost.Serve))

	for _,v := range vhost.Serve {
		pair	:= v.TLSConf()
		if pair != nil {
			vhost.serverpairs = append( vhost.serverpairs, pair )
		}
	}

	return vhost
}

func (vhost *Vhost)ServerPairs() ([]*TLSConf) {
	return vhost.serverpairs
}


func (vhost *Vhost)Servables() map[string]Servable {
	zoneS	:= vhost.Served()
	zoneR	:= vhost.Redirected()

	ret	:= make(map[string]Servable, len(zoneS)+len(zoneR))

	for k,v	:= range zoneS {
		ret[k] = v
	}

	for k,v	:= range zoneR {
		ret[k] = v
	}

	return	ret

}


func (vhost *Vhost)Zones() []string {
	zoneS	:= vhost.Served()
	zoneR	:= vhost.Redirected()

	ret	:= make([]string,0, len(zoneS)+len(zoneR))
	for k,_	:= range zoneS {
		ret = append(ret, k)
	}

	for k,_	:= range zoneR {
		ret = append(ret, k)
	}

	vhost.logger.Printf("handled zones : %s", ret )
	return	ret
}

func (vhost *Vhost)Redirected() map[string]Servable {
	ret	:= make(map[string]Servable, len(vhost.Redirect))

	for _, redirect := range vhost.Redirect {
		for _,servable := range redirect.Servable(vhost.Internal.Owner, vhost.Internal.Project) {
			ret[servable.Zone] = servable
		}
	}

	return	ret
}


func (vhost *Vhost)Served() map[string]Servable {
	ret	:= make(map[string]Servable, len(vhost.Serve))

	for _, serve := range vhost.Serve {
		for _,servable := range serve.Servable(vhost.Internal.Owner, vhost.Internal.Project) {
			ret[servable.Zone] = servable
		}
	}

	return	ret

}
