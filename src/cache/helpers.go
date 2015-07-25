package	cache

import	(
	"time"
)


func httpDate2Time(httpdate string, def time.Time) time.Time {
	ret,err	:= time.Parse(time.RFC1123, httpdate)
	if err == nil {
		return ret
	}

	ret,err	= time.Parse(time.RFC1123Z, httpdate)
	if err == nil {
		return ret
	}

	return def
}
