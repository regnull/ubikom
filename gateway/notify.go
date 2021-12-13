package gateway

import (
	"bytes"
	"context"
	"text/template"

	"github.com/regnull/easyecc"
	"github.com/regnull/ubikom/mail"
	"github.com/regnull/ubikom/pb"
	"github.com/regnull/ubikom/protoutil"
	"github.com/rs/zerolog/log"
)

const messageTmpl = `
Greetings from your friendly Ubikom Gateway!

Regretfully, I had to block your message from being sent externally. 
This happened because you have exceeded your message per hour rate limit.

Originating user: {{.OriginatingUser}} 

Blocked message intended recipient: {{.Recipient}} 

Subject: {{.Subject}}

Reducing spam (and hopefully eliminating it completely) is one of our main goals.
If you are sending unsolicited messages to the good people of the Internet,
please don't. Maybe you just have a lot of friends you are corresponding with.
This it totally cool. If this is the case, you can try to space out your messages
through the day, instead of sending them all together. Or, if there is a legitimate
case for you to send a large message volume, please contact us (lgx@ubikom.cc) 
and we can figure something out. 

Yours truly,

Ubikom Gateway
`

const sendFailedTmpl = `
Greetings from your friendly Ubikom Gateway!

Unfortunately, there was an error sending your message. Here are some details:

Originating user: {{.OriginatingUser}} 

Recipient: {{.Recipient}} 

Subject: {{.Subject}}

Error: {{.ErrorText}}

Please check your recipient address and try again.

Yours truly,

Ubikom Gateway
`

type messageArs struct {
	OriginatingUser string
	Recipient       string
	Subject         string
	ErrorText       string
}

func NotifyMessageBlocked(ctx context.Context,
	privateKey *easyecc.PrivateKey,
	lookupClient pb.LookupServiceClient,
	notifyUser string,
	messageTo string,
	subject string) error {

	tmpl, err := template.New("notification").Parse(messageTmpl)
	if err != nil {
		return err
	}
	var b bytes.Buffer
	err = tmpl.Execute(&b, &messageArs{
		OriginatingUser: notifyUser,
		Recipient:       messageTo,
		Subject:         subject,
	})
	if err != nil {
		return err
	}
	body := mail.NewMessage(notifyUser, "gateway", "outgoing message blocked", b.String())
	log.Debug().Str("to", notifyUser).Msg("sending notification")

	// TODO: Pass gateway name instead of hardcoding.
	return protoutil.SendMessage(ctx, privateKey, []byte(body), "gateway", notifyUser, lookupClient)
}

func NotifyMessageFailedToSend(ctx context.Context,
	privateKey *easyecc.PrivateKey,
	lookupClient pb.LookupServiceClient,
	notifyUser string,
	messageTo string,
	subject string,
	errorText string) error {
	tmpl, err := template.New("notification").Parse(sendFailedTmpl)
	if err != nil {
		return err
	}
	var b bytes.Buffer
	err = tmpl.Execute(&b, &messageArs{
		OriginatingUser: notifyUser,
		Recipient:       messageTo,
		Subject:         subject,
		ErrorText:       errorText,
	})
	if err != nil {
		return err
	}
	body := mail.NewMessage(notifyUser, "gateway", "message delivery failure", b.String())
	log.Debug().Str("to", notifyUser).Msg("sending notification")

	// TODO: Pass gateway name instead of hardcoding.
	return protoutil.SendMessage(ctx, privateKey, []byte(body), "gateway", notifyUser, lookupClient)
}
