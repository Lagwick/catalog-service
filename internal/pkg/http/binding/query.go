package binding

import (
	"net/http"

	"github.com/go-playground/form/v4"
)

var formDecoder = form.NewDecoder()

type queryBinding struct{}

func (queryBinding) Name() string {
	return "URL-QUERY"
}

func (queryBinding) Bind(req *http.Request, obj any) error {
	if req == nil {
		return http.ErrNoLocation
	}

	params := req.URL.Query()

	if err := formDecoder.Decode(obj, params); err != nil {
		return err
	}

	return validate(obj)
}
