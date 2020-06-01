package httphelpers

import (
	"errors"
	"net/http"
)

// Status is used for reporting http statuses to the statusHandler.
// If an Status is returned to statusHandler, which is not defined in
// the list below, it will be treated as an internal server error.
// This is meant to be a convenience for the user, letting the
// status handler format errors consistently.
type Status error

type CustomStatus struct {
	httpStatusCode int
	statusCode     string
	data           interface{}
}

func (c CustomStatus) Error() string {
	return c.statusCode
}

func NewStatus(statusCode string, httpStatusCode int) CustomStatus {
	return NewStatusData(statusCode, httpStatusCode, nil)
}

func NewStatusData(statusCode string, httpStatusCode int, data interface{}) CustomStatus {
	return CustomStatus{
		statusCode:     statusCode,
		httpStatusCode: httpStatusCode,
		data:           data,
	}
}

// Status messages that are currently handled by statusHandler.
// Add more when required.
var (
	StatusNotFound   Status = errors.New("requested resource not found")
	StatusCreated    Status = errors.New("resource created")
	StatusConflict   Status = errors.New("conflicting resource request")
	StatusBadRequest Status = errors.New("bad request")
	StatusForbidden  Status = errors.New("forbidden")

	errToStatusCode = map[Status]int{
		StatusNotFound:   http.StatusNotFound,
		StatusConflict:   http.StatusConflict,
		StatusCreated:    http.StatusCreated,
		StatusBadRequest: http.StatusBadRequest,
		StatusForbidden:  http.StatusForbidden,
	}
)

const (
	internalServerError = "internal server error"
)

type StatusHandlerErrorResponse struct {
	Response   string      `json:"response"`
	StatusCode string      `json:"status_code"`
	Data       interface{} `json:"data"`
}

func StatusHandler(f func(w http.ResponseWriter, r *http.Request) Status) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := GetLogger(r)

		err := f(w, r)
		if err == nil {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Handle pre-defined statuses
		if statusCode, ok := errToStatusCode[err]; ok {
			w.WriteHeader(statusCode)
			WriteJSON(w, StatusHandlerErrorResponse{
				Response: err.Error(),
			})
			return
		}

		// Handle custom statuses
		if customStatus, ok := err.(CustomStatus); ok {
			w.WriteHeader(customStatus.httpStatusCode)
			WriteJSON(w, StatusHandlerErrorResponse{
				StatusCode: customStatus.statusCode,
				Data:       customStatus.data,
			})
			return
		}

		statusCode := http.StatusInternalServerError
		errMsg := internalServerError

		log.Errorf("ERROR: '%s' for %+v <- %+v: %+v%+v %+v:\n\t%q\n", err, r.RemoteAddr, r.Method, r.Host, r.URL, statusCode, errMsg)
		w.WriteHeader(statusCode)
		WriteJSON(w, StatusHandlerErrorResponse{
			Response: errMsg,
		})
		return
	}
}
