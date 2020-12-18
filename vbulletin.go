package forumposter

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"mime/multipart"
	"regexp"
	"time"

	log "github.com/sirupsen/logrus"
)

//VBulletinInfoSite provides the info to make https request for post
type VBulletinInfoSite struct {
	URL           string
	User          string
	Password      string
	F             string // Forum number, for new post
	T             string // Thread Number, for reply
	SecurityToken string // Token after login
	Version       int    // Version of VBulletin Forum (3,4,5)

}

//VBulletin manage vbulletin forum
func (c *Collector) VBulletin(i VBulletinInfoSite, p Payload) error {

	hash := md5.Sum([]byte(i.Password))
	hashPassword := hex.EncodeToString(hash[:])

	// Make LOGIN
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	/*
			s	""
		do	"login"
		vb_login_md5password	"02eb2a6aab26405366779456c9ebfa86"
		vb_login_md5password_utf	"02eb2a6aab26405366779456c9ebfa86"
		vb_login_username	"skinner122"
		vb_login_password	""
	*/

	log.Debugf("[Forum-Poster] - VBulletin Version %d", c.Version)

	_ = writer.WriteField("s", "")
	_ = writer.WriteField("do", "login")
	_ = writer.WriteField("vb_login_md5password", hashPassword)

	_ = writer.WriteField("vb_login_username", i.User)

	if c.Version == 3 {
		_ = writer.WriteField("vb_login_password", i.Password)
		_ = writer.WriteField("vb_login_md5password_utf", "")
		_ = writer.WriteField("cookieuser", "1")
		_ = writer.WriteField("securitytoken", "guest")
	}

	if c.Version == 4 || c.Version == 5 {
		_ = writer.WriteField("vb_login_password", i.Password)
		_ = writer.WriteField("vb_login_md5password_utf", hashPassword)
		_ = writer.WriteField("vb_login_password_hint", "Password")
		_ = writer.WriteField("securitytoken", "guest")
		_ = writer.WriteField("cookieuser", "1")
	}

	err := writer.Close()
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

	resp, err := c.fetch(postLogin)
	if err != nil {
		return err
	}

	log.Traceln("[Forum-Poster]VBulletin - Login response", string(resp))

	// Set SecurityToken

	err = c.getSecurityToken(string(resp))

	if err != nil {
		return err
	}

	return nil
}

func (c *Collector) getSecurityToken(resp string) error {
	// Extract security token
	var re = regexp.MustCompile(`(?m)var SECURITYTOKEN = \"(.*?)\"`)
	a := re.FindStringSubmatch(resp)

	if a == nil {
		log.Errorf("[Forum-Poster]VBulletin - Login Error, missing security token")
		return fmt.Errorf("[Forum-Poster]VBulletin - Login Error, missing security token")
	}

	log.Debugln("[Forum-Poster]VBulletin - Login - Found SecurityToken", a[1])
	log.Infoln("[Forum-Poster]VBulletin - Login made with success")
	c.SecurityToken = string(a[1])
	return nil
}

func (c *Collector) getVersionForum(resp string) error {

	// Set default version 3
	c.Version = 3

	var re4 = regexp.MustCompile(`(?mi)SIMPLEVERSION\s=\s"(\d+)"`)
	a := re4.FindStringSubmatch(resp)

	if a != nil {
		log.Infoln("[Forum-Poster] VBulletin - Version 4 found")
		c.Version = 4
		return nil
	}

	var re5 = regexp.MustCompile(`(?mi)"simpleversion":\s"v=(\d+)"`)
	b := re5.FindStringSubmatch(resp)

	if b != nil {
		log.Infoln("[Forum-Poster] VBulletin - Version 5 found")
		c.Version = 5
		return nil
	}

	log.Infoln("[Forum-Poster] VBulletin - Version 3 found")

	return nil

}

