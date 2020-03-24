package main
import (
	"pwd-cli/api"
)

func main() {
	api.NewConsole().Serve()
}