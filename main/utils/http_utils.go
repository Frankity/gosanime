package utils

import (
	"log"

	tlsclient "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
	"github.com/go-resty/resty/v2"
	srt "github.com/juzeon/spoofed-round-tripper"
)

func NewHTTPClient() *resty.Client {
	tr, err := srt.NewSpoofedRoundTripper(
		tlsclient.WithRandomTLSExtensionOrder(),
		tlsclient.WithClientProfile(profiles.Chrome_120),
	)
	if err != nil {
		log.Fatalf("Error al crear SpoofedRoundTripper: %v", err)
	}

	client := resty.New().
		SetTransport(tr).
		SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) "+
			"AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	return client
}
