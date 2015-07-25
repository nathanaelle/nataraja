package cache


import (
	"io"
	"net/http"
	"strconv"
)


type	Status	struct {
	Code	int
	Message	string
}



func InternalServerError(message string) *Status {
	return &Status { http.StatusInternalServerError, message }
}


func BadRequest(message string) *Status {
	return &Status { http.StatusBadRequest, message }
}

func MovedPermanently(message string) *Status {
	return &Status { http.StatusMovedPermanently, message }
}


func RangeNotSatisfiable(message string) *Status {
	return &Status { http.StatusRequestedRangeNotSatisfiable, message }
}

func NotModified() *Status {
	return &Status { http.StatusNotModified, "" }
}


func (s *Status)PrematureExit(rw http.ResponseWriter, datalog *Datalog) {
	datalog.Status		= s.Code
	datalog.BodySize	= 0
	datalog.ContentType	= ""

	switch s.Code {
		case	http.StatusBadRequest:
			msg			:="400 Bad Request\n\n"+ s.Message +"\n"
			datalog.BodySize	= int64(len(msg))
			datalog.ContentType	= "text/plain"

			rw.Header().Set("Content-Length", strconv.FormatInt(datalog.BodySize,10))
			rw.Header().Set("Content-Type", "text/plain")
			rw.WriteHeader(http.StatusBadRequest)
			io.WriteString(rw, msg)

		case	http.StatusInternalServerError:
			msg			:="500 Internal Server Error\n\n"+ s.Message +"\n"
			datalog.BodySize	= int64(len(msg))
			datalog.ContentType	= "text/plain"

			rw.Header().Set("Content-Length", strconv.FormatInt(datalog.BodySize,10))
			rw.Header().Set("Content-Type", "text/plain")
			rw.WriteHeader( http.StatusInternalServerError )
			io.WriteString(rw, msg)

		case	http.StatusMovedPermanently:
			rw.Header().Set("Location", s.Message)
			rw.WriteHeader( http.StatusMovedPermanently )

		case	http.StatusRequestedRangeNotSatisfiable:
			rw.Header().Set("Content-Range", s.Message)
			rw.WriteHeader( http.StatusRequestedRangeNotSatisfiable )

		case	http.StatusNotModified:
			rw.WriteHeader( http.StatusNotModified )

		default:
	}
}
