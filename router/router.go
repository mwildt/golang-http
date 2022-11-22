package router

import (
	"log"
	"net/http"
	"strings"
)

// ############################################################
// R O U T E R  R E G I S T R A T I O N  S E C T I O N
// ############################################################

type RouterConfiguration func(router *Router)

type RouterRegistration struct {
	filter RequestFilter
	path   string
	router *Router
}

// Die Router Registraction passt, wenn der Pfad matcht und die filter nicht blockieren
func (registration RouterRegistration) matches(prefix string, r *http.Request) (bool, string, ParameterMap) {

	remainingPath := strings.Replace(r.URL.Path, prefix, "", 1)

	pathMatch, matchedPath, parameterMap := MatchPath(registration.path, remainingPath)

	if !pathMatch {
		return false, "", make(ParameterMap)
	}

	if !registration.filter(r) {
		return false, "", make(ParameterMap)
	}

	return true, matchedPath, parameterMap

}

// ############################################################
// R O U T E R   S E C T I O N
// ############################################################

type Router struct {
	registrations []RouterRegistration
	handler       http.Handler
}

func EmptyRouter() *Router {
	return &Router{
		registrations: make([]RouterRegistration, 0),
		handler:       nil,
	}
}

func (router *Router) Register(path string, routerConfiguration RouterConfiguration) *Router {
	subrouter := EmptyRouter()
	routerConfiguration(subrouter)
	router.registrations = append(router.registrations, RouterRegistration{
		path:   path,
		filter: All(),
		router: subrouter,
	})
	return router
}

func (router *Router) Handle(handler http.Handler) {
	router.handler = handler
}

func (router *Router) HandleFunc(handlerFunc http.HandlerFunc) {
	router.handler = http.HandlerFunc(handlerFunc)
}

func (router *Router) Filter(filter RequestFilter) *Router {
	subrouter := EmptyRouter()
	router.registrations = append(router.registrations, RouterRegistration{
		path:   "",
		filter: filter,
		router: subrouter,
	})
	return subrouter
}

func (router *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	router.handle("", w, r)
}

func (router *Router) handle(prefix string, w http.ResponseWriter, r *http.Request) bool {
	handled := false

	for _, registration := range router.registrations {
		match, matchedPath, parameters := registration.matches(prefix, r)

		if match {

			subrouter := registration.router
			if nil != subrouter {
				newPrefix := prefix + matchedPath
				handled = subrouter.handle(newPrefix, w, r.WithContext(JoinStringParameters(r, parameters)))
				if handled {
					break
				}
			}
		}
	}

	isExactMatch := prefix == r.URL.Path

	if !handled && isExactMatch && router.handler != nil {
		log.Printf("handle request '%s'", r.URL.Path)
		router.handler.ServeHTTP(w, r)
		return true
	}
	return handled
}
