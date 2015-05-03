package main

import (
	"reflect"
	"os"
	"log"
)


func main() {
	nataraja:= SummonNataraja()
	nataraja.ReadFlags()

	nataraja.GenerateServer()
	nataraja.SignalHandler()

	nataraja.Run()
}



func exterminate(err error)  {
	var s reflect.Value

	if err == nil {
		return
	}

	s_t	:= reflect.ValueOf(err)

	for  s_t.Kind() == reflect.Ptr {
		s_t = s_t.Elem()
	}

	switch s_t.Kind() {
		case reflect.Interface:	s = s_t.Elem()
		default:		s = s_t
	}

	typeOfT := s.Type()
	pkg	:= typeOfT.PkgPath() + "/" + typeOfT.Name()

	log.Printf("\n------------------------------------\nKind : %d %d\n%s\n\n", s_t.Kind(), s.Kind(), err.Error())

	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		if f.CanInterface() {
			log.Printf("%s %d: %s %s = %v\n", pkg, i, typeOfT.Field(i).Name, f.Type(), f.Interface())
		} else {
			log.Printf("%s %d: %s %s = %s\n", pkg, i, typeOfT.Field(i).Name, f.Type(), f.String())
		}
	}

	os.Exit(500)

	//syscall.Kill(syscall.Getpid(),syscall.SIGTERM)
}
