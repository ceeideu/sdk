Here’s an improved, professional version of the README:

---

# CEEId SDK

![CI Status](https://github.com/ceeideu/sdk/actions/workflows/ci.yml/badge.svg)

The CEEId SDK provides a secure, efficient integration with the CEEId xID service, enabling compliant user identity management through encrypted identifiers.

## Installation

To use the CEEId SDK, import it into your project:

```go
import "github.com/ceeideu/sdk"
```

### Client Initialization

Initialize the xID client as follows:

```go
xidClient, err := client.NewXID(
    "[CEEID_ADDRESS]",
    client.XApiMockValue,
    client.WithHTTPClient(http.DefaultClient),
)
```

Here, `client.XApiMockValue` is a sample token provided for testing and integration. For production environments, use a valid authentication token obtained from CEEId.

---

### Encryption Key Management

To keep encryption data (keys) updated, ensure periodic refreshing by calling:

```go
err := xidClient.Refresh(context.Background())
```

This method only triggers a network call when a refresh is required, ensuring efficient resource use. It’s recommended to schedule this in a routine:

```go
go func() {
    for range time.Tick(1 * time.Second) {
        err := xidClient.Refresh(context.Background())
        if err != nil {
            // handle error
        }
    }
}()
```

> **Note**: The interval of `1 * time.Second` is suggested but can be adjusted based on your requirements.

---

### xID Generation and Usage

To generate an xID, send a hashed email value (HEM) to the xID service:

```go
xid, err := xidClient.Send(context.Background(), hem.FromEmail(email))
if err != nil {
    // handle error
}
```

Invoke this function when a user logs in or interacts with the site. The `xid.Value` received represents the unique user identifier.

#### Additional Properties

To add properties to the HEM request, use:

```go
xid, err := xidClient.Send(context.Background(),
    hem.FromEmail(email).
    WithProperties(properties.WithConsent("TCF").WithIP("1.2.3.4")))
```

### Token Encryption

For secure storage or transmission, encrypt the `xID` as follows:

```go
token, err := xidClient.TokenFromXID(xid)
if err != nil {
    // handle error
}
```

The encrypted token can be safely stored or included in bid streams as needed.

#### HEM Utility Functions

`hem.FromHex(hem string)` generates an HEM request from a precomputed HEM string.

Example of a valid HEM string:

```plaintext
9a16c6b80ba80f0bafea41185219a4c8ca94b51b047c539e8170d0dfac9555e1
```

Additional HEM generation functions, such as `hem.FromPhone(phone)`, will be available in future releases.

---

### Refreshing Known User xID

When a known user (one with an existing token) interacts with the site, refresh the xID with:

```go
_xid, err := xidClient.RefreshXID(context.Background(), xid.RefreshRequest(_xid))
```

This method updates the xID to ensure data currency. Encrypted tokens can then be regenerated using `TokenFromXID(xid)` as needed.

#### Additional Properties

For additional properties in the refresh request, use:

```go
_xid, err := xidClient.RefreshXID(context.Background(),
    xid.RefreshRequest(_xid).
    WithProperties(properties.WithConsent("TCF").WithIP("1.2.3.4")))
```

---

### Token Decryption

To decrypt tokens received, for example, in the `EIDS` field of a bid request, call:

```go
decrypted, err := xidClient.DecryptToken(tkn)
```

This will return the decrypted `xID` value for authorized use in downstream processes.

--- 

This documentation provides a foundation for using the CEEId SDK effectively in identity management and token handling. For further details, refer to the SDK documentation or reach out to our support team.
