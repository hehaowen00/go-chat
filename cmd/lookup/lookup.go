package main

import "sync"

type Lookup struct {
	nodes map[string]string
	sync.RWMutex
}
