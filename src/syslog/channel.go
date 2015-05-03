package	syslog

import (
	"log"
	"io"
)



type	Channel	interface {
	io.Writer
	IsDevNull()	bool
	AppName(string) Channel
	Msgid(string)	Channel
	Logger(prefix string) *log.Logger
}





type	trueChannel	struct {
	priority	Priority
	hostname	string
	pid		string
	appname		string
	msgid		string
	output		chan<- message
}

func (d *trueChannel)Logger(prefix string) *log.Logger {
	switch d.priority.Severity(){
		case	LOG_DEBUG:
			return log.New( d, prefix, log.Lshortfile)
		default:
			return log.New( d, prefix, 0)
	}
}


func (d *trueChannel)IsDevNull() bool {
	return false
}

func (c *trueChannel)Write(d[]byte) (int,error) {
	c.output <- build_message( c.priority,c.hostname,c.appname,c.pid,c.msgid, string(d) )

	return len(d),nil
}


func (d *trueChannel)AppName(sup string) Channel {
	var appname string

	switch d.appname {
		case "-":	appname = sup
		default:	appname = d.appname +"/"+ sup
	}

	return &trueChannel {
		priority:	d.priority,
		hostname:	d.hostname,
		pid:		d.pid,
		appname:	appname,
		msgid:		d.msgid,
		output:		d.output,
	}
}

func (d *trueChannel)Msgid(msgid string) Channel {
	return &msgChannel {
		priority:	d.priority,
		hostname:	d.hostname,
		pid:		d.pid,
		appname:	d.appname,
		msgid:		msgid,
		output:		d.output,
	}
}





type	msgChannel	struct {
	priority	Priority
	hostname	string
	pid		string
	appname		string
	msgid		string

	output		chan<- message
}

func (d *msgChannel)Logger(prefix string) *log.Logger {
	switch d.priority.Severity(){
		case	LOG_DEBUG:
			return log.New( d, prefix, log.Lshortfile)
		default:
			return log.New( d, prefix, 0)
	}
}


func (d *msgChannel)AppName(string) Channel {
	return d
}

func (d *msgChannel)Msgid(string) Channel {
	return d
}

func (d *msgChannel)IsDevNull() bool {
	return false
}

func (c *msgChannel)Write(d[]byte) (int,error) {
	c.output <- build_message( c.priority,c.hostname,c.appname,c.pid,c.msgid, string(d) )

	return len(d),nil
}




type	devnull		struct {
}

func (d *devnull)Logger(prefix string) *log.Logger {
	return log.New( d, prefix, log.Lshortfile)
}


func (dn *devnull)IsDevNull() bool {
	return true
}

func (dn *devnull)AppName(string) Channel {
	return dn
}

func (dn *devnull)Msgid(string) Channel {
	return dn
}


func (dn *devnull)Write(d[]byte) (int,error) {
	return len(d),nil
}
