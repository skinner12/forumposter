package forumposter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	log "github.com/sirupsen/logrus"
)

//IntPornInfoSite provides the info to make https request for post
type IntPornInfoSite struct {
	URL                    string
	User                   string
	Password               string
	F                      string // Forum number, for new post
	T                      string // Thread Number, for reply
	CSRF                   string // csrf for POST
	AttachmentHash         string
	AttachmentHashCombined string
	LastDate               string
	LastKnowDate           string
}

type intPornReponse struct {
	Status string `json:"status"`
	HTML   struct {
		Content string   `json:"content"`
		CSS     []string `json:"css"`
		Js      []string `json:"js"`
	} `json:"html"`
	LastDate int `json:"lastDate"`
	Visitor  struct {
		ConversationsUnread string `json:"conversations_unread"`
		AlertsUnread        string `json:"alerts_unread"`
		TotalUnread         string `json:"total_unread"`
	} `json:"visitor"`
}

//IntPornLogin function to make login. Return Error
func (c *Collector) IntPornLogin(i IntPornInfoSite, url string) error {

	postLogin := &Request{
		Body:   nil,
		URL:    url,
		Method: "POST",
		Writer: nil,
	}

	resp, err := c.fetch(postLogin)
	if err != nil {
		return err
	}

	log.Traceln("[Forum-Poster] Response:", string(resp))

	// Check if Login
	if err := i.checkLogin(string(resp)); err != nil {
		return err
	}

	return nil
}

// Check class .p-navgroup-linkText and looks for the username
func (i *IntPornInfoSite) checkLogin(p string) error {

	// Load the HTML document
	log.Debugf("Looking for username %s into page", i.User)
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(p)))
	if err != nil {
		log.Fatal(err)
	}

	// Extract
	// * username

	found := false

	doc.Find(".p-navgroup-linkText").Each(func(n int, s *goquery.Selection) {
		if s.Text() == i.User {
			log.Debugf("Login Check Successfull")
			found = true
		}
	})

	if !found {
		return ErrLoginFailed
	}

	return nil

}

func (i *IntPornInfoSite) getCSRF(p string) error {

	// Load the HTML document
	log.Debugln("Extracting CSRF")
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(p)))
	if err != nil {
		return fmt.Errorf("Can't decode for extract CSRF")
	}

	// Extract
	// * csrf

	var ok bool

	i.CSRF, ok = doc.Find("input[name='_xfToken']").Attr("value")
	if !ok {
		return fmt.Errorf("Can't find form_token")
	}

	return nil
}

func (i *IntPornInfoSite) getValuePost(p string) error {

	// Extract:
	// attachment_hash
	// attachment_hash_combined
	// last_date
	// last_known_date

	// Load the HTML document
	log.Debugln("Extracting Value for Post")
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(p)))
	if err != nil {
		return fmt.Errorf("Can't decode for extract value for post")
	}

	var ok bool

	i.AttachmentHash, ok = doc.Find("input[name='attachment_hash']").Attr("value")
	if !ok {
		return fmt.Errorf("Can't find attachment_hash")
	}

	i.AttachmentHashCombined, ok = doc.Find("input[name='attachment_hash_combined']").Attr("value")
	if !ok {
		return fmt.Errorf("Can't find attachment_hash_combined")
	}

	i.LastDate, ok = doc.Find("input[name='last_date']").Attr("value")
	if !ok {
		return fmt.Errorf("Can't find last_date")
	}

	i.LastKnowDate, ok = doc.Find("input[name='last_known_date']").Attr("value")
	if !ok {
		return fmt.Errorf("Can't find last_known_date")
	}

	return nil
}

