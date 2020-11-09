package main

import (
	"github.com/civet148/applepay"
	"github.com/civet148/gotools/log"
)

const (
	APPLE_PAY_VERIFY_URL_PROD    = "https://buy.itunes.apple.com/verifyReceipt"     //production env
	APPLE_PAY_VERIFY_URL_SANDBOX = "https://sandbox.itunes.apple.com/verifyReceipt" //sandbox env
)

func main() {

	pay := applepay.NewApplePay("A0sd1Fw9df0", APPLE_PAY_VERIFY_URL_PROD)
	if _, err := pay.VerifyReceipt("MIIbWQYJKoZIhvcNAQcCoIIbSjCCG0YCAQExCzAJBgUrDgMCGgUA=="); err != nil {
		log.Errorf("apple pay receipt verify error [%s]", err)
	} else {
		log.Errorf("apple pay receipt verify ok")
	}
}
