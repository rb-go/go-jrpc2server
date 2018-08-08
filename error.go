// Copyright 2009 The Go Authors. All rights reserved.
// Copyright 2012 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package jrpc2server

// ErrorCode type for error codes
type ErrorCode int

const (
	// JErrorParse Parse error - Invalid JSON was received by the server.
	// An error occurred on the server while parsing the JSON text.
	JErrorParse ErrorCode = -32700
	// JErrorInvalidReq Invalid Request - The JSON sent is not a valid Request object.
	JErrorInvalidReq ErrorCode = -32600
	// JErrorNoMethod Method not found - The method does not exist / is not available.
	JErrorNoMethod ErrorCode = -32601
	// JErrorInvalidParams Invalid params - Invalid method parameter(s).
	JErrorInvalidParams ErrorCode = -32602
	// JErrorInternal Internal error - Internal JSON-RPC error.
	JErrorInternal ErrorCode = -32603
	// JErrorServer Server error - Reserved for implementation-defined server-errors.
	JErrorServer ErrorCode = -32000
)

//var ErrNullResult = errors.New("result is null")

// Error basic error struct for API answer
type Error struct {
	// A Number that indicates the error type that occurred.
	Code ErrorCode `json:"code"` /* required */

	// A String providing a short description of the error.
	// The message SHOULD be limited to a concise single sentence.
	Message string `json:"message"` /* required */

	// A Primitive or Structured value that contains additional information about the error.
	Data interface{} `json:"data"` /* optional */
}

// Error returns error message in string format
func (e *Error) Error() string {
	return e.Message
}
