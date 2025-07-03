package mcp

import "fmt"

type AuthRequiredErr struct {
	ProtectedResourceValue string
	Err                    error
}

func (e AuthRequiredErr) Error() string {
	return fmt.Sprintf("authentication required: %v", e.Err)
}
