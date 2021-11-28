package imap

import (
	"bufio"
	"bytes"
	"io"
	"strings"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/backend/backendutil"
	"github.com/emersion/go-message"
	"github.com/emersion/go-message/textproto"
	"github.com/regnull/ubikom/pb"
	"github.com/regnull/ubikom/util"
	"github.com/rs/zerolog/log"
)

type Message struct {
	Uid   uint32
	Date  time.Time
	Size  uint32
	Flags []string
	Body  []byte
}

func NewMessageFromProto(m *pb.ImapMessage) *Message {
	return &Message{
		Uid:   m.GetUid(),
		Date:  util.TimeFromMs(int64(m.GetReceivedTimestamp())),
		Size:  uint32(m.GetSize()),
		Flags: m.GetFlag(),
		Body:  m.GetContent()}
}

func (m *Message) ToProto() *pb.ImapMessage {
	return &pb.ImapMessage{
		Content:           m.Body,
		Flag:              m.Flags,
		ReceivedTimestamp: uint64(m.Date.UnixNano() / 1000000),
		Size:              uint64(m.Size),
		Uid:               m.Uid,
	}
}

func filterMalformedHeaders(body []byte) []byte {
	bodyStr := string(body)
	lines := strings.Split(bodyStr, "\n")
	var newLines []string
	headers := true
	for _, line := range lines {
		if headers && (line == "" || line == "\r") {
			// Done with headers.
			headers = false
		}
		if headers &&
			(strings.HasPrefix(line, ">From") || strings.HasPrefix(line, "From") &&
				!strings.HasPrefix(line, "From:")) {
			continue
		}
		newLines = append(newLines, line)
	}
	newBody := strings.Join(newLines, "\n")
	return []byte(newBody)
}

func (m *Message) entity() (*message.Entity, error) {
	return message.Read(bytes.NewReader(filterMalformedHeaders(m.Body)))
}

func (m *Message) headerAndBody() (textproto.Header, io.Reader, error) {
	body := bufio.NewReader(bytes.NewReader(filterMalformedHeaders(m.Body)))
	hdr, err := textproto.ReadHeader(body)
	return hdr, body, err
}

func (m *Message) Fetch(seqNum uint32, items []imap.FetchItem) (*imap.Message, error) {
	fetched := imap.NewMessage(seqNum, items)
	for _, item := range items {
		switch item {
		case imap.FetchEnvelope:
			hdr, _, err := m.headerAndBody()
			if err != nil {
				return nil, err
			}
			fetched.Envelope, _ = backendutil.FetchEnvelope(hdr)
		case imap.FetchBody, imap.FetchBodyStructure:
			hdr, body, err := m.headerAndBody()
			if err != nil {
				return nil, err
			}
			fetched.BodyStructure, _ = backendutil.FetchBodyStructure(hdr, body, item == imap.FetchBodyStructure)
		case imap.FetchFlags:
			fetched.Flags = m.Flags
		case imap.FetchInternalDate:
			fetched.InternalDate = m.Date
		case imap.FetchRFC822Size:
			fetched.Size = m.Size
		case imap.FetchUid:
			fetched.Uid = m.Uid
		default:
			section, err := imap.ParseBodySectionName(item)
			if err != nil {
				break
			}

			body := bufio.NewReader(bytes.NewReader(filterMalformedHeaders(m.Body)))
			hdr, err := textproto.ReadHeader(body)
			if err != nil {
				return nil, err
			}

			l, err := backendutil.FetchBodySection(hdr, body, section)
			if err != nil {
				return nil, err
			}
			fetched.Body[section] = l
		}
	}

	return fetched, nil
}

func (m *Message) Match(seqNum uint32, c *imap.SearchCriteria) (bool, error) {
	e, err := m.entity()
	if err != nil {
		log.Error().Err(err).Msg("failed to create email entity")
		return false, err
	}
	return backendutil.Match(e, seqNum, m.Uid, m.Date, m.Flags, c)
}
