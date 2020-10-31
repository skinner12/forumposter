package forumposter

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"strings"

	"github.com/PuerkitoBio/goquery"
	log "github.com/sirupsen/logrus"
)

//ForumophiliaInfoSite provides the info to make https request for post
type ForumophiliaInfoSite struct {
	URL      string
	User     string
	Password string
	F        string // Forum number
	T        string // Thread Number

}

//forumophilia manage forumophilia forum
func (c *Collector) forumophilia(i ForumophiliaInfoSite, p Payload) error {

	// Load home page to get SID from cookie
	initialLoad := &Request{
		Body:   nil,
		URL:    fmt.Sprintf("%s/", i.URL),
		Method: "GET",
		Writer: nil,
	}

	_, err := c.fetch(initialLoad)
	if err != nil {
		return err
	}

	log.Debugln("SID", c.Sid)

	// Make LOGIN

	/*
			{
			"username": "aaa",
			"password": "sss",
			"redirect": "",
			"login": "Log+in"
		}
	*/
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	_ = writer.WriteField("username", i.User)
	_ = writer.WriteField("password", i.Password)
	_ = writer.WriteField("redirect", "")
	_ = writer.WriteField("login", "Log+in")
	err = writer.Close()
	if err != nil {
		log.Debugf("[Forum-Poster] Login - %v", err)
		return fmt.Errorf("[Forum-Poster] Login - %v", err)
	}

	postLogin := &Request{
		Body:   payload,
		URL:    fmt.Sprintf("%s/login.php", i.URL),
		Method: "POST",
		Writer: writer,
	}

	_, err = c.fetch(postLogin)
	if err != nil {
		return err
	}

	return nil
}

//ForumophiliaPost post new thread
//a is for chose if reply or new thread
func (c *Collector) ForumophiliaPost(i ForumophiliaInfoSite, p Payload, a string) (string, error) {

	var url string
	var mode string

	// Set post NEW or REPLY
	switch a {
	case "new":
		log.Infoln("* Post new thread to", i.URL)
		url = fmt.Sprintf("%s/posting.php?mode=newtopic&f=%s", i.URL, i.F)
		mode = "newtopic"
	case "reply":
		log.Infoln("* Reply thread to", i.URL)
		url = fmt.Sprintf("%s/posting.php?mode=reply&f=%s&t=%s", i.URL, i.F, i.T)
		mode = "reply"
	default:
		return "", fmt.Errorf("[Forum-Poster] - Choice are: new or reply. Set the right one")
	}

	// Login first
	err := c.forumophilia(i, p)
	if err != nil {
		return "", err
	}

	log.Debugln("SID", c.Sid)

	// Post New Thread

	// post new
	/*{
		"subject": "Hello",
		"addbbcode20": "#444444",
		"addbbcode22": "0",
		"message": "Hi",
		"poll_title": "",
		"add_poll_option_text": "",
		"poll_length": "",
		"mode": "newtopic",
		"sid": "ed5ed0765dacf197f29a3442a0688b81",
		"f": "9",
		"post": "Submit"
	}*/

	// post reply
	/*{
		"subject": "",
		"addbbcode20": "#444444",
		"addbbcode22": "0",
		"helpbox": "Tip:+Styles+can+be+applied+quickly+to+selected+text.",
		"message": "too+muchhhh",
		"mode": "reply",
		"sid": "ed5ed0765dacf197f29a3442a0688b81",
		"t": "25689",
		"post": "Submit"
	}*/
	postload := &bytes.Buffer{}
	writerLoad := multipart.NewWriter(postload)
	_ = writerLoad.WriteField("sid", c.Sid)
	_ = writerLoad.WriteField("subject", p.Title)
	_ = writerLoad.WriteField("message", p.Message)
	_ = writerLoad.WriteField("post", "Submit")
	_ = writerLoad.WriteField("mode", mode)
	_ = writerLoad.WriteField("addbbcode20", "#444444")
	_ = writerLoad.WriteField("addbbcode22", "0")

	switch a {
	case "new":
		_ = writerLoad.WriteField("poll_title", "")
		_ = writerLoad.WriteField("add_poll_option_text", "")
		_ = writerLoad.WriteField("poll_length", "")
		_ = writerLoad.WriteField("f", i.F)
		_ = writerLoad.WriteField("subject", p.Title)
	case "reply":
		_ = writerLoad.WriteField("t", i.T)
		_ = writerLoad.WriteField("helpbox", "Tip:+Styles+can+be+applied+quickly+to+selected+text.")
		_ = writerLoad.WriteField("subject", "")
	default:
		return "", fmt.Errorf("[Forum-Poster] - Choice are: new or reply. Set the right one")
	}

	err = writerLoad.Close()
	if err != nil {
		fmt.Println(err)
		return "", fmt.Errorf("[Forum-Poster] Post - %v", err)
	}

	log.Debugln("Posting to FORUM ID", i.F)

	postThread := &Request{
		Body:   postload,
		URL:    url,
		Method: "POST",
		Writer: writerLoad,
	}

	resp, err := c.fetch(postThread)
	if err != nil {
		return "", err
	}

	log.Traceln("[Forum-Poster] Response:", string(resp))

	if !checkFinalURL(c.FinalURL) {
		return "", fmt.Errorf("[Forum-Poster] NOT Posted - %s", c.FinalURL)
	}

	// Load the HTML document
	log.Debugln("Extracting Value")
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(resp)))
	if err != nil {
		log.Fatal(err)
	}

	// check if login is present into response
	title := doc.Find("title").Text()

	_, ok := doc.Find("input[name='username']").Attr("value")
	if ok {

		log.WithFields(log.Fields{
			"Title": title,
			"SID":   c.Sid,
			"URL":   url,
		}).Error("[Forum-Poster] - Extract Values")
		return "", fmt.Errorf("[Forum-Poster] - Present login form, release not posted! Get this web's title %s", title)
	}

	switch a {
	case "new":
		log.Infof("The URL of new thread is: %v\n", c.FinalURL)
	case "reply":
		log.Infof("The URL of reply is: %v\n", c.FinalURL)
	}

	return c.FinalURL, nil
}
