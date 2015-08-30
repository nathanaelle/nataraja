package	vhost

import (
	types "github.com/nathanaelle/useful.types"
)


type	RedirectZone	struct {
	To	types.FQDN
	From	[]types.FQDN
}


func (rz RedirectZone)Servable(Owner, Project string) []Servable {
	ret	:= make([]Servable, 0, len(rz.From))
	to	:= string(rz.To)

	for _,from := range rz.From {
		ret = append(ret, Servable {
			Owner	: Owner,
			Project	: Project,
			Zone	: string(from),
			Redirect: to,
		})
	}

	return	ret
}
