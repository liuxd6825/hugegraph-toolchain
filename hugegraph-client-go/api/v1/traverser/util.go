/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements. See the NOTICE file distributed with this
 * work for additional information regarding copyright ownership. The ASF
 * licenses this file to You under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
 * WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
 * License for the specific language governing permissions and limitations
 * under the License.
 */

package traverser

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

// intToString converts an int to its base-10 string form.
func intToString(v int) string {
	return strconv.FormatInt(int64(v), 10)
}

// int64ToString converts an int64 to its base-10 string form.
func int64ToString(v int64) string {
	return strconv.FormatInt(v, 10)
}

// floatToString converts a float64 to its string form.
func floatToString(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}

// boolToString converts a bool to its "true"/"false" string form.
func boolToString(v bool) string {
	return strconv.FormatBool(v)
}

// quoteVertexID wraps a vertex ID with the double quotes the HugeGraph server
// expects when a vertex ID is used as a URL query parameter.
func quoteVertexID(id string) string {
	return "\"" + id + "\""
}

// APIError represents an error response returned by the HugeGraph server.
// HugeGraph returns a JSON body of the form:
//
//	{
//	  "exception": "class java.lang.IllegalArgumentException",
//	  "message":   "...",
//	  "cause":     "...",
//	  "trace":     [...]
//	}
//
// on 4xx/5xx responses. APIError captures the structured fields and
// implements the standard error interface so callers can use errors.As
// to extract it.
type APIError struct {
	StatusCode int    `json:"-"`
	Exception  string `json:"exception"`
	Message    string `json:"message"`
	Cause      string `json:"cause,omitempty"`
}

// Error implements the error interface.
func (e *APIError) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.Cause != "" {
		return fmt.Sprintf("hugegraph: HTTP %d %s: %s (cause: %s)",
			e.StatusCode, e.Exception, e.Message, e.Cause)
	}
	return fmt.Sprintf("hugegraph: HTTP %d %s: %s",
		e.StatusCode, e.Exception, e.Message)
}

// parseAPIError builds an error from a non-2xx response body. When the body
// contains a recognizable HugeGraph error payload, an *APIError is returned
// with the structured fields populated. Otherwise a generic error is
// returned that includes the raw body for debugging.
func parseAPIError(statusCode int, body []byte) error {
	apiErr := &APIError{StatusCode: statusCode}
	if err := json.Unmarshal(body, apiErr); err == nil && apiErr.Message != "" {
		return apiErr
	}
	if len(body) == 0 {
		return fmt.Errorf("hugegraph: HTTP %d (empty body)", statusCode)
	}
	return fmt.Errorf("hugegraph: HTTP %d (body: %s)",
		statusCode, string(body))
}

// checkHTTPStatus returns nil when the response status code indicates
// success (2xx); otherwise it returns an *APIError (or a generic error
// when the body cannot be parsed as a HugeGraph error).
//
// body must be the already-read response body. The function does not read
// from res.Body itself so callers may still consume it after this call.
func checkHTTPStatus(res *http.Response, body []byte) error {
	if res.StatusCode >= 200 && res.StatusCode < 300 {
		return nil
	}
	return parseAPIError(res.StatusCode, body)
}
