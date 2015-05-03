package	types

import	(
)






func FieldsFuncN(s string, hope int, f func(rune) bool) []string {
	p_is_sep:= true
	is_sep	:= true
	begin	:= -1
	end	:= -1
	res	:= make( []string, 0, hope )

	for i,rune := range s {
		p_is_sep = is_sep
		is_sep = f(rune)
		switch {
			case is_sep && !p_is_sep:
				end = i
				res = append( res, s[begin:end] )

			case !is_sep && p_is_sep:
				begin	= i
		}
	}

	if(begin>-1 && begin>end ) {
		res = append( res, s[begin:len(s)] )
	}

	return res
}
