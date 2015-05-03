package	syslog

import (
	"io"
	"os"
	"errors"
	"strconv"
)


type	Syslog	struct {
	devnull		*devnull
	facility	Priority
	hostname	string
	pid		string
	appname		string
	channels	[]Channel
	output		chan message
	min_sev		int
}


func New(out io.Writer, min_priority *Priority, appname string) (syslog *Syslog, err error)  {
	hostname,err	:= os.Hostname()
	if err != nil {
		hostname = "-"
	}

	if appname == "" {
		err = errors.New("syslog.New needs a non empty appname")
		return
	}

	syslog		= &Syslog {
		devnull:	new(devnull),
		facility:	min_priority.Facility(),
		hostname:	hostname,
		pid:		strconv.Itoa(os.Getpid()),
		appname:	appname,
		output:		make(chan message, 100),
		min_sev:	int(min_priority.Severity()),
	}

	if syslog.pid == "" {
		syslog.pid = "-"
	}

	syslog.channels	= []Channel {
		syslog.devnull, syslog.devnull,
		syslog.devnull, syslog.devnull,
		syslog.devnull, syslog.devnull,
		syslog.devnull, syslog.devnull,
	}

	for sev := 0;  sev <= syslog.min_sev; sev++ {
		syslog.channels[sev] = &trueChannel {
			priority:	syslog.facility|Priority(sev),
			hostname:	hostname,
			pid:		syslog.pid,
			appname:	appname,
			msgid:		"-",
			output:		syslog.output,
		}
	}

	go task_logger(syslog.output, out)

	return syslog, nil
}


func (syslog *Syslog)Channel(sev Priority) Channel {
	return syslog.channels[sev.Severity()]
}


func (syslog *Syslog)SubSyslog(sub_appname string) (sub *Syslog) {
	var appname string

	switch syslog.appname {
		case "-":	appname = sub_appname
		default:	appname = syslog.appname +"/"+ sub_appname
	}

	sub		= &Syslog {
		devnull:	syslog.devnull,
		facility:	syslog.facility,
		hostname:	syslog.hostname,
		pid:		syslog.pid,
		appname:	appname,
		output:		syslog.output,
		channels:	[]Channel {
			syslog.devnull, syslog.devnull,
			syslog.devnull, syslog.devnull,
			syslog.devnull, syslog.devnull,
			syslog.devnull, syslog.devnull,
		},
		min_sev:	syslog.min_sev,
	}

	for sev := 0;  sev <= syslog.min_sev; sev++ {
		sub.channels[sev] = &trueChannel {
			priority:	syslog.facility|Priority(sev),
			hostname:	syslog.hostname,
			pid:		syslog.pid,
			appname:	appname,
			msgid:		"-",
			output:		syslog.output,
		}
	}

	return
}
