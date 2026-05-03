package hikvideo

import (
	"context"
	"fmt"
	"testing"
)

func TestCapture(t *testing.T) {
	cli := NewClient("2050227470996942860", "MIICdwIBADANBgkqhkiG9w0BAQEFAASCAmEwggJdAgEAAoGBAIdN791RDH9Ia5W0Gq+r0eH0H620UXtBOBUN4uLGYix9KuE5pDV5BoGOor5qzos6ugBuck+spt6sxl7o2x4Q5dFztC/kg56nqlBcWvD0NlPNUlFs+9QDeCkQvVAz97ToHLrIrakRhoHJh/E/928ucqYwIQJGXWPDOcd0zOne0Nb/AgMBAAECgYA0uoaztRYttEdY++s6crdEWyLNSuxGIFB+w/6babxwogbH5vK7dAK5EqERnAYJ9ETwThp2Ok59kM9txUk/Gk61FEkTMX0Ii7piJb1qk+D53lnY4Tu25jB60M4GMycOGxitJ0729Qy0NymHQmuda8xvH7rqGNfItRRO2JcHYGHxQQJBAL9CPgHkXSIpekUtWAdTHEcjVctVXjuG9Z72pyHy+U+dDqk8JtXwlEkwQ6/P8UEZBxjSsJL/0apxUvKsxgJa+HkCQQC1Guf00o5LN6VgCIHbu1iRfi0d2vta5I+wEtlU7DkHNdtaimypfJNfFkzv6PJn8FhxngH0x68Tev6K96+Drd03AkEApuBHdiMo18vU8VL1Ab8UZ0V/cCCWTd4dpYuUnFyCB2MEDcl8ISL+XzWLeXU4DRKnTJNYmYo4CD1EoJT7V8bEEQJBAJoEqtWrp3XSeiMkuQNc3aLGUqo8TF1tWcGdFhVB2/IE3GqwpF6zYkWQmpfBXT4FycG+Zd19YKhJhmY65Jow55sCQH82gTxPNZVZQ8zR9Ba1kRDWJDUJBpNVXEmVuU5U5xzK+i5zN6X/m3GYZDzLDFIXbYqd4CEFa2qdIsfaEbPNunU=", "15161515197", "Jason927")

	cli.Encrypted = true
	
	url, err := cli.Capture(context.Background(),
		"GS1239868",
		1,
	)

	if err != nil {
		t.Error(err)
	}

	fmt.Println("图片地址:", url)
}
