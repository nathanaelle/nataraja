package	types

import (
	"bytes"
	"time"
)





/*
 *	type:		Duration
 *	content:	time duration aka intergers with time units
 */
type Duration time.Duration

func (d *Duration) UnmarshalTOML(data []byte) error {
	tmp_d, err := time.ParseDuration(string(bytes.Trim(data,"\"")))
	if err == nil {
		*d = Duration(tmp_d)
	}
	return err
}
