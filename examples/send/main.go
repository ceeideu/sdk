package main

import (
	"context"
	"log"
	"net/http"

	client "github.com/ceeideu/sdk"
	"github.com/ceeideu/sdk/hem"
)

func main() {

	_client, err := client.NewXID(
		"http://localhost:9000",
		client.XApiInMemoryValue,
		client.WithHTTPClient(http.DefaultClient),
	)
	if err != nil {
		exit(err)
	}

	err = _client.Refresh(context.Background())
	if err != nil {
		exit(err)
	}

	email := "alama@kota.pl"

	resp, err := _client.Send(context.Background(), hem.FromEmail(email))
	if err != nil {
		exit(err)
	}

	token, err := _client.TokenFromXID(resp.Value)
	if err != nil {
		log.Println(err)
	}

	log.Printf("%s: token: \"%s\", xid: \"%s\"", email, token, resp)
}

func exit(e error) {
	log.Fatal(e)
}
