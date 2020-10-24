package forumposter

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	log "github.com/sirupsen/logrus"
)

//PHPBB3InfoSite provides the info to make https request for post
type PHPBB3InfoSite struct {
	URL      string
	User     string
	Password string
	F        string // Forum number
	T        string // Thread Number

}

//PHPBB3 manage phpbb v3 forum
func (c *Collector) PHPBB3(i PHPBB3InfoSite, p Payload) error {

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
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	_ = writer.WriteField("username", i.User)
	_ = writer.WriteField("password", i.Password)
	_ = writer.WriteField("sid", c.Sid)
	_ = writer.WriteField("login", "Login")
	err = writer.Close()
	if err != nil {
		log.Debugf("[Forum-Poster] Login - %v", err)
		return fmt.Errorf("[Forum-Poster] Login - %v", err)
	}

	postLogin := &Request{
		Body:   payload,
		URL:    fmt.Sprintf("%s/ucp.php?mode=login&sid=%s", i.URL, c.Sid),
		Method: "POST",
		Writer: writer,
	}

	_, err = c.fetch(postLogin)
	if err != nil {
		return err
	}

	return nil
}

//PHPBB3Post post new thread
//a is for chose if reply or new thread
func (c *Collector) PHPBB3Post(i PHPBB3InfoSite, p Payload, a string) (string, error) {

	var url string

	// Set post NEW or REPLY
	switch a {
	case "new":
		log.Infoln("* Post new thread to", i.URL)
		url = fmt.Sprintf("%s/posting.php?mode=post&f=%s", i.URL, i.F)
	case "reply":
		log.Infoln("* Reply thread to", i.URL)
		url = fmt.Sprintf("%s/posting.php?mode=reply&f=%s&t=%s", i.URL, i.F, i.T)
	default:
		return "", fmt.Errorf("[Forum-Poster] - Choice are: new or reply. Set the right one")
	}

	// Login first
	err := c.PHPBB3(i, p)
	if err != nil {
		return "", err
	}

	log.Debugln("SID", c.Sid)

	// Load home page to get SID from cookie
	readPostForm := &Request{
		Body:   nil,
		URL:    url,
		Method: "GET",
		Writer: nil,
	}

	body, err := c.fetch(readPostForm)
	if err != nil {
		return "", err
	}

	if err != nil {
		return "", err
	}

	// Load the HTML document
	log.Debugln("Extracting Value")
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		log.Fatal(err)
	}

	// Extract
	// * creation_time
	// * form_token
	// * lastclick
	title := doc.Find("title").Text()

	token, ok := doc.Find("input[name='form_token']").Attr("value")
	if !ok {

		return "", fmt.Errorf("[Forum-Poster] - Can't find form_token")
	}

	creationTime, ok := doc.Find("input[name='creation_time']").Attr("value")
	if !ok {

		return "", fmt.Errorf("[Forum-Poster] - Can't find creation_time")
	}

	lastclick, ok := doc.Find("input[name='lastclick']").Attr("value")
	if !ok {

		return "", fmt.Errorf("[Forum-Poster] - Can't find lastclick")
	}

	log.WithFields(log.Fields{
		"form_token":    token,
		"creation_time": creationTime,
		"lastclick":     lastclick,
		"Title":         title,
		"SID":           c.Sid,
		"URL":           url,
	}).Debug("[Forum-Poster] - Extract Values")

	// Find the review items
	doc.Find(".sidebar-reviews article .content-block").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the band and title
		band := s.Find("a").Text()
		title := s.Find("i").Text()
		fmt.Printf("Review %d: %s - %s\n", i, band, title)
	})

	// Post New Thread
	postload := &bytes.Buffer{}
	writerLoad := multipart.NewWriter(postload)
	_ = writerLoad.WriteField("form_token", token)
	_ = writerLoad.WriteField("creation_time", creationTime)
	_ = writerLoad.WriteField("sid", c.Sid)
	_ = writerLoad.WriteField("lastclick", lastclick)
	_ = writerLoad.WriteField("subject", p.Title)
	_ = writerLoad.WriteField("message", p.Message)
	_ = writerLoad.WriteField("post", "Submit")
	_ = writerLoad.WriteField("attach_sig", "on")
	_ = writerLoad.WriteField("topic_type", "0")
	_ = writerLoad.WriteField("topic_time_limit", "0")

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

	log.Traceln("[Forum-Poster] Response:", resp)

	if !checkFinalURL(c.FinalURL) {
		return "", fmt.Errorf("[Forum-Poster] NOT Posted - %s", c.FinalURL)
	}

	switch a {
	case "new":
		log.Infof("The URL of new thread is: %v\n", c.FinalURL)
	case "reply":
		log.Infof("The URL of reply is: %v\n", c.FinalURL)
	}

	return c.FinalURL, nil
}

// checkFinalURL chek if finalURL contains f= (means that thread was created)
func checkFinalURL(url string) bool {
	match, _ := regexp.MatchString(`(?m)\/viewtopic\.php\?f=[0-9]*&t=([0-9]+)`, url)
	return match
}
