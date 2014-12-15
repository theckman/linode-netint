linode-netint
=============
[![GoDoc](https://img.shields.io/badge/netint-GoDoc-blue.svg)](https://godoc.org/github.com/theckman/linode-netint)

This is the Linode `netint` API client. This uses the Linode network internals undocumented API.

The documentation on GoDoc should be fairly clear. Tests and examples coming soon!

Short Example
-------------

```Go
package main

import (
	"fmt"
	"os"

	"github.com/theckman/linode-netint"
)

func main() {
	// get system state as seen by Atlanta
	a, err := netint.Atlanta()

	if err != nil {
		fmt.Fprint(os.Stderr, "%v\n", err.Error())
		return
	}

	// show results for Atlanta <=> Dallas from Atlanta
	fmt.Printf("To Dallas: RTT: %d, Loss: %d, Jitter: %d\n", a.Dallas.RTT, a.Dallas.Loss, a.Dallas.Jitter)
}
```
