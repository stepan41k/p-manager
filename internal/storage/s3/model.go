package s3

import "time"

type TrustedDevice struct {
	Hash      string    `json:"hash"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

type Metadata struct {
	Salt                []byte          `json:"salt"`
	Verifier            []byte          `json:"verifier"`
	TrustedDevices[]TrustedDevice `json:"trusted_devices"`
}
