package mail

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/mr-tron/base58"
	"github.com/regnull/easyecc/v2"
	"github.com/regnull/ubikom/bc"
	bcmocks "github.com/regnull/ubikom/bc/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var testMessage = `From - Fri May 28 16:10:23 2021
X-Account-Key: account3
X-UIDL: a0670b6f6749ce00f2c3ad6777ae2f8db98c83bdbbfc6321c39f3a888a460809
X-Mozilla-Status: 0001
X-Mozilla-Status2: 00000000
X-Mozilla-Keys:                                                                                 
To: Leonid Gorkin <regnull@gmail.com>, Geonid Lorkin <nullreg@gmail.com>
From: Leonid Gorkin <lgx@x>
Subject: testing
Message-ID: <a0a78dfe-f4f4-3f41-f52a-f965071d7404@x>
Date: Fri, 28 May 2021 16:10:21 -0400
User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:78.0)
	Gecko/20100101 Thunderbird/78.10.2
MIME-Version: 1.0
Content-Type: text/plain; charset=utf-8; format=flowed
Content-Transfer-Encoding: 7bit
Content-Language: en-US

Hello	
`

var messageWithMultipleInternalRecipients = `From - Tue Jun 15 13:23:32 2021
X-Account-Key: account3
X-UIDL: 2f95ef4512a5ed4d9c2cfe0b9eda3f112e42f6304c849982cbfe3d05650bd8b5
X-Mozilla-Status: 0001
X-Mozilla-Status2: 00000000
X-Mozilla-Keys:                                                                                 
To: Leonid Gorkin <lgx@ubikom.cc>, Geonid Lorkin <glx@ubikom.cc>, Some OtherGuy <someotherguy@gmail.com>
From: Spongebob Squarepants <spongebob@bikinibottom.com>
Subject: test headers
Message-ID: <8124dc1a-7e25-d056-6797-9bf935cff444@x>
Date: Tue, 15 Jun 2021 13:23:28 -0400
User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:78.0)
 Gecko/20100101 Thunderbird/78.11.0
MIME-Version: 1.0
Content-Type: text/plain; charset=utf-8; format=flowed
Content-Transfer-Encoding: 7bit
Content-Language: en-US

Hello
`

var messageWithSingleInternalRecipient = `From - Tue Jun 15 13:23:32 2021
X-Account-Key: account3
X-UIDL: 2f95ef4512a5ed4d9c2cfe0b9eda3f112e42f6304c849982cbfe3d05650bd8b5
X-Mozilla-Status: 0001
X-Mozilla-Status2: 00000000
X-Mozilla-Keys:                                                                                 
To: Leonid Gorkin <lgx@ubikom.cc>
From: Spongebob Squarepants <spongebob@bikinibottom.com>
Subject: test headers
Message-ID: <8124dc1a-7e25-d056-6797-9bf935cff444@x>
Date: Tue, 15 Jun 2021 13:23:28 -0400
User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:78.0)
 Gecko/20100101 Thunderbird/78.11.0
MIME-Version: 1.0
Content-Type: text/plain; charset=utf-8; format=flowed
Content-Transfer-Encoding: 7bit
Content-Language: en-US

