package router

import (
	"io/ioutil"
	"net/http"
)

func MethodNotAllowed() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	})
}

func NotFound() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
}

func ServeStaticFile(filename string) RouterConfiguration {

	handler := handleStaticFile(filename)

	return func(router *Router) {
		router.Filter(Get()).HandleFunc(handler)
		router.Handle(MethodNotAllowed())
	}
}

func handleStaticFile(filename string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		contentBytes, _ := ioutil.ReadFile(filename)
		w.WriteHeader(200)
		w.Header().Add("Content-Type", "text/html")
		w.Write(contentBytes)
	}
}

func ServeStaticDirectory(directory string, prefixToStrip string) RouterConfiguration {
	directoryFileServer := http.FileServer(http.Dir(directory))
	handler := http.StripPrefix(prefixToStrip, directoryFileServer)

	return func(router *Router) {
		router.Filter(Get()).Handle(handler)
		router.Handle(MethodNotAllowed())
	}
}
