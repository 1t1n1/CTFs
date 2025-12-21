package main

import (
    "fmt"
    sandbox "goexe-runner/internal/sandbox"
)

func main() {
    rr, err := sandbox.PrepareRunRoot("base")
    if err != nil {
        panic(err)
    }
    fmt.Println(rr.Root)
}
