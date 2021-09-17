package mail

import (
	"bytes"
	"fmt"
	"io"
	"net/mail"
	"strings"
)

const (
	ubikomShortSuffix = "@x"
	ubikomLongSuffix  = "@ubikom.cc"
)

// InternalToExternalAddress parses the address and returns the short and full address suitable
// for sending messages to the external destinations.
func InternalToExternalAddress(address string) (shortAddr, fullAddr string, err error) {
	mailAddr, err := mail.ParseAddress(address)
	if err != nil {
		return
	}

	shortAddr = mailAddr.Address
	shortAddr = strings.Replace(shortAddr, ubikomShortSuffix, ubikomLongSuffix, 1)
	fullAddr = fmt.Sprintf("%s <%s>", mailAddr.Name, shortAddr)
	return
}

// ExtractSenderAddress returns full sender's address extracted from the content.
func ExtractSenderAddress(content string) (address string, err error) {
	contentReader := strings.NewReader(content)
	mailMsg, err := mail.ReadMessage(contentReader)
	if err != nil {
		return
	}
	address = mailMsg.Header.Get("From")
	return
}

func ExtractReceiverInternalName(content string) (receiver string, err error) {
	contentReader := strings.NewReader(content)
	mailMsg, err := mail.ReadMessage(contentReader)
	if err != nil {
		return
	}
	addressStr := mailMsg.Header.Get("To")
	address, err := mail.ParseAddress(addressStr)
	if err != nil {
		return
	}
	receiver = strings.Replace(address.Address, ubikomLongSuffix, "", 1)
	return
}

func ExtractReceiverInternalNames(content string) (receiver []string, err error) {
	contentReader := strings.NewReader(content)
	mailMsg, err := mail.ReadMessage(contentReader)
	if err != nil {
		return nil, err
	}
	addressStr := mailMsg.Header.Get("To")
	receivers := strings.Split(addressStr, ",")
	for _, r := range receivers {
		var address *mail.Address
		address, err = mail.ParseAddress(r)
		if err != nil {
			return nil, err
		}
		if !strings.HasSuffix(address.Address, ubikomLongSuffix) {
			// This is not a message for us.
			continue
		}
		receiver = append(receiver, strings.Replace(address.Address, ubikomLongSuffix, "", 1))
	}
	return receiver, nil
}

// RewriteFromHeader rewrites the message to change sender from internal to external format.
func RewriteFromHeader(message string) (rewrittenMessage string, fromAddr string, toAddr []string, err error) {
	contentReader := strings.NewReader(message)
	mailMsg, err := mail.ReadMessage(contentReader)
	if err != nil {
		return
	}

	from := mailMsg.Header.Get("From")
	fromAddr, fullAddr, err := InternalToExternalAddress(from)
	if err != nil {
		return
	}

	// Remove internal addresses from the list of recipients.
	var externalAddresses []string
	for _, to := range strings.Split(mailMsg.Header.Get("To"), ",") {
		a, err := mail.ParseAddress(to)
		if err != nil {
			// Invalid address.
			return "", "", nil, err
		}
		if !IsInternal(a.Address) {
			externalAddresses = append(externalAddresses, to)
			toAddr = append(toAddr, a.Address)
		}
	}

	var buf bytes.Buffer
	buf.Write([]byte(fmt.Sprintf("To: %s\n", strings.Join(externalAddresses, ","))))
	buf.Write([]byte(fmt.Sprintf("From: %s\n", fullAddr)))
	for name, values := range mailMsg.Header {
		if name == "From" {
			continue
		}
		if name == "To" {
			continue
		}
		for _, value := range values {
			buf.Write([]byte(fmt.Sprintf("%s: %s\n", name, value)))
		}
	}
	buf.Write([]byte("\n"))
	io.Copy(&buf, mailMsg.Body)
	rewrittenMessage = string(buf.Bytes())
	return
}

// StripDomain removes the domain name from name.
func StripDomain(name string) string {
	i := strings.Index(name, "@")
	if i == -1 {
		return name
	}
	return name[:i]
}

// IsInternal returns true if the name refers to an internal recipient.
func IsInternal(name string) bool {
	i := strings.Index(name, "@")
	if i == -1 {
		return true
	}
	if strings.HasSuffix(name, ubikomShortSuffix) {
		return true
	}
	if strings.HasSuffix(name, ubikomLongSuffix) {
		return true
	}
	return false
}

// AddHeaders adds headers to a message.
func AddHeaders(content string, headers map[string]string) string {
	lines := strings.Split(content, "\n")
	var newLines []string

	for i, line := range lines {
		line = strings.Replace(line, "\r", "", -1)
		if line == "" {
			// This line separates headers from the body.
			for name, value := range headers {
				headerLine := name + ": " + value
				newLines = append(newLines, headerLine)
			}
			newLines = append(newLines, lines[i:]...)
			return strings.Join(newLines, "\n")
		}
		newLines = append(newLines, line)
	}
	// This means the body was not found, weird.
	return strings.Join(newLines, "\n")
}
