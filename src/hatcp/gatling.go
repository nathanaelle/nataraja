package	hatcp

import	(
	"time"
)


type	bullet	struct {
	F	func(int, interface {}) error
	A	interface{}
}

type	gatling []bullet

func gatling_run(g gatling, fd int) (err error) {
	for _,b := range g {
		err	= b.F(fd, b.A)
		if err != nil {
			return err
		}
	}
	return nil
}

func bullet_bool(f func(int,bool)error) (func(int,interface{})error)  {
	return func(fd int, i interface{}) error {
		v, ok := i.(bool)
		if ok {
			return f(fd, v)
		}
		panic("WTF Wrong Type for bool !!!")
	}
}

func bullet_int(f func(int,int)error) (func(int,interface{})error)  {
	return func(fd int, i interface{}) error {
		v, ok := i.(int)
		if ok {
			return f(fd, v)
		}
		panic("WTF Wrong Type for int !!!")
	}
}

func bullet_duration(f func(int,time.Duration)error) (func(int,interface{})error)  {
	return func(fd int, i interface{}) error {
		v, ok := i.(time.Duration)
		if ok {
			return f(fd, v)
		}
		panic("WTF Wrong Type for Duration !!!")
	}
}
