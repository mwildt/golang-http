package router

import "net/http"

type RequestFilter = func(r *http.Request) bool

func All() RequestFilter {
	return func(r *http.Request) bool {
		return true
	}
}

func Get() RequestFilter {
	return Method("GET")
}

func Put() RequestFilter {
	return Method("PUT")
}

func Delete() RequestFilter {
	return Method("DELETE")
}

func Patch() RequestFilter {
	return Method("PATCH")
}

func Post() RequestFilter {
	return Method("POST")
}

func Method(method string) RequestFilter {

	return func(r *http.Request) bool {
		return r.Method == method
	}
}
