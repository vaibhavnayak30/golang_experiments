package main

import (
	"context"
	"time"
)

// Jobs 
type Job struct {
	Name string 
	Data string
	Response chan string
}