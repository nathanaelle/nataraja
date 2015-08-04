package waf

import	(
	"index/suffixarray"
	"bytes"
)


type	WAF struct {
	bad_robots	[][]byte
	scanners	[][]byte
	robot_index	[][][]byte

}

func (waf *WAF)load_bad_robots(bad_robots[]string)  {
	br	:= make([][]byte,0,len(bad_robots))
	bri	:= make([][][]byte,256)

	for _,r := range bad_robots {
		if len(r) == 0 {
			continue
		}
		robyte	:= []byte(r)
		br 	= append(br, robyte)
		bri[robyte[0]]	= append(bri[robyte[0]], robyte )
	}
	waf.bad_robots = br
	waf.robot_index = bri
}


func (waf *WAF)GoSufArray_UserAgentIsClean(UA []byte) bool {
	index := suffixarray.New( UA )
	for _,robot := range waf.bad_robots {
		if len( index.Lookup(robot,1) ) > 0 {
			return false
		}
	}

	return true
}


func (waf *WAF)UserAgentIsClean(UA []byte) bool {
	index := Index( UA )
	for _,robot := range waf.bad_robots {
		if index.Match(robot) {
			return false
		}
	}

	return true
}

func (waf *WAF)BRI_UserAgentIsClean(UA []byte) bool {
	for i,b := range UA {
		for _,robo := range waf.robot_index[b] {
			if i+len(robo) <= len(UA) && bytes.HasPrefix(UA[i:i+len(robo)], robo) {
				return false
			}
		}
	}
	return true
}


type	state	struct {
	pos	int
	pattern	[]byte
}

func (waf *WAF)BRS_UserAgentIsClean(UA []byte) bool {
	entangled_states:= make([]state,0,50)
	next_entangled	:= make([]state,0,50)

	for i,b := range UA {
		for _,robo := range waf.robot_index[b] {
			if i+len(robo) <= len(UA) && UA[i+1] == robo[1] && UA[i+2] == robo[2] {
				next_entangled = append(next_entangled, state{ 1, robo[:] })
			}
		}

		for _,robo := range entangled_states {
			if robo.pattern[robo.pos] == b {
				if robo.pos+1 == len(robo.pattern) {
					return false
				}
				next_entangled = append(next_entangled, state{ robo.pos+1, robo.pattern[:] } )
			}
		}
		next_entangled, entangled_states =  entangled_states[:0], next_entangled[:]
	}
	return true
}
