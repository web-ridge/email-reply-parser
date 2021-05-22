package email_reply_parser //nolint:stylecheck,golint

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIsQuotedEmailStart(t *testing.T) {
	shouldReturnTrue := []string{
		"On Monday, November 4, 2013 4:29 PM, John Smith <john.smith@example.org> wrote:",
		"2013/11/1 John Smith <john@smith.org>",
		"On Monday, November 4, 2013 4:29 PM, John Smith <john.smith@example.org> wrote:",
	}
	for _, should := range shouldReturnTrue {
		if isQuotedEmailStart(strings.ToLower(should)) != true {
			t.Errorf("Should return true: %v", should)
		}
	}
}

func TestIsName(t *testing.T) {
	shouldReturnTrue := []string{
		"Richard Lindhout",
		"Karen The Green",
		"Jan de Smit",
		"Richard Lindhout | Software Engineer",
		"Richard Lindhout, Software Engineer",
	}
	for _, should := range shouldReturnTrue {
		if isName(should) != true {
			t.Errorf("Should return true: %v", should)
		}
	}

	shouldReturnFalse := []string{
		"tel 01666666 ",
		"email 01666666 ",
		"kvk 01666666 ",
		"btw 01666666 ",
		"TEL 01666666 ",
		"Email 01666666 ",
		"KvK 01666666 ",
		"BTW 01666666 ",
		"Street 2, City, Zeeland, 4694EG, NL",
	}
	for _, should := range shouldReturnFalse {
		if isName(should) != false {
			t.Errorf("Should return false: %v", should)
		}
	}
}

func TestPossibleSignature(t *testing.T) {
	shouldReturnTrue := []string{
		"WEB webRidge.nl <https://webridge.nl/>",
		"IBAN NL93 BUNQ 0000 1111 22",
		"BTW NL0000000AA0",
		"Richard Lindhout",
		"Karen The Green",
		"Graphic Designer",
		"Jan de Smit",
		"Tel: +44423423423423",
		"Fax: +44234234234234",
		"Richard Lindhout | Software Engineer",
		"Richard Lindhout, Software Engineer",
		"tel 01666666",
		"email 01666666",
		"kvk 01666666",
		"btw 01666666",
		"TEL 01666666",
		"Email 01666666",
		"KvK 01666666",
		"BTW 01666666",
		"karen@webby.com",
		"www.thing.com",
		"thing.com",
		// TODO: address lines
	}
	for _, should := range shouldReturnTrue {
		if isPossibleSignatureLine(should) != true {
			t.Errorf("Should return true: %v", should)
		}
	}
	shouldReturnFalse := []string{
		"Lorem Ipsum is simply dummy text of the printing and typesetting industry.",
		"Haha dit is supper grappig!",
		"Ok!",
		"Haha ok",
		"That's not true",
		"Her email address is karen@webby.com",
		"Her website is facebook.com",
	}
	for _, should := range shouldReturnFalse {
		if isPossibleSignatureLine(should) != false {
			t.Errorf("Should return false: %v", should)
		}
	}
}

func TestKarenSignature(t *testing.T) {
	content := Parse(karenMail)
	expected := "Hi this is my email"
	if content != expected {
		t.Errorf("expected: %v but is %v", expected, content)
	}
}

const karenMail = `
Hi this is my email

Karen The Green
Graphic Designer
Office
Tel: +44423423423423
Fax: +44234234234234
karen@webby.com
Street 2, City, Zeeland, 4694EG, NL
www.thing.com

Lorem Ipsum is simply dummy text of the printing and typesetting industry. Lorem Ipsum has been the industry's standard dummy text ever since the 1500s, when an unknown printer took a galley of type and scrambled it to make a type specimen book. It has survived not only five centuries, but also the leap into electronic typesetting, remaining essentially unchanged.

> Steps 0-2 are in prod. Gonna let them sit for a bit then start cleaning up
> the old code with 3 & 4.
>
>
`

func TestRichardSignature(t *testing.T) {
	content := Parse(richardMail)
	expected := ":+1:"
	if content != expected {
		t.Errorf("expected: %v but is %v", expected, content)
	}
}

const richardMail = `
:+1:

*Richard Lindhout* | *Eigenaar*
*Bel mij +31 6 22 22 22 22* <+31622222222>

[image: logo webridge]

Beatrixlaan 2, 4694EG Scherpenisse

*KVK      50000000*
*BTW     NL0000000AA0*
*IBAN    NL93 BUNQ 0000 1111 22*
*WEB     webRidge.nl <https://webridge.nl/>*
`

func TestAllKindOfCombinations(t *testing.T) {
	var howManyCombinations int
	var howManySuccess int

	mailErr := filepath.Walk("./dataset/basemail/",
		func(mailPath string, mailFile os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if mailFile.IsDir() {
				return nil
			}

			fmt.Println(mailPath)
			//
			mailContent, err := ioutil.ReadFile(filepath.Join(mailPath))
			if err != nil {
				return err
			}

			signatureErr := filepath.Walk("./dataset/signatures/",
				func(signaturePath string, signatureFile os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if signatureFile.IsDir() {
						return nil
					}

					signatureContent, err := ioutil.ReadFile(signaturePath)
					if err != nil {
						return err
					}

					quotedReplyErr := filepath.Walk("./dataset/quoted_reply/",
						func(quotedReplyPath string, quotedReplyFile os.FileInfo, err error) error {
							if err != nil {
								return err
							}
							if quotedReplyFile.IsDir() {
								return nil
							}

							howManyCombinations++

							quotedReplyContent, err := ioutil.ReadFile(quotedReplyPath)
							if err != nil {
								return err
							}

							expectedContent := string(mailContent)

							parsed := Parse(
								expectedContent + "\n" +
									string(signatureContent) + "\n" +
									string(quotedReplyContent),
							)
							if parsed != expectedContent {
								t.Errorf(`mail: %v, reply: %v, signature: %v: expected
								
								"%v"
								
								but got
								
								"%v"
								
								`,
									mailFile.Name(),
									quotedReplyFile.Name(),
									signatureFile.Name(),
									expectedContent,
									parsed,
								)
							} else {
								howManySuccess++
							}
							return nil
						})

					// TODO: signature first
					// parsed := Parse(string(signatureContent) +"\n" +string(mailContent))
					if quotedReplyErr != nil {
						return quotedReplyErr
					}
					return nil
				})
			if signatureErr != nil {
				return signatureErr
			}
			return err
		})

	if mailErr != nil {
		t.Error(mailErr)
	}

	t.Logf("%v/%v were successfully parsed", howManySuccess, howManyCombinations)
}
