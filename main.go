package main

import (
	"github.com/mittwald/brudi/cmd"
	"github.com/mittwald/brudi/internal"
)

func init() {
	internal.InitLogger()
}

func main() {
	err := cmd.Execute()
	if err != nil {
		panic(err)
	}
}
