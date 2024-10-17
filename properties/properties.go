package properties

type Value map[string]string

const (
	// Determines whether the user has consented to processing the request. Compliant with the TCF 2.2 standard. Required.
	Consent = "consent"
	// Contains the value of the user's browser user agent. Recommended but not required.
	UserAgent = "user-agent"
	// Contains the domain or "surface" where the user is navigating. Recommended but not required.
	Referer = "referer"
	// The IP address of the user. This property is optional and can be included for additional context or tracking purposes.
	IPAddress = "ip"
)

func WithConsent(consent string) Value {
	v := make(Value)
	v[Consent] = consent

	return v
}

func (v Value) WithUserAgent(ua string) Value {
	v[UserAgent] = ua

	return v
}

func (v Value) WithReferer(referer string) Value {
	v[Referer] = referer

	return v
}

func (v Value) WithIP(ip string) Value {
	v[IPAddress] = ip

	return v
}
