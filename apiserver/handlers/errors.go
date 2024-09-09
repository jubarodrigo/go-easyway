package handlers

import (
	"net/http"

	"github.com/go-chi/render"
)

type ErrResponse struct {
	Err            error `json:"-"` // low-level runtime error
	HTTPStatusCode int   `json:"-"` // http response status code

	StatusCode int    `json:"status"`            // http response status code
	StatusText string `json:"error"`             // user-level status message
	AppCode    int64  `json:"code,omitempty"`    // application-specific error code
	ErrorText  string `json:"message,omitempty"` // application-level error message, for debugging
}

func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

func ErrInvalidRequest(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 400,

		StatusCode: 400,
		StatusText: "Invalid request.",
		ErrorText:  err.Error(),
	}
}

func ErrRender(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 422,

		StatusCode: 422,
		StatusText: "Error rendering response.",
		ErrorText:  err.Error(),
	}
}

func ErrInternalServer(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 500,

		StatusCode: 500,
		StatusText: "Internal Server Error.",
		ErrorText:  err.Error(),
	}
}

func ErrNotAuthorized(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 401,

		StatusCode: 401,
		StatusText: "Invalid credentials",
		ErrorText:  err.Error(),
	}
}

var ErrNotFound = &ErrResponse{HTTPStatusCode: 404, StatusText: "Resource not found."}

// var ErrNotAuthorized = &ErrResponse{HTTPStatusCode: 401, StatusText: "Not authorized.", ErrorText: "Invalid credentials."}
var ErrForbidden = &ErrResponse{HTTPStatusCode: 403, StatusText: "Forbidden.", ErrorText: "You do not have permission to access this resource."}
var ErrDuplicateContact = &ErrResponse{HTTPStatusCode: 409, StatusText: "Duplicate contact.", ErrorText: "A contact with same name already exists."}
