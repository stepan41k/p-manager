package s3

type Metadata struct {
	Salt                []byte   `json:"salt"`
	Verifier            []byte   `json:"verifier"`
	TrustedDeviceHashes []string `json:"trusted_device_hashes"`
}
