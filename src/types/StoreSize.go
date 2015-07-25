package	types

import (
	"bytes"
	"errors"
	"strconv"
)



var	validchar []byte =  []byte{ 'o','B','i','k','K','M','G','T','P','E','Z','Y' }


func is_validchar(b byte) (bool,int) {
	for i,valid := range validchar {
		if b == valid {
			return true,i
		}
	}
	return false,0
}



/*
 *	type:		Duration
 *	content:	time duration aka intergers with time units
 */
type	StoreSize	uint64

func (s *StoreSize) UnmarshalTOML(data []byte) error {
	tmp_s := bytes.Trim(data,"\"")

	max		:= len(tmp_s)-1
	digit_only	:= true
	binary_unit	:= false
	power		:= 0

	for i,b	:= range tmp_s {
		if !digit_only {
			if b == 'i' {
				if i != max-1 {
					return errors.New("invalid StoreSize : "+string(data))
				}
				binary_unit = true
				continue
			}

			if b == 'o' || b == 'B' {
				if i != max {
					return errors.New("invalid StoreSize : "+string(data))
				}
				continue
			}

			return errors.New("invalid StoreSize : "+string(data))
			continue
		}

		if b >= '0' && b <= '9' {
			continue
		}

		digit_only	= false

		ok,pos := is_validchar(b)
		if  !ok {
			return errors.New("invalid StoreSize : "+string(data))
		}

		if i < max-2 {
			return errors.New("invalid StoreSize : "+string(data))
		}

		if b == 'i' {
			return errors.New("invalid StoreSize : "+string(data))
		}

		if b == 'o' || b == 'B' {
			if i != max {
				return errors.New("invalid StoreSize : "+string(data))
			}
			continue
		}

		if pos > 3 {
			power = pos-3
			continue
		}
		power=1
	}


	v,_ := strconv.ParseUint(string(tmp_s), 10, 0 )

	if digit_only {
		*s = StoreSize(v)
		return nil
	}

	factor	:= uint64(1000)
	if binary_unit {
		factor = 1024
	}

	for power > 0 {
		v = v*factor
		power--
	}

	*s = StoreSize(v)

	return nil
}
