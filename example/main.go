package main

import (
	"github.com/civet148/applepay"
	"github.com/civet148/gotools/log"
)

func main() {

	pay := applepay.NewApplePay("A0sd1Fw9df0", applepay.APPLE_PAY_VERIFY_URL_PROD)
	if _, err := pay.VerifyReceipt("MIIbWQYJKoZIhvcNAQcCoIIbSjCCG0YCAQExCzAJBgUrDgMCGgUA=="); err != nil {
		log.Errorf("apple pay receipt verify error [%s]", err)
	} else {
		log.Errorf("apple pay receipt verify ok")
	}
}
