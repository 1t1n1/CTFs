package main

import (
	"encoding/json"
	"fmt"
	"os"
)

const (
	MAX_CALLS = 10
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: ./gooseberry <filename>")
		os.Exit(1)
	}

	filename := os.Args[1]

	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("Error opening file:", err)
		os.Exit(1)
	}

	defer file.Close()

	parser := json.NewDecoder(file)
	var storage Data

	err = parser.Decode(&storage)
	if err != nil {
		fmt.Println("Error parsing data:", err)
		os.Exit(1)
	}

	var encoded []byte
	for i = range MAX_CALLS {
		var call Call
		err = parser.Decode(&call)

		if err != nil {
			break
		}

		call.Eval(&storage)
		encoded, err = storage.ToJson()
	}

	fmt.Println(string(encoded))

}