//VBulletinPost post new thread
//a is for chose if reply or new thread
func (c *Collector) VBulletinPost(i VBulletinInfoSite, p Payload, a string) (string, error) {

	var url string
	var action string

	// TODO: check if "simpleversion": "v=564" if present. V3 not have that

	// Check version of VBulletin (3,4,5)
	checkVersion := &Request{
		URL:    fmt.Sprintf("%s/", i.URL),
		Method: "GET",
	}

	resp, err := c.fetch(checkVersion)
	if err != nil {
		return "", err
	}

	log.Traceln("[Forum-Poster]VBulletin - Login response", string(resp))

	err = c.getVersionForum(string(resp))

	if err != nil {
		return "", err
	}

	// Set post NEW or REPLY
	switch a {
	case "new":
		log.Infoln("* Post new thread to", i.URL)
		url = fmt.Sprintf("%s/newthread.php?do=newthread&f=%s", i.URL, i.F)
		action = "postthread"
	case "reply":
		log.Infoln("* Reply thread to", i.URL)
		url = fmt.Sprintf("%s/newreply.php?do=postreply&t=%s", i.URL, i.T)
		action = "postreply"
	default:
		return "", fmt.Errorf("[Forum-Poster] - Choice are: new or reply. Set the right one")
	}

	// Login first
	err = c.VBulletin(i, p)
	if err != nil {
		return "", err
	}

	log.Debugln("SECURITY TOKEN", c.SecurityToken)

	if c.SecurityToken == "guest" {
		log.Errorf("[Forum-Poster]VBulletin - Login Error, security token is GUEST")
		return "", ErrLoginFailed
	}

	var re = regexp.MustCompile(`(?m)(\d+)-(.+)"`)

	if !re.MatchString(c.SecurityToken) {
		log.Infoln("[Forum-Poster]VBulletin - No securitytoken")

		// Load home page to get SID from cookie
		readPostForm := &Request{
			Body:   nil,
			URL:    url,
			Method: "POST",
			Writer: nil,
		}

		log.WithFields(log.Fields{
			"readPostForm": readPostForm,
			"URL":          url,
		}).Debug("[Forum-Poster]VBulletin - Extract Values")

		body, err := c.fetch(readPostForm)
		if err != nil {
			return "", err
		}

		err = c.getSecurityToken(string(body))

		if err != nil {
			return "", err
		}

	}

	time.Sleep(1 * time.Second)

	// Post New Thread
	postload := &bytes.Buffer{}
	writerLoad := multipart.NewWriter(postload)
	_ = writerLoad.WriteField("do", action)
	_ = writerLoad.WriteField("message", p.Message)
	_ = writerLoad.WriteField("securitytoken", c.SecurityToken)
	if a == "new" {
		if p.Title == "" {
			return "", fmt.Errorf("[Forum-Poster]VBulletin - Thread not allowed without title")
		}
		_ = writerLoad.WriteField("f", i.F)
		_ = writerLoad.WriteField("subject", p.Title)
		_ = writerLoad.WriteField("vB_Editor_001_mode", "wysiwyg")
	}

	if a == "reply" {
		if p.Title == "" {
			p.Title = "One more..."
		}
		_ = writerLoad.WriteField("t", i.T)
		_ = writerLoad.WriteField("title", p.Title)
	}

	err = writerLoad.Close()
	if err != nil {
		fmt.Println(err)
		return "", fmt.Errorf("[Forum-Poster]VBulletin Post - %v", err)
	}

	if i.T != "" {
		log.Debugln("[Forum-Poster]VBulletin - Repling to THREAD ID", i.T)
	} else {
		log.Debugln("[Forum-Poster]VBulletin - Posting to FORUM ID", i.F)
	}

	postThread := &Request{
		Body:   postload,
		URL:    url,
		Method: "POST",
		Writer: writerLoad,
	}

	resp, err = c.fetch(postThread)
	if err != nil {
		return "", err
	}

	log.Traceln("[Forum-Poster]VBulletin - Response:", string(resp))

	if !checkFinalURL(c.FinalURL) {
		return "", fmt.Errorf("[Forum-Poster]VBulletin - NOT Posted - %s", c.FinalURL)
	}

	switch a {
	case "new":
		log.Infof("[Forum-Poster]VBulletin - The URL of new thread is: %v\n", c.FinalURL)
	case "reply":
		log.Infof("[Forum-Poster]VBulletin - The URL of reply is: %v\n", c.FinalURL)
	}

	return c.FinalURL, nil
}
