package storage

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"testing"
	"time"

	"github.com/axllent/mailpit/config"
	"github.com/jhillyerd/enmime"
)

var (
	testTextEmail []byte
	testMimeEmail []byte
)

func TestTextEmailInserts(t *testing.T) {
	setup(false)
	t.Log("Testing memory storage")

RepeatTest:
	start := time.Now()
	for i := 0; i < 1000; i++ {
		if _, err := Store(DefaultMailbox, testTextEmail); err != nil {
			t.Log("error ", err)
			t.Fail()
		}
	}

	count, err := Count(DefaultMailbox)
	if err != nil {
		t.Log("error ", err)
		t.Fail()
	}

	assertEqual(t, count, 1000, "incorrect number of text emails stored")

	t.Logf("inserted 1,000 text emails in %s\n", time.Since(start))

	delStart := time.Now()
	if err := DeleteAllMessages(DefaultMailbox); err != nil {
		t.Log("error ", err)
		t.Fail()
	}

	count, err = Count(DefaultMailbox)
	if err != nil {
		t.Log("error ", err)
		t.Fail()
	}

	assertEqual(t, count, 0, "incorrect number of text emails deleted")

	t.Logf("deleted 1,000 text emails in %s\n", time.Since(delStart))

	db.Close()
	if config.DataDir == "" {
		setup(true)
		t.Logf("Testing physical storage to %s", config.DataDir)
		defer os.RemoveAll(config.DataDir)
		goto RepeatTest
	}

}

func TestMimeEmailInserts(t *testing.T) {
	setup(false)
	t.Log("Testing memory storage")

RepeatTest:
	start := time.Now()
	for i := 0; i < 1000; i++ {
		if _, err := Store(DefaultMailbox, testMimeEmail); err != nil {
			t.Log("error ", err)
			t.Fail()
		}
	}

	count, err := Count(DefaultMailbox)
	if err != nil {
		t.Log("error ", err)
		t.Fail()
	}

	assertEqual(t, count, 1000, "incorrect number of mime emails stored")

	t.Logf("inserted 1,000 emails with mime attachments in %s\n", time.Since(start))

	delStart := time.Now()
	if err := DeleteAllMessages(DefaultMailbox); err != nil {
		t.Log("error ", err)
		t.Fail()
	}

	count, err = Count(DefaultMailbox)
	if err != nil {
		t.Log("error ", err)
		t.Fail()
	}

	assertEqual(t, count, 0, "incorrect number of mime emails deleted")

	t.Logf("deleted 1,000 mime emails in %s\n", time.Since(delStart))

	db.Close()
	if config.DataDir == "" {
		setup(true)
		t.Logf("Testing physical storage to %s", config.DataDir)
		defer os.RemoveAll(config.DataDir)
		goto RepeatTest
	}
}

func TestRetrieveMimeEmail(t *testing.T) {
	setup(false)
	t.Log("Testing memory storage")

RepeatTest:
	id, err := Store(DefaultMailbox, testMimeEmail)
	if err != nil {
		t.Log("error ", err)
		t.Fail()
	}

	msg, err := GetMessage(DefaultMailbox, id)
	if err != nil {
		t.Log("error ", err)
		t.Fail()
	}

	assertEqual(t, msg.From.Name, "Sender Smith", "\"From\" name does not match")
	assertEqual(t, msg.From.Address, "sender@example.com", "\"From\" address does not match")
	assertEqual(t, msg.Subject, "inline + attachment", "subject does not match")
	assertEqual(t, len(msg.To), 1, "incorrect number of recipients")
	assertEqual(t, msg.To[0].Name, "Recipient Ross", "\"To\" name does not match")
	assertEqual(t, msg.To[0].Address, "recipient@example.com", "\"To\" address does not match")
	assertEqual(t, len(msg.Attachments), 1, "incorrect number of attachments")
	assertEqual(t, msg.Attachments[0].FileName, "Sample PDF.pdf", "attachment filename does not match")
	assertEqual(t, len(msg.Inline), 1, "incorrect number of inline attachments")
	assertEqual(t, msg.Inline[0].FileName, "inline-image.jpg", "inline attachment filename does not match")
	attachmentData, err := GetAttachmentPart(DefaultMailbox, id, msg.Attachments[0].PartID)
	assertEqual(t, len(attachmentData.Content), msg.Attachments[0].Size, "attachment size does not match")
	inlineData, err := GetAttachmentPart(DefaultMailbox, id, msg.Inline[0].PartID)
	assertEqual(t, len(inlineData.Content), msg.Inline[0].Size, "inline attachment size does not match")

	db.Close()

	if config.DataDir == "" {
		setup(true)
		t.Logf("Testing physical storage to %s", config.DataDir)
		defer os.RemoveAll(config.DataDir)
		goto RepeatTest
	}
}

