package main

import (
	"time"
)

// message representing a single message
type message struct {
	Name    string
	Message string
	When    time.Time
}
