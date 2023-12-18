# Krawl

Krawl is a web crawling library written in Go, designed to efficiently crawl and extract data from web pages.

## Getting Started


## Installing

To use `krawl` in your Go project, follow these steps:

1. Ensure you have Go installed on your system. `krawl` requires Go version 1.16 or higher. You can check your current Go version with:

    ```bash
    go version
    ```

2. Inside your Go project, add `krawl` as a module dependency:

    ```bash
    go get github.com/AgentGino/krawl
    ```

3. Import `krawl` in your Go files where you need it:

    ```go
    import "github.com/AgentGino/krawl"
    ```


## Usage

A quick example of how to use the library in your Go application:

```go
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/AgentGino/krawl"
)

func SaveToJSON(data []*krawl.RunnerOutput, filename string) error {
	fileData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, fileData, 0644)
}

func main() {
	allData, err := krawl.Runner(
		krawl.RunnerInput{
			StartUrl:     "https://go.dev",
			PathPatterns: []string{"doc/.*"},
		},
	)
	if err != nil {
		panic(err)
	}

	// Saving data
	if err := SaveToJSON(allData, "output.json"); err != nil {
		fmt.Errorf("Error while save to json file", err)
	}
	fmt.Println("  ✔️  Data saved to output.json")
}

```

## Planned Features

- Concurrent web crawling capabilities (coming soon).
- Respect robots.txt file


## License

This project is licensed under the Apache 2.0 License - see the [LICENSE.md](https://www.apache.org/licenses/LICENSE-2.0) file for details.
