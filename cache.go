package main

type Cache struct {
	m map[string]struct{}
}

func (c *Cache) put(k string, v struct{}) {
	c.m[k] = v
}

func (c *Cache) has(k string) bool {
	_, ok := c.m[k]
	return ok
}