Hello
`

func Test_InternalToExternalAddress(t *testing.T) {
	assert := assert.New(t)

	shortAddr, fullAddr, err := InternalToExternalAddress("Spongebob Squarepants <spongebob@x>")
	assert.NoError(err)
	assert.EqualValues("spongebob@ubikom.cc", shortAddr)
	assert.EqualValues("Spongebob Squarepants <spongebob@ubikom.cc>", fullAddr)
}

func Test_ExtractSenderAddress(t *testing.T) {
	assert := assert.New(t)

	addr, err := ExtractSenderAddress(testMessage)
	assert.NoError(err)
	assert.EqualValues("Leonid Gorkin <lgx@x>", addr)
}

func Test_RewriteInternalAddresses(t *testing.T) {
	assert := assert.New(t)

	rewritten, err := RewriteInternalAddresses(testMessage, "From")
	assert.NoError(err)
	assert.Contains(rewritten, "lgx@ubikom.cc")
	assert.Contains(rewritten, "regnull@gmail.com")
	assert.Contains(rewritten, "nullreg@gmail.com")

	rewritten, err = RewriteInternalAddresses(testMessage, "Xyz")
	assert.NoError(err)
	assert.Equal(rewritten, testMessage)
}

func Test_StripDomain(t *testing.T) {
	assert := assert.New(t)

	assert.EqualValues("foo", StripDomain("foo"))
	assert.EqualValues("foo", StripDomain("foo@bar"))
	assert.EqualValues("foo", StripDomain("foo@bar.com"))
}

func Test_IsInternal(t *testing.T) {
	assert := assert.New(t)

	assert.True(IsInternal("foo"))
	assert.True(IsInternal("foo@x"))
	assert.True(IsInternal("foo@ubikom.cc"))

	assert.False(IsInternal("foo@gmail.com"))
	assert.False(IsInternal("foo@somewhere"))
}

func Test_AddHeaders(t *testing.T) {
	assert := assert.New(t)

	headers := map[string]string{
		"X-foo": "bar",
		"X-baz": "bazbaz",
	}
	newMessage := AddHeaders(testMessage, headers)
	assert.True(len(newMessage) > len(testMessage))
	assert.True(strings.Contains(newMessage, "X-foo:"))
	assert.True(strings.Contains(newMessage, "X-baz:"))
}

func Test_ExtractReceiverInternalNames(t *testing.T) {
	assert := assert.New(t)

	recipients, err := ExtractReceiverInternalNames(messageWithMultipleInternalRecipients)
	assert.NoError(err)
	assert.EqualValues(2, len(recipients))
	assert.Contains(recipients, "lgx")
	assert.Contains(recipients, "glx")
}

func Test_AddReceivedHeader(t *testing.T) {
	assert := assert.New(t)

	modified, err := AddReceivedHeader(testMessage, []string{"from foo.bar", "by bar.foo"})
	assert.NoError(err)
	assert.True(strings.Contains(modified, "Received: from foo.bar\n"))
	assert.True(strings.Contains(modified, "by bar.foo;"))
}

func Test_ExtractAddresses(t *testing.T) {
	assert := assert.New(t)

	addresses, err := ExtractAddresses(testMessage, "To")
	assert.NoError(err)
	assert.Contains(addresses, "regnull@gmail.com")
	assert.Contains(addresses, "nullreg@gmail.com")
}

func Test_ExtractSubject(t *testing.T) {
	assert := assert.New(t)

	subj, err := ExtractSubject(testMessage)
	assert.NoError(err)
	assert.Equal("testing", subj)

	_, err = ExtractSubject("not a valid message")
	assert.Error(err)
}

func Test_ExtractReceiverInternalName(t *testing.T) {
	assert := assert.New(t)

	rec, err := ExtractReceiverInternalName(messageWithSingleInternalRecipient)
	assert.NoError(err)
	assert.Equal("lgx", rec)

	_, err = ExtractReceiverInternalName(messageWithMultipleInternalRecipients)
	assert.Error(err)
}

func Test_NewMessage(t *testing.T) {
	assert := assert.New(t)

	msg := NewMessage("bob", "alice", "test subject", "Here's the message body")

	subj, err := ExtractSubject(msg)
	assert.NoError(err)
	assert.Equal("test subject", subj)

	rec, err := ExtractReceiverInternalName(msg)
	assert.NoError(err)
	assert.Equal("bob", rec)
}

func Test_AddUbibkomHeaders(t *testing.T) {
	assert := assert.New(t)

	ctx := context.Background()
	senderKey, err := easyecc.NewPrivateKey(easyecc.SECP256K1)
	assert.NoError(err)
	receiverKey, err := easyecc.NewPrivateKey(easyecc.SECP256K1)
	assert.NoError(err)

	receiverAddress, err := receiverKey.PublicKey().BitcoinAddress()
	assert.NoError(err)
	addressBytes, err := base58.Decode(receiverAddress)
	assert.NoError(err)
	receiverAddr := common.BytesToAddress(addressBytes)

	senderAddress, err := senderKey.PublicKey().BitcoinAddress()
	assert.NoError(err)

	caller := new(bcmocks.MockNameRegistryCaller)
	caller.EXPECT().LookupName(mock.Anything, "alice").Return(struct {
		Owner     common.Address
		PublicKey []byte
		Price     *big.Int
	}{
		Owner:     receiverAddr,
		PublicKey: receiverKey.PublicKey().CompressedBytes(),
		Price:     big.NewInt(0),
	}, nil)

	bchain := bc.NewBlockchainWithCaller(caller)

	msg, err := AddUbikomHeaders(ctx, testMessage, "bob", "alice", senderKey.PublicKey(), bchain)
	assert.NoError(err)
	assert.True(strings.Contains(msg, "X-Ubikom-Sender: bob"))
	assert.True(strings.Contains(msg, "X-Ubikom-Receiver: alice"))
	assert.True(strings.Contains(msg, fmt.Sprintf("X-Ubikom-Sender-Key: %s", senderAddress)))
	assert.True(strings.Contains(msg, fmt.Sprintf("X-Ubikom-Receiver-Key: %s", receiverAddress)))
}
