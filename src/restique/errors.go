package main

func assert(err error) {
	if err != nil {
		panic(err)
	}
}

func catch(err *error) {
	if e := recover(); e != nil {
		*err = e.(error)
	}
}
