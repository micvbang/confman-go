package httphelpers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestStatusHandlerNoError verifies that status code 200 is returned when
// Status is nil.
func TestStatusHandlerNoError(t *testing.T) {
	statusCode, _ := testStatusHandler(t, nil)
	require.Equal(t, http.StatusOK, statusCode, "Unexpected http status code")
}

// TestStatusHandlerNoErrorData verifies that status code 200 is returned
// when Status is nil, and that the expected data is written to the response.
func TestStatusHandlerNoErrorData(t *testing.T) {
	type testResponse struct {
		Number int    `json:"data"`
		Str    string `json:"str"`
	}

	expectedResponse := testResponse{
		Number: 1,
		Str:    "im string!",
	}

	// Execute status handler and write response to response recorder
	w := httptest.NewRecorder()
	StatusHandler(func(w http.ResponseWriter, r *http.Request) Status {
		return WriteJSON(w, expectedResponse)
	}).ServeHTTP(w, &http.Request{})

	// Assert data
	require.Equal(t, http.StatusOK, w.Code, "Unexpected http status code")

	response := testResponse{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err, "Unexpected error when json unmarshalling")
	require.Equal(t, expectedResponse, response)
}

// TestStatusHandlerStatusCodes verifies that all predefined Status types
// return the expected status code and textual response.
func TestStatusHandlerStatusCodes(t *testing.T) {
	for status, httpStatusCode := range errToStatusCode {
		statusCode, response := testStatusHandler(t, status)
		require.Equal(t, httpStatusCode, statusCode, "Unexpected http status code")
		require.Equal(t, status.Error(), response.Response, "Unexpected status response")
	}
}

// TestStatusHandlerCustomStatus verifies that the expected status code and
// http status code are returned for CustomStatus errors.
func TestStatusHandlerCustomStatus(t *testing.T) {
	tests := []struct {
		statusCode     string
		httpStatusCode int
	}{
		{"give_monies", http.StatusPaymentRequired},
		{"you_broke_it", http.StatusBadRequest},
		{"bad_guy_come_in", http.StatusForbidden},
		{"im_teapot", http.StatusTeapot},
		{"hi_june", http.StatusExpectationFailed},
	}

	for i, test := range tests {
		statusCode, response := testStatusHandler(t, NewStatus(test.statusCode, test.httpStatusCode))
		require.Equal(t, test.httpStatusCode, statusCode, fmt.Sprintf("Test %d: unexpected http status code", i+1))
		require.Equal(t, test.statusCode, response.StatusCode, fmt.Sprintf("Test %d: unexpected status code", i+1))
	}
}

// TestStatusHandlerCustomStatusData verifies that the expected status code,
// http status code, and data is returned for CustomStatus errors.
func TestStatusHandlerCustomStatusData(t *testing.T) {
	type testData struct {
		Data string `json:"data"`
	}

	expectedData := testData{"im data"}
	status := NewStatusData("give_monies", http.StatusPaymentRequired, &expectedData)
	statusCode, response := testStatusHandler(t, status)
	require.Equal(t, status.httpStatusCode, statusCode, "Unexpected http status code")
	require.Equal(t, status.statusCode, response.StatusCode, "Unexpected status code")

	responseData := response.Data.(map[string]interface{})["data"]
	require.Equal(t, expectedData.Data, responseData, "Unexpected response data")
}

// TestStatusHandlerCustomStatusData verifies that the expected status code,
// http status code, and data is returned for CustomStatus errors when the
// data is nil.
func TestStatusHandlerCustomStatusNilData(t *testing.T) {
	type testData struct {
		Data string `json:"data"`
	}

	status := NewStatusData("bad_bad", http.StatusBadRequest, nil)
	statusCode, response := testStatusHandler(t, status)
	require.Equal(t, status.httpStatusCode, statusCode, "Unexpected http status code")
	require.Equal(t, status.statusCode, response.StatusCode, "Unexpected status code")

	require.Equal(t, nil, response.Data, "Unexpected response data")
}

func testStatusHandler(t *testing.T, status Status) (int, StatusHandlerErrorResponse) {
	w := httptest.NewRecorder()
	StatusHandler(func(w http.ResponseWriter, r *http.Request) Status {
		return status
	}).ServeHTTP(w, &http.Request{})

	return w.Code, parseResponse(t, w.Body.Bytes())
}

func parseResponse(t *testing.T, data []byte) StatusHandlerErrorResponse {
	response := StatusHandlerErrorResponse{}
	if len(data) == 0 {
		return response
	}

	err := json.Unmarshal(data, &response)
	require.NoError(t, err, "Unexpected error when parsing StatusHandlerErrorResponse")
	return response
}
