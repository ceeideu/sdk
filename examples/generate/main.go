package main

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	client "github.com/ceeideu/sdk"
	"github.com/ceeideu/sdk/hem"
	"github.com/ceeideu/sdk/properties"
	"github.com/ceeideu/sdk/xid"
)

func main() {
	const chanLen = 1000

	// this channel represents bid stream
	tokenChan := make(chan xid.Token, chanLen)

	waitGroup := sync.WaitGroup{}
	waitGroup.Add(1)

	DSP(tokenChan, "dsp 1")
	DSP(tokenChan, "dsp 2")

	// publisher sending hem values

	Publisher(tokenChan, "pub 1", hem.FromEmail("foo@boo.com"))
	Publisher(tokenChan, "pub 2", hem.FromHex("c9d394ff979740df69c329c50249e939ab05f25974d8c0ad66352c91ee5338ea")) // boo@foo.com

	waitGroup.Wait()
}

func Publisher(tokenChan chan xid.Token, name string, request hem.Request) {
	// publisher sending hem values
	// _client represents publisher client sending hem values
	_client, err := client.NewXID(
		"[CEEID_ADDRESS]",
		client.XApiMockValue,
		client.WithHTTPClient(http.DefaultClient),
	)
	if err != nil {
		exit(err)
	}

	refresh(_client)

	go func() {
		for range time.Tick(time.Second) {
			refresh(_client)
		}
	}()

	go func() {
		for {
			resp, err := _client.Send(context.Background(), request.WithProperties(
				properties.WithConsent("TCF").
					WithUserAgent("generate").WithReferer("referer").WithIP("1.2.3.4")))
			if err != nil {
				log.Println(err)

				continue
			}

			token, err := _client.TokenFromXID(resp.Value)
			if err != nil {
				log.Println(err)
			}

			log.Printf("%s: [type: %s], [hem: %s], [token: %s]", name, request.Type, request.Value, token)
			tokenChan <- token
		}
	}()
}

func DSP(tokenChan <-chan xid.Token, name string) {
	_client, err := client.NewXID(
		"[CEEID_ADDRESS]",
		client.XApiMockValue,
		client.WithHTTPClient(http.DefaultClient),
	)
	if err != nil {
		exit(err)
	}

	refresh(_client)

	go func() {
		for range time.Tick(time.Second) {
			refresh(_client)
		}
	}()

	go func() {
		// this routine represents DSP decrypting xid tokens
		for tkn := range tokenChan {
			_xid, err := _client.DecryptToken(tkn)
			if err != nil {
				log.Printf("%s: decrypt error: %s", name, err)
			}

			log.Printf("%s decrypted token: xid=\"%s\"", name, _xid)

			refreshedXID, err := _client.RefreshXID(context.Background(), xid.RefreshRequest(_xid).WithProperties(
				properties.WithConsent("TCF"),
			))
			if err != nil {
				log.Printf("%s: regenerate error: %s", name, err)
			}

			log.Printf("%s: regenerated xid: %s", name, refreshedXID.Value)

			if err != nil {
				log.Printf("%s: decrypt error: %s", name, err)
			}
		}
	}()
}

func exit(e error) {
	log.Fatal(e)
}

func refresh(xid *client.XID) {
	err := xid.Refresh(context.Background())
	if err != nil {
		exit(err)
	}
}
