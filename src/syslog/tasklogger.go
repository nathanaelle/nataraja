package	syslog


import	(
	"io"
	"strings"
	"time"
)


const	RFC5424TimeStamp string = "2006-01-02T15:04:05.999999Z07:00"

type	message string

func build_message(prio Priority, hostname string, appname string, pid string, msgid string, data string) message {
	return	message(strings.Join(
		[]string{
			prio.MarshallSyslog(),
			time.Now().Format(RFC5424TimeStamp),
			hostname,
			appname,
			pid,
			msgid,
			"-",
			data,
		},
		" " )+ "\000" )

}


func task_logger(pipeline <-chan message, output io.Writer)  {
	for {
		select {
			case msg := <-pipeline :
				output.Write([]byte(msg))
		}
	}


}
