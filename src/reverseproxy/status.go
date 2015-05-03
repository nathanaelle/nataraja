package reverseproxy


import (
	"io"
	"net/http"
)


type	Status	struct {
	Code	int
	Message	string
}

func (s *Status)PrematureExit(rw http.ResponseWriter, datalog *Datalog) {
	switch s.Code {
		case	http.StatusBadRequest:
			datalog.Status = http.StatusBadRequest
			rw.WriteHeader(http.StatusBadRequest)
			io.WriteString(rw, "400 Bad Request\n"+ s.Message +"\n")

		case	http.StatusMovedPermanently:
			datalog.Status = http.StatusMovedPermanently
			rw.Header().Set("Location", s.Message)
			rw.WriteHeader( http.StatusMovedPermanently )

		default:
	}
}
