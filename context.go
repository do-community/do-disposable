package main

import (
	c "context"
	"time"
)

// This defines the context which is used in the application.
func context() c.Context {
	x, _ := c.WithTimeout(c.TODO(), 10*time.Second)
	return x
}
