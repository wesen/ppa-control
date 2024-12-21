package client

import "fmt"

// ClientError represents base error type for client package
type ClientError struct {
	Op   string // Operation that failed
	Addr string // Address of the client if applicable
	Err  error  // Underlying error
}

func (e *ClientError) Error() string {
	if e.Addr != "" {
		return fmt.Sprintf("%s failed for %s: %v", e.Op, e.Addr, e.Err)
	}
	return fmt.Sprintf("%s failed: %v", e.Op, e.Err)
}

func (e *ClientError) Unwrap() error {
	return e.Err
}

// ErrClientBusy indicates the client is busy (e.g., waiting for shutdown)
type ErrClientBusy struct {
	Operation string
}

func (e *ErrClientBusy) Error() string {
	return fmt.Sprintf("client is busy with %s", e.Operation)
}

// ErrClientExists indicates attempt to add a client that already exists
type ErrClientExists struct {
	Addr string
}

func (e *ErrClientExists) Error() string {
	return fmt.Sprintf("client for %s already exists", e.Addr)
}

// ErrClientNotFound indicates attempt to operate on a non-existent client
type ErrClientNotFound struct {
	Addr string
}

func (e *ErrClientNotFound) Error() string {
	return fmt.Sprintf("client for %s not found", e.Addr)
}

// NewClientError creates a new ClientError
func NewClientError(op string, addr string, err error) error {
	return &ClientError{
		Op:   op,
		Addr: addr,
		Err:  err,
	}
} 