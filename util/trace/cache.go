package trace

import (
	"fmt"

	"github.com/dustin/go-humanize"
)

type Cache struct {
	Tracer

	Action string
	Key    string

	size uint64
}

func (c *Cache) Reset() {
	c.Tracer.Reset()

	c.size = 0
}

func (c *Cache) Size(size uint64) {
	c.size = size
}

func (c *Cache) String() string {
	if c.complete.IsZero() {
		c.Complete()
	}

	return fmt.Sprintf(`Trace for action %s on key: %s
%s
              Size: %s
	`, c.Action, c.Key, c.Tracer.String(), humanize.Bytes(c.size))
}
