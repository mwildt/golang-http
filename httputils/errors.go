package httputils

import (
	"net/http"

	"github.com/mwildt/golang-utils/errorutils"
)

func Status(err error) int {
	switch err.(type) {
	case errorutils.NotFound:
		return http.StatusNotFound
	case errorutils.ClientError:
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