func TestSearch(t *testing.T) {
	setup(false)
	t.Log("Testing memory storage")

RepeatTest:
	for i := 0; i < 1000; i++ {
		msg := enmime.Builder().
			From(fmt.Sprintf("From %d", i), fmt.Sprintf("from-%d@example.com", i)).
			Subject(fmt.Sprintf("Subject line %d end", i)).
			Text([]byte(fmt.Sprintf("This is the email body %d <jdsauk;dwqmdqw;>.", i))).
			To(fmt.Sprintf("To %d", i), fmt.Sprintf("to-%d@example.com", i))

		env, err := msg.Build()
		if err != nil {
			t.Log("error ", err)
			t.Fail()
		}

		buf := new(bytes.Buffer)

		if err := env.Encode(buf); err != nil {
			t.Log("error ", err)
			t.Fail()
		}

		if _, err := Store(DefaultMailbox, buf.Bytes()); err != nil {
			t.Log("error ", err)
			t.Fail()
		}
	}

	for i := 1; i < 101; i++ {
		// search a random something that will return a single result
		searchIndx := rand.Intn(4) + 1
		var search string
		switch searchIndx {
		case 1:
			search = fmt.Sprintf("from-%d@example.com", i)
		case 2:
			search = fmt.Sprintf("to-%d@example.com", i)
		case 3:
			search = fmt.Sprintf("Subject line %d end", i)
		default:
			search = fmt.Sprintf("the email body %d jdsauk dwqmdqw", i)
		}

		summaries, err := Search(DefaultMailbox, search, 0, 200)
		if err != nil {
			t.Log("error ", err)
			t.Fail()
		}

		assertEqual(t, len(summaries), 1, "1 search result expected")

		assertEqual(t, summaries[0].From.Name, fmt.Sprintf("From %d", i), "\"From\" name does not match")
		assertEqual(t, summaries[0].From.Address, fmt.Sprintf("from-%d@example.com", i), "\"From\" address does not match")
		assertEqual(t, summaries[0].To[0].Name, fmt.Sprintf("To %d", i), "\"To\" name does not match")
		assertEqual(t, summaries[0].To[0].Address, fmt.Sprintf("to-%d@example.com", i), "\"To\" address does not match")
		assertEqual(t, summaries[0].Subject, fmt.Sprintf("Subject line %d end", i), "\"Subject\" does not match")
	}

	// search something that will return 200 rsults
	summaries, err := Search(DefaultMailbox, "This is the email body", 0, 200)
	if err != nil {
		t.Log("error ", err)
		t.Fail()
	}
	assertEqual(t, len(summaries), 200, "200 search results expected")

	db.Close()

	if config.DataDir == "" {
		setup(true)
		t.Logf("Testing physical storage to %s", config.DataDir)
		defer os.RemoveAll(config.DataDir)
		goto RepeatTest
	}
}

func BenchmarkImportText(b *testing.B) {
	setup(false)

	for i := 0; i < b.N; i++ {
		if _, err := Store(DefaultMailbox, testTextEmail); err != nil {
			b.Log("error ", err)
			b.Fail()
		}
	}

	db.Close()
}

func BenchmarkImportMime(b *testing.B) {
	setup(false)

	for i := 0; i < b.N; i++ {
		if _, err := Store(DefaultMailbox, testMimeEmail); err != nil {
			b.Log("error ", err)
			b.Fail()
		}
	}
	db.Close()
}

func setup(dataDir bool) {
	config.NoLogging = true
	config.MaxMessages = 0

	if dataDir {
		config.DataDir = fmt.Sprintf("%s-%d", path.Join(os.TempDir(), "mailpit-tests"), time.Now().UnixNano())
	} else {
		config.DataDir = ""
	}

	if err := InitDB(); err != nil {
		panic(err)
	}

	var err error

	testTextEmail, err = ioutil.ReadFile("testdata/plain-text.eml")
	if err != nil {
		panic(err)
	}

	testMimeEmail, err = ioutil.ReadFile("testdata/mime-attachment.eml")
	if err != nil {
		panic(err)
	}
}

func assertEqual(t *testing.T, a interface{}, b interface{}, message string) {
	if a == b {
		return
	}
	message = fmt.Sprintf("%s: \"%v\" != \"%v\"", message, a, b)
	t.Fatal(message)
}
