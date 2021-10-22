package mail

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
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

var anotherMessage = `From - Tue Jun 15 13:23:32 2021
X-Account-Key: account3
X-UIDL: 2f95ef4512a5ed4d9c2cfe0b9eda3f112e42f6304c849982cbfe3d05650bd8b5
X-Mozilla-Status: 0001
X-Mozilla-Status2: 00000000
X-Mozilla-Keys:                                                                                 
To: Leonid Gorkin <lgx@x>
From: Leonid Gorkin <lgx@x>
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

func Test_RewriteFromHeader(t *testing.T) {
	assert := assert.New(t)

	_, from, to, err := RewriteFromHeader(testMessage)
	assert.NoError(err)
	assert.EqualValues("lgx@ubikom.cc", from)
	assert.Contains(to, "regnull@gmail.com")
	assert.Contains(to, "nullreg@gmail.com")
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
	assert.True(strings.Index(newMessage, "X-foo:") != -1)
	assert.True(strings.Index(newMessage, "X-baz:") != -1)
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
