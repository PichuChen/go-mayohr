package main

import (
	"fmt"
	"os"

	"github.com/pichuchen/go-mayohr"
)

func main() {
	username := os.Getenv("MAYOHR_USERNAME") // Change this to your username
	password := os.Getenv("MAYOHR_PASSWORD") // Change this to your password

	c := mayohr.NewClient(username, password)
	err := c.Login()
	if err != nil {
		fmt.Printf("failed to login: %v", err)
		return
	}
	fmt.Printf("ID Token: %v", c.IDToken)

	basicInfo, err := c.GetBasicInfo()
	if err != nil {
		fmt.Printf("failed to get basic info: %v", err)
		return
	}

	fmt.Printf("Basic Info: %v", basicInfo)

}
