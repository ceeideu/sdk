![ci](https://github.com/ceeideu/sdk/actions/workflows/ci.yml/badge.svg)

Client code will help you with xid-service integration.

Import with:

    import "github.com/ceeideu/sdk"

Initialize client with: 

	xidClient, err := client.NewXID(
		"[CEEID_ADDRESS]",
		client.XApiMockValue,
		client.WithHTTPClient(http.DefaultClient),
	)

`client.XApiMockValue` is an authentication token used for intergration and testing purposes.

---

Use:

    err := xidClient.Refresh(context.Background())

in a routine, to refresh encryption data. This data (keys) is needed to encrypt and decrypt tokens. This function will trigger a network call only when refresh is needed.

Example:

	go func() {
		for range time.Tick(1 * time.Second) {
			err := xidClient.Refresh(context.Background())
			if err != nil { ...	}
		}
	}()


**PLEASE NOTE**: `time.Tick(1 * time.Second)` is arbitrary. 1 second is recommended, but not required.

--- 

Use: 

    xid, err := xidClient.Send(context.Background(), hem.FromEmail(email))
    if err != nil { ... }

to send e-mail HEM value to xid-service. `Send` should be triggered every time a user logs in or interacts with website. This function will result in `xid.Value` which represents a unique user ID.

`hem.FromEmail(email)` returns a struct `hem.Request` that can be extended with defined properties

Example with properties:

		xid, err := xidClient.Send(context.Background(),
		hem.FromEmail(email).
		WithProperties(properties.WithConsent("TCF").WithIP("1.2.3.4")))

**PLEASE NOTE**: Unencrypted identfiers should never be forwarded into bid stream or local storage. To encrypt plain identifier please use:

			token, err := xidClient.TokenFromXID(xid)
			if err != nil { ...	}

This function call will result in token string representation which can be put into local storage or bid stream.

`hem.FromHex(hem string)` generates hem request based on hem string.

Example valid hem: `9a16c6b80ba80f0bafea41185219a4c8ca94b51b047c539e8170d0dfac9555e1`

Many other `hem.From...` functions will follow in future releases. I.e. `hem.FromPhone(phone)` will generate phone based HEM values.

--- 

Whenever a known returning user (user with token) interacts with website call:

	_xid, err := xidClient.RefreshXID(context.Background(), xid.RefreshRequest(_xid))

This will result in a fresh, updated xid value. Fresh xid should be encrypted with `TokenFromXID(xid)` function as mentioned earlier.

`xid.RefreshRequest(_xid)` returns a struct `xid.RefreshReq` that can be extended with properties

Example with properties:

		_xid, err := xidClient.RefreshXID(context.Background(), xid.RefreshRequest(_xid).
		WithProperties(properties.WithConsent("TCF").WithIP("1.2.3.4")))

---

Whenever a token needs decryption (i.e. ortb server/bidder receives EIDS token field) call:

    decrypted, err := xidClient.DecryptToken(tkn)
