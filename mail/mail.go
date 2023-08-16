package mail

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/mail"
	"strings"
	"time"

	"github.com/regnull/easyecc/v2"
	"github.com/regnull/ubikom/bc"
	"github.com/rs/zerolog/log"
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

// ExtractSubject returns the message subject.
func ExtractSubject(content string) (subject string, err error) {
	contentReader := strings.NewReader(content)
	mailMsg, err := mail.ReadMessage(contentReader)
	if err != nil {
		return
	}
	subject = mailMsg.Header.Get("Subject")
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

func ExtractReceiverInternalNames(content string) ([]string, error) {
	contentReader := strings.NewReader(content)
	mailMsg, err := mail.ReadMessage(contentReader)
	if err != nil {
		return nil, err
	}
	// Check for forwarded email first.
	var receivers []string
	addressStr := mailMsg.Header.Get("X-Forwarded-To")
	if addressStr != "" {
		// This email was forwarded.
		receivers = append(receivers, strings.Split(addressStr, ",")...)
		log.Debug().Msg("this message was forwarded")
	} else {
		addressStr = mailMsg.Header.Get("To")
		if addressStr != "" {
			receivers = append(receivers, strings.Split(addressStr, ",")...)
		}

		addressStr = mailMsg.Header.Get("Cc")
		if addressStr != "" {
			receivers = append(receivers, strings.Split(addressStr, ",")...)
		}

		addressStr = mailMsg.Header.Get("Bcc")
		if addressStr != "" {
			receivers = append(receivers, strings.Split(addressStr, ",")...)
		}
	}
	log.Debug().Interface("receivers", receivers).Msg("got receivers")

	var receiver []string
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

// RewriteInternalAddresses rewrites the message to change sender from internal to external format.
func RewriteInternalAddresses(message string, header string) (rewrittenMessage string, err error) {
	contentReader := strings.NewReader(message)
	mailMsg, err := mail.ReadMessage(contentReader)
	if err != nil {
		return
	}

	headerVal := mailMsg.Header.Get(header)
	if headerVal == "" {
		rewrittenMessage = message
		return
	}
	var externalAddresses []string
	for _, to := range strings.Split(headerVal, ",") {
		_, fullAddr, err := InternalToExternalAddress(to)
		if err != nil {
			return "", err
		}
		externalAddresses = append(externalAddresses, fullAddr)
	}

	var buf bytes.Buffer
	buf.Write([]byte(fmt.Sprintf("%s: %s\n", header, strings.Join(externalAddresses, ","))))
	for name, values := range mailMsg.Header {
		if name == header {
			continue
		}
		for _, value := range values {
			buf.Write([]byte(fmt.Sprintf("%s: %s\n", name, value)))
		}
	}
	buf.Write([]byte("\n"))
	io.Copy(&buf, mailMsg.Body)
	rewrittenMessage = buf.String()
	return
}

// ExtractAddresses extracts short email addresses from the email.
func ExtractAddresses(message string, header string) ([]string, error) {
	contentReader := strings.NewReader(message)
	mailMsg, err := mail.ReadMessage(contentReader)
	if err != nil {
		return nil, err
	}

	aa, err := mailMsg.Header.AddressList(header)
	if err != nil {
		return nil, err
	}
	var addresses []string
	for _, addr := range aa {
		addresses = append(addresses, addr.Address)
	}
	return addresses, nil
}

func AddReceivedHeader(message string, header []string) (string, error) {
	if len(header) == 0 {
		return "", fmt.Errorf("invalid header")
	}

	contentReader := strings.NewReader(message)
	mailMsg, err := mail.ReadMessage(contentReader)
	if err != nil {
		return "", err
	}

	var lines []string
	for i, h := range header {
		line := ""
		if i == 0 {
			line = "Received: "
		} else {
			line = "    "
		}
		line += h
		if i != len(header)-1 {
			line += "\n"
		}
		lines = append(lines, line)
	}
	lines[len(lines)-1] += fmt.Sprintf("; %s", time.Now().Format("02 Jan 06 15:04:05 -0700\n"))

	var buf bytes.Buffer
	for _, line := range lines {
		buf.Write([]byte(line))
	}
	for name, values := range mailMsg.Header {
		for _, value := range values {
			buf.Write([]byte(fmt.Sprintf("%s: %s\n", name, value)))
		}
	}
	buf.Write([]byte("\n"))
	io.Copy(&buf, mailMsg.Body)
	return buf.String(), nil
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

func NewMessage(to, from string, subject string, body string) string {
	return fmt.Sprintf(`To: %s@ubikom.cc
From: %s@ubikom.cc
Subject: %s
Date: %s
Content-Type: text/plain; charset=utf-8; format=flowed
Content-Transfer-Encoding: 7bit
Content-Language: en-US

%s
`, to, from, subject, time.Now().Format("02 Jan 06 15:04:05 -0700"), body)
}

func AddUbikomHeaders(ctx context.Context, body string, sender, receiver string,
	senderKey *easyecc.PublicKey, bchain bc.Blockchain) (string, error) {
	// Get receiver's public key.
	receiverKey, err := bchain.PublicKey(ctx, receiver)
	if err != nil {
		return "", fmt.Errorf("failed to get receiver public key: %w", err)
	}
	senderBitcoinAddress, _ := senderKey.BitcoinAddress()
	receiverBitcoinAddress, _ := receiverKey.BitcoinAddress()
	headers := map[string]string{
		"X-Ubikom-Sender":       sender,
		"X-Ubikom-Sender-Key":   senderBitcoinAddress,
		"X-Ubikom-Receiver":     receiver,
		"X-Ubikom-Receiver-Key": receiverBitcoinAddress}
	withHeaders := AddHeaders(string(body), headers)
	return withHeaders, nil
}
