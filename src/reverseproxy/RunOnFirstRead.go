// cut 'n paste from net/http/httputil/reverseproxy.go
// usefull in first time but not complete
package reverseproxy

import (
	"io"
)

type runOnFirstRead struct {
	io.Reader

	fn func() // Run before first Read, then set to nil
}


func (c *runOnFirstRead) Read(bs []byte) (int, error) {
	if c.fn != nil {
		c.fn()
		c.fn = nil
	}
	return c.Reader.Read(bs)
}
