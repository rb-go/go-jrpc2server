// Copyright 2009 The Go Authors. All rights reserved.
// Copyright 2012 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package jrpc2server

// ErrorCode basic error code type
type ErrorCode int

const (
	// E_PARSE - parse error
	E_PARSE ErrorCode = -32700
	// E_INVALID_REQ - invalid request
	E_INVALID_REQ ErrorCode = -32600
	// E_NO_METHOD - method not found
	E_NO_METHOD ErrorCode = -32601
	// E_BAD_PARAMS - bad parametrs
	E_BAD_PARAMS ErrorCode = -32602
	// E_INTERNAL - internal error
	E_INTERNAL ErrorCode = -32603
	// E_SERVER - server error
	E_SERVER ErrorCode = -32000
)

//var ErrNullResult = errors.New("result is null")

// Error basic error struct for answer
type Error struct {
	// A Number that indicates the error type that occurred.
	Code ErrorCode `json:"code"` /* required */

	// A String providing a short description of the error.
	// The message SHOULD be limited to a concise single sentence.
	Message string `json:"message"` /* required */

	// A Primitive or Structured value that contains additional information about the error.
	Data interface{} `json:"data"` /* optional */
}

// Error returns error message string
func (e *Error) Error() string {
	return e.Message
}
