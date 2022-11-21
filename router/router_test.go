package router

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestShift(t *testing.T) {

	test := func(path string, success bool, token string, remaining string) {
		act_success, act_token, act_remaining := shiftPart(path)

		if success != act_success {
			t.Errorf("wrong success. expected %v but was %v", success, act_success)
		}

		if token != act_token {
			t.Errorf("wrong token. expected %s but was %s", token, act_token)
		}

		if remaining != act_remaining {
			t.Errorf("wrong remaining. expected %s but was %s", remaining, act_remaining)
		}
	}

	test("/api/todo", true, "", "api/todo")
	test("api/todo", true, "api", "todo")
	test("todo", true, "todo", "")
	test("", true, "", "")
	test("{irgendwas}/hallo", true, "{irgendwas}", "hallo")
}

func TestMatches(t *testing.T) {
	ok, matched, params := MatchPath("/api/todo", "/api/todo")

	if !ok {
		t.Error("No Match")
	}

	if matched != "/api/todo" {
		t.Error("wrong matched path")
	}

	if len(params) > 0 {
		t.Error("no params expected")
	}

}

func TestMatchesVariable(t *testing.T) {
	ok, pathMatched, params := MatchPath("/api/todo/{todoId}", "/api/todo/123-123")

	if !ok {
		t.Error("No Match")
	}

	if pathMatched != "/api/todo/123-123" {
		t.Errorf("wrong matched path %s", pathMatched)
	}

	if len(params) != 1 {
		t.Error("expected 1 parameter")
	}

	if params[ParamsKey("todoId")] != "123-123" {
		t.Errorf("wrong todoId-parameer. expected '123-123' but was %s", params[ParamsKey("todoId")])
	}

}

func TestMatchesVariableNoValue(t *testing.T) {
	ok, _, _ := MatchPath("/api/todo/{todoId}", "/api/todo/")
	if ok {
		t.Error("das hier sollte nicht matchen")
	}
}

//func TestMatchesVariableMissingTrailingSlash(t *testing.T) {
//ok, _, _ := MatchPath("/api/todo/", "/api/todo")
//	if ok {
//		t.Error("das hier sollte nicht matchen")
//	}
//}

func TestMatchesVariableMissingTailingPart(t *testing.T) {
	ok, _, _ := MatchPath("/api/todo/", "/api/")

	if ok {
		t.Error("das hier sollte nicht matchen")
	}

}

func TestMatchesVariableWithRest(t *testing.T) {
	ok, pathMatched, params := MatchPath("/api/todo/{todoId}", "/api/todo/123-123/action")

	if !ok {
		t.Error("No Match")
	}

	if pathMatched != "/api/todo/123-123" {
		t.Errorf("wrong matched path %s", pathMatched)
	}

	if len(params) != 1 {
		t.Error("expected 1 parameter")
	}

	if params[ParamsKey("todoId")] != "123-123" {
		t.Errorf("wrong todoId-parameer. expected '123-123' but was %s", params[ParamsKey("todoId")])
	}
}

func dummiHandler(status int, response string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		for k, v := range ReadParameters(r) {
			w.Header().Set(string(k), v)
		}
		fmt.Fprint(w, response)
	})
}

func TestRouting(t *testing.T) {

	router := EmptyRouter().Register("/api/todos", func(todosRouter *Router) {

		todosRouter.Register("/{todoId}", func(todoRouter *Router) {
			todoRouter.Filter(Get()).Handle(dummiHandler(200, "todo-id-handler"))
			todoRouter.Handle(dummiHandler(405, "405:todo-id"))
		})

		todosRouter.Filter(Get()).Handle(dummiHandler(200, "todo-list-handler"))

		// default Handler für exacte matches auf /api/todos
		todosRouter.Handle(dummiHandler(405, "405:todo-list"))

	}).Register("/**", func(catchAll *Router) {
		catchAll.Handle(dummiHandler(404, "404:catchAll global"))
	})

	test := func(method string, path string, expectedStatus int, exprectedResponse string, adv func(rr *httptest.ResponseRecorder)) {
		req, err := http.NewRequest(method, path, nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()

		router.handle("", rr, req)

		if rr.Body.String() != exprectedResponse {
			t.Errorf("[%s:%s] handler returned unexpected body: got '%v' want '%v'",
				method, path, rr.Body.String(), exprectedResponse)
		}

		if status := rr.Code; status != expectedStatus {
			t.Errorf("[%s:%s] handler returned wrong status code: got %v want %v",
				method, path, status, expectedStatus)
		}

		adv(rr)

	}

	test("GET", "/api/todos", http.StatusOK, "todo-list-handler", func(rr *httptest.ResponseRecorder) {})
	test("PATCH", "/api/todos", http.StatusMethodNotAllowed, "405:todo-list", func(rr *httptest.ResponseRecorder) {})
	test("GET", "/api/todos/123-456", http.StatusOK, "todo-id-handler", func(rr *httptest.ResponseRecorder) {
		if rr.Header().Get("todoId") != "123-456" {
			t.Error("der erwartete Parameter-Echo Header ist nicht da")
		}
	})
	test("PATCH", "/api/todos/123-456", http.StatusMethodNotAllowed, "405:todo-id", func(rr *httptest.ResponseRecorder) {})

	test("GET", "/api/todos/123-456/not-existent", http.StatusNotFound, "404:catchAll global", func(rr *httptest.ResponseRecorder) {})
	test("GET", "/api/not-existent", http.StatusNotFound, "404:catchAll global", func(rr *httptest.ResponseRecorder) {})
	test("GET", "/not-existent", http.StatusNotFound, "404:catchAll global", func(rr *httptest.ResponseRecorder) {})
}
