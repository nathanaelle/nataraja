package waf

import	(
	"index/suffixarray"
)



type	state	struct {
	pos	int
	pattern	[]byte
}

type	s_state	struct {
	n_char	byte
	sub_pat	[]byte
}



type	WAF struct {
	bad_robots	[][]byte
	robot_index	[][][]byte
	lens		[]int
}





func (waf *WAF)load_bad_robots(bad_robots[]string)  {
	waf.bad_robots	= make([][]byte,0,len(bad_robots))
	waf.robot_index	= make([][][]byte,256)

	t_dedup		:= make(map[string]bool)

	for _,r := range bad_robots {
		if len(r) == 0 {
			continue
		}
		if _,ok := t_dedup[r]; ok {
			continue
		}
		t_dedup[r]	= true

		robyte	:= []byte(r)
		waf.bad_robots 			= append(waf.bad_robots, robyte[:] )
		waf.robot_index[robyte[0]]	= append(waf.robot_index[robyte[0]], robyte[:] )
	}

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

/*
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
*/

func (waf *WAF)B__RS_UserAgentIsClean(UA []byte) bool {
	entangled_states:= make([]state,0,len(waf.bad_robots)/10)
	next_entangled	:= make([]state,0,len(waf.bad_robots)/10)

	for i,b := range UA {
		for _,robo := range entangled_states {
			if robo.pattern[robo.pos] == b {
				if robo.pos+1 == len(robo.pattern) {
					return false
				}
				next_entangled = append(next_entangled, state{ robo.pos+1, robo.pattern[:] } )
			}
		}
		for _,robo := range waf.robot_index[b] {
			if i+len(robo) <= len(UA) && UA[i+1] == robo[1] && UA[i+2] == robo[2] {
				next_entangled = append(next_entangled, state{ 1, robo[:] })
			}
		}

		next_entangled, entangled_states =  entangled_states[:0], next_entangled[:]
	}
	return true
}







func (waf *WAF)BRS_UserAgentIsClean(UA []byte) bool {
	entangled_states:= make([]s_state,0,50)//len(waf.bad_robots)/10)
	next_entangled	:= make([]s_state,0,50)//len(waf.bad_robots)/10)

	for i,b := range UA[:] {
		for _,robo := range entangled_states {
			if robo.n_char == b {
				if len(robo.sub_pat) == 1 {
					return false
				}
				robo.n_char = robo.sub_pat[1]
				robo.sub_pat = robo.sub_pat[1:][:]
				next_entangled = append(next_entangled, robo )
//				next_entangled = append(next_entangled, s_state{ robo.sub_pat[1], robo.sub_pat[1:][:] })
			}
		}

		for _,robo := range waf.robot_index[b] {
			if i+len(robo) <= len(UA) && UA[i+1] == robo[1] && UA[i+2] == robo[2] {
				next_entangled = append(next_entangled, s_state{ robo[1], robo[1:][:] })
			}
		}

		next_entangled, entangled_states =  entangled_states[:0], next_entangled[:]
	}
	return true
}
