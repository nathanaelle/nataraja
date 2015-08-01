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


func (waf *WAF)OLD_UserAgentIsClean(UA []byte) bool {
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
		if len(waf.robot_index[b]) == 0 {
			continue
		}
		for _,robo := range waf.robot_index[b] {
			if bytes.HasPrefix(UA[i:], robo) {
				return false
			}
		}
	}
	return true
}