// IntPorn function to post to IntoPorn.com
func (c *Collector) IntPorn(i IntPornInfoSite, p Payload, a string) (string, error) {
	var url string
	//var action string

	// Read Homepgae to extract CSRF
	homePage := &Request{
		URL:    fmt.Sprintf("%s/", i.URL),
		Method: "GET",
	}

	hp, err := c.fetch(homePage)
	if err != nil {
		return "", err
	}

	log.Traceln("[Forum-Poster]IntPorn - HomePage Response", string(hp))

	// Get CSRF
	if err := i.getCSRF(string(hp)); err != nil {
		return "", fmt.Errorf("[Forum-Poster]IntPorn - Can not get CSRF: %s", err)
	}

	log.WithFields(log.Fields{
		"csrf":    i.CSRF,
		"URL":     fmt.Sprintf("%s/", i.URL),
		"Version": c.Version,
	}).Debug("[Forum-Poster] - Extract Values")

	// Check if login is still available
	if err := i.checkLogin(string(hp)); err != nil {
		log.Debugf("[Forum-Poster]IntPorn - Make new login")
		url = fmt.Sprintf("%s/login/login?login=%s&password=%s&remember=1&_xfRedirect=%s&_xfToken=%s", i.URL, i.User, i.Password, i.URL, i.CSRF)

		if err := c.IntPornLogin(i, url); err != nil {
			return "", fmt.Errorf("[Forum-Poster]IntPorn Login - %s", err)
		}
	}

	// Set post NEW or REPLY
	switch a {
	/*case "new":
	log.Infoln("* Post new thread to", i.URL)
	url = fmt.Sprintf("%s/newthread.php?do=newthread&f=%s", i.URL, i.F)
	action = "postthread"*/
	case "reply":
		log.Infoln("* Reply thread to", i.T)
		url = fmt.Sprintf("%s/threads/%s", i.URL, i.T)
	default:
		return "", fmt.Errorf("[Forum-Poster]IntPorn - Choice are: new or reply. Set the right one")
	}

	// For post new/reply, forum needs exact URL like
	// https://www.intporn.org/threads/test.1234567/
	// where to make request, so make first request only with T
	// number, and save response URL
	readRealThread := &Request{
		Body:   nil,
		URL:    url,
		Method: "GET",
		Writer: nil,
	}

	log.WithFields(log.Fields{
		"readRealThread": readRealThread,
		"URL":            url,
	}).Debug("[Forum-Poster]VBulletin - Extract Values")

	body, err := c.fetch(readRealThread)
	if err != nil {
		return "", fmt.Errorf("[Forum-Poster]IntPorn Read Page for Get Data - %s", err)
	}

	log.Traceln("[Forum-Poster]IntPorn - Real Thread Response", string(body))

	// Extract value for post data
	if err := i.getValuePost(string(body)); err != nil {
		return "", fmt.Errorf("[Forum-Poster]IntPorn Extract Data - %s", err)
	}

	// Get CSRF
	if err := i.getCSRF(string(body)); err != nil {
		return "", fmt.Errorf("[Forum-Poster]IntPorn - Can not get CSRF: %s", err)
	}

	// Get last part of URL
	xfRequestUris := strings.Split(c.FinalURL, "threads")
	xfRequestURI := fmt.Sprintf("/threads%s", xfRequestUris[len(xfRequestUris)-1])

	// Make URL to Post
	// Set post NEW or REPLY
	switch a {
	/*case "new":
	log.Infoln("* Post new thread to", i.URL)
	url = fmt.Sprintf("%s/newthread.php?do=newthread&f=%s", i.URL, i.F)
	action = "postthread"*/
	case "reply":
		url = fmt.Sprintf("%sadd-reply", c.FinalURL)
	default:
		return "", fmt.Errorf("[Forum-Poster]IntPorn - Choice are: new or reply. Set the right one")
	}

	log.WithFields(log.Fields{
		"attachment_hash":          i.AttachmentHash,
		"attachment_hash_combined": i.AttachmentHashCombined,
		"last_date":                i.LastDate,
		"last_known_date":          i.LastKnowDate,
		"Thread":                   i.T,
		"URL":                      url,
		"URL Redirected":           c.FinalURL,
		"CSRF":                     i.CSRF,
		"xfRequestUri":             xfRequestURI,
	}).Debug("[Forum-Poster]IntPorn - Extract Values")

	// Post Reply Thread
	postload := &bytes.Buffer{}
	writerLoad := multipart.NewWriter(postload)
	_ = writerLoad.WriteField("attachment_hash", i.AttachmentHash)
	_ = writerLoad.WriteField("attachment_hash_combined", i.AttachmentHashCombined)
	_ = writerLoad.WriteField("last_date", i.LastDate)
	_ = writerLoad.WriteField("last_known_date", i.LastKnowDate)

	_ = writerLoad.WriteField("message_html", p.Message)

	_ = writerLoad.WriteField("_xfToken", i.CSRF)
	_ = writerLoad.WriteField("_xfRequestUri", xfRequestURI) // /threads/testing.1874299/
	_ = writerLoad.WriteField("_xfWithData", "1")
	_ = writerLoad.WriteField("_xfToken", i.CSRF)
	_ = writerLoad.WriteField("_xfResponseType", "json")

	err = writerLoad.Close()
	if err != nil {
		fmt.Println(err)
		return "", fmt.Errorf("[Forum-Poster]IntPorn -  Post Thread %v", err)
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

	log.Debugln(string(resp))

	log.Traceln("[Forum-Poster]IntPorn -  Response:", string(resp))

	// Unmarshal Response
	rp := intPornReponse{}
	if err := json.Unmarshal(resp, &rp); err != nil {
		return "", fmt.Errorf("[Forum-Poster]IntPorn -  Unmarshal Response %v", err)
	}

	if rp.Status != "ok" {
		return "", fmt.Errorf("[Forum-Poster]IntPorn -  Not Posted")
	}

	// Get ID of post
	var re = regexp.MustCompile(`(?m)post-(\d+)`)

	id := re.FindString(rp.HTML.Content)

	return fmt.Sprintf("%s%s%s", i.URL, xfRequestURI, id), nil

}
