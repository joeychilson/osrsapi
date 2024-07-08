# osrsapi

A library for interacting with the Old School Runescape API.

## Usage

```go
package main

import (
	"context"
	"fmt"

	"github.com/joeychilson/osrsapi"
)

func main() {
	client := osrsapi.NewClient()

	item, err := client.Item(context.Background(), 4151)
	if err != nil {
		panic(err)
	}

	fmt.Println(item)
}
```
