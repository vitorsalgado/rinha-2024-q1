package main

import "sync"

type Cache struct {
	m sync.Map
}

func (c *Cache) put(k string, v struct{}) {
	c.m.Store(k, v)
}

func (c *Cache) has(k string) bool {
	_, ok := c.m.Load(k)
	return ok
}
