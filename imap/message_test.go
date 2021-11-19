package imap

import (
	"io/ioutil"
	"testing"

	"github.com/emersion/go-imap"
	"github.com/regnull/ubikom/mail"
	"github.com/regnull/ubikom/pb"
	"github.com/regnull/ubikom/util"
	"github.com/stretchr/testify/assert"
)

func Test_NewMessageFromProto(t *testing.T) {
	assert := assert.New(t)

	p := &pb.ImapMessage{
		Content:           []byte("content"),
		Flag:              []string{"someFlag"},
		ReceivedTimestamp: 12345,
		Size:              555,
		Uid:               1001,
	}
	m := NewMessageFromProto(p)
	assert.EqualValues("content", m.Body)
	assert.EqualValues([]string{"someFlag"}, m.Flags)
	assert.EqualValues(12345, m.Date.UnixNano()/1000000)
	assert.EqualValues(555, m.Size)
	assert.EqualValues(1001, m.Uid)
}

func Test_MessageToProto(t *testing.T) {
	assert := assert.New(t)

	p := &pb.ImapMessage{
		Content:           []byte("content"),
		Flag:              []string{"someFlag"},
		ReceivedTimestamp: 12345,
		Size:              555,
		Uid:               1001,
	}
	m := NewMessageFromProto(p)
	p1 := m.ToProto()
	assert.Equal(p, p1)
}

func Test_Fetch(t *testing.T) {
	assert := assert.New(t)

	msgBody := mail.NewMessage("foo", "bar", "yo", "here's my message")
	p := &pb.ImapMessage{
		Content:           []byte(msgBody),
		Flag:              []string{"flag1", "flag2"},
		ReceivedTimestamp: uint64(util.NowMs()),
		Size:              1000,
		Uid:               1001,
	}
	m := NewMessageFromProto(p)

	i, err := m.Fetch(1, []imap.FetchItem{imap.FetchEnvelope})
	assert.NoError(err)
	assert.EqualValues("bar", i.Envelope.From[0].MailboxName)
	assert.EqualValues("foo", i.Envelope.To[0].MailboxName)
	assert.EqualValues("yo", i.Envelope.Subject)

	i, err = m.Fetch(1, []imap.FetchItem{imap.FetchBody})
	assert.NoError(err)
	// TODO: See what we can verify here.
	assert.NotNil(i)

	i, err = m.Fetch(1, []imap.FetchItem{imap.FetchFlags})
	assert.NoError(err)
	assert.EqualValues(2, len(i.Flags))
	assert.Contains(i.Flags, "flag1")
	assert.Contains(i.Flags, "flag2")

	i, err = m.Fetch(1, []imap.FetchItem{imap.FetchInternalDate})
	assert.NoError(err)
	assert.EqualValues(p.ReceivedTimestamp, i.InternalDate.UnixMilli())

	i, err = m.Fetch(1, []imap.FetchItem{imap.FetchRFC822Size})
	assert.NoError(err)
	assert.EqualValues(p.Size, i.Size)

	i, err = m.Fetch(1, []imap.FetchItem{imap.FetchUid})
	assert.NoError(err)
	assert.EqualValues(p.Uid, i.Uid)

	i, err = m.Fetch(1, []imap.FetchItem{imap.FetchRFC822Text})
	assert.NoError(err)
	assert.NoError(err)
	for _, v := range i.Body {
		data, err := ioutil.ReadAll(v)
		assert.NoError(err)
		// Looks like it adds a new line.
		assert.EqualValues([]byte("here's my message\n"), data)
		break
	}
}

func Test_MessageMatch(t *testing.T) {
	assert := assert.New(t)

	msgBody := mail.NewMessage("foo", "bar", "yo", "here's my message")
	p := &pb.ImapMessage{
		Content:           []byte(msgBody),
		Flag:              []string{"flag1", "flag2"},
		ReceivedTimestamp: uint64(util.NowMs()),
		Size:              1000,
		Uid:               1001,
	}
	m := NewMessageFromProto(p)

	// Match message by UID.
	ss := new(imap.SeqSet)
	ss.Add("1001")
	b, err := m.Match(0, &imap.SearchCriteria{Uid: ss})
	assert.NoError(err)
	assert.True(b)
}
