package main

import (
	"context"
	"flag"
	"log"

	accessClient "github.com/100mslive/access/pkg/client"
)

func main() {
	flagAddr := flag.String("addr", "localhost:8009", "server address")
	flag.Parse()
	if *flagAddr == "" {
		log.Fatal("addr not provided")
	}
	config := accessClient.DefaultConfig()
	config.Addr = *flagAddr
	client := accessClient.New(config)
	if err := client.Connect(context.TODO()); err != nil {
		log.Fatalf("client failed to connect to : %s, %v", config.Addr, err)
	}
	if err := client.Health(context.TODO()); err != nil {
		log.Fatalf("client failed to ping to : %s, %v", config.Addr, err)
	}
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ2ZXJzaW9uIjoyLCJ0eXBlIjoibWFuYWdlbWVudCIsImFwcF9kYXRhIjpudWxsLCJhY2Nlc3Nfa2V5IjoiNjEwYzE3Nzc1NWRhYTZjZDNlYjU3MWEwIiwiZXhwIjoxNjYxOTU2NDEyLCJqdGkiOiJkNWY0OTg4ZS0xNjFkLTQ0NTUtYTQxOS04MmNlYzFmOWQ2YzMiLCJpYXQiOjE2NTMzMTY0MTIsImlzcyI6IjYxMGMxNzc2NTVkYWE2Y2QzZWI1NzE5YyIsIm5iZiI6MTY1MzMxNjQxMiwic3ViIjoiYXBpIn0.QsDemW3MKyxDy-HQviG8mZXaXhYvkcCamntTeD9jRTI"
	authorized, err := client.AuthorizeToken(context.TODO(), token, "data", "read")
	log.Println(err)
	log.Println(authorized)
}
