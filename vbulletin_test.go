package forumposter

import (
	"context"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"os"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
)

func TestCollector_VBulletinLogin(t *testing.T) {
	type fields struct {
		UserAgent string
		Context   context.Context
		LogLevel  string
		LogFile   bool
		Cookie    *cookiejar.Jar
		Client    *http.Client
	}
	type args struct {
		i VBulletinInfoSite
		p Payload
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Fatalf("Failed to set cookie: %s", err)
	}

	client := &http.Client{
		Jar: jar,
	}

	url := os.Getenv("url_vbulletin3")
	if url == "" {
		log.Fatalln("Missing ENV: export url like this: 'export url_vbulletin3=https://somedomain.com'")
	}

	user := os.Getenv("user_vbulletin3")
	if user == "" {
		log.Fatalln("Missing ENV: export user like this: 'export user_vbulletin3=someuser'")
	}

	password := os.Getenv("password_vbulletin3")
	if password == "" {
		log.Fatalln("Missing ENV: export password like this: 'export password_vbulletin3=someuser'")
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			"login",
			fields{
				"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/71.0.3578.98 Safari/537.36",
				nil,
				"debug",
				false,
				jar,
				client,
			},
			args{
				VBulletinInfoSite{
					URL:      url,
					User:     user,
					Password: password,
					T:        "174447",
					F:        "2",
				},
				Payload{
					Title: "asssssss",
					Message: `
					[b]THIS IS [/b] 
					[i]a reply ok?[/i]
				
					`,
				},
			},
			false,
		},
	}

	// Root folder
	cwd, err := os.Getwd()

	if err != nil {
		log.Fatalf("Failed to determine working directory: %s", err)
	}

	// Set as global variable
	os.Setenv("root", cwd)

	logFolder := fmt.Sprintf("%s/log/", cwd) //log folder
	os.Setenv("logFolder", logFolder)

	// Check if folder log exist
	if _, err := os.Stat(logFolder); os.IsNotExist(err) {
		os.MkdirAll(logFolder, os.ModePerm)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Collector{
				UserAgent: tt.fields.UserAgent,
				Context:   tt.fields.Context,
				LogLevel:  tt.fields.LogLevel,
				LogFile:   tt.fields.LogFile,
				Cookie:    tt.fields.Cookie,
				Client:    tt.fields.Client,
			}
			NewCollector()

			LogLevel(tt.fields.LogLevel)

			// Check version of VBulletin (3,4,5)
			checkVersion := &Request{
				URL:    fmt.Sprintf("%s/", url),
				Method: "GET",
			}

			resp, err := c.fetch(checkVersion)
			if err != nil {
				t.Errorf("Collector.VBulletin() error = %v", err)
			}

			log.Traceln("[Forum-Poster]VBulletin - Login response", string(resp))

			err = c.getVersionForum(string(resp))

			if err != nil {
				t.Errorf("Collector.VBulletin() error = %v", err)
			}

			time.Sleep(2 * time.Second)
			if err := c.VBulletin(tt.args.i, tt.args.p); (err != nil) != tt.wantErr {
				t.Errorf("Collector.VBulletin() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCollector_VBulletinPost(t *testing.T) {
	type fields struct {
		UserAgent string
		Context   context.Context
		LogLevel  string
		LogFile   bool
		Cookie    *cookiejar.Jar
		Client    *http.Client
	}
	type args struct {
		i VBulletinInfoSite
		p Payload
		a string
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Fatalf("Failed to set cookie: %s", err)
	}

	client := &http.Client{
		Jar: jar,
	}

	url := os.Getenv("url_vbulletin3")
	if url == "" {
		log.Fatalln("Missing ENV: export url like this: 'export url_vbulletin3=https://somedomain.com'")
	}

	user := os.Getenv("user_vbulletin3")
	if user == "" {
		log.Fatalln("Missing ENV: export user like this: 'export user_vbulletin3=someuser'")
	}

	password := os.Getenv("password_vbulletin3")
	if password == "" {
		log.Fatalln("Missing ENV: export password like this: 'export password_vbulletin3=someuser'")
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			"login",
			fields{
				"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/71.0.3578.98 Safari/537.36",
				nil,
				"trace",
				false,
				jar,
				client,
			},
			args{
				VBulletinInfoSite{
					URL:      url,
					User:     user,
					Password: password,
					T:        "174855",
					F:        "2",
				},
				Payload{
					Title: "GUY",
					Message: `
					[b]Im Vanino [/b]
					[i]im happy to see you good porn sex fisting guy ok?[/i]

					`,
				},
				"reply",
			},
			false,
		},
		/*{
			"login",
			fields{
				"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/71.0.3578.98 Safari/537.36",
				nil,
				"debug",
				false,
				jar,
				client,
			},
			args{
				VBulletinInfoSite{
					URL:      url,
					User:     user,
					Password: password,
					F:        "78",
				},
				Payload{
					Title: "Hi Guys",
					Message: `
					[b]Hello [/b]
					[i]porn hunters[/i]

					`,
				},
				"new",
			},
			false,
		},*/
	}

	// Root folder
	cwd, err := os.Getwd()

	if err != nil {
		log.Fatalf("Failed to determine working directory: %s", err)
	}

	// Set as global variable
	os.Setenv("root", cwd)

	logFolder := fmt.Sprintf("%s/log/", cwd) //log folder
	os.Setenv("logFolder", logFolder)

	// Check if folder log exist
	if _, err := os.Stat(logFolder); os.IsNotExist(err) {
		os.MkdirAll(logFolder, os.ModePerm)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Collector{
				UserAgent: tt.fields.UserAgent,
				Context:   tt.fields.Context,
				LogLevel:  tt.fields.LogLevel,
				LogFile:   tt.fields.LogFile,
				Cookie:    tt.fields.Cookie,
				Client:    tt.fields.Client,
			}
			NewCollector()
			LogLevel("debug")
			var a string
			if a, err = c.VBulletinPost(tt.args.i, tt.args.p, tt.args.a); (err != nil) != tt.wantErr {
				t.Errorf("Collector.VBulletin() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			log.Infoln("Final Response:", a)
		})
	}
}

func TestCollector_VBulletinCheckVersion(t *testing.T) {

	type fields struct {
		UserAgent string
		Context   context.Context
		LogLevel  string
		LogFile   bool
		Cookie    *cookiejar.Jar
		Client    *http.Client
	}
	type args struct {
		url     string
		version string
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Fatalf("Failed to set cookie: %s", err)
	}

	client := &http.Client{
		Jar: jar,
	}

	fields1 := fields{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/71.0.3578.98 Safari/537.36",
		nil,
		"debug",
		false,
		jar,
		client,
	}

	tests := []struct {
		name        string
		URL         string
		wantVersion int
	}{
		{
			"1",
			"https://forum.vbulletin.com/",
			5,
		},
	}

	// Root folder
	cwd, err := os.Getwd()

	if err != nil {
		log.Fatalf("Failed to determine working directory: %s", err)
	}

	// Set as global variable
	os.Setenv("root", cwd)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Collector{
				UserAgent: fields1.UserAgent,
				Context:   fields1.Context,
				LogLevel:  fields1.LogLevel,
				LogFile:   fields1.LogFile,
				Cookie:    fields1.Cookie,
				Client:    fields1.Client,
			}
			NewCollector()

			LogLevel(fields1.LogLevel)

			// Check version of VBulletin (3,4,5)
			checkVersion := &Request{
				URL:    fmt.Sprintf("%s/", tt.URL),
				Method: "GET",
			}

			resp, err := c.fetch(checkVersion)
			if err != nil {
				t.Errorf("Collector.VBulletin() error = %v", err)
				return
			}

			if err := c.getVersionForum(string(resp)); err != nil {
				t.Errorf("Collector.VBulletin() error = %v", err)
				return
			}

			if c.Version != tt.wantVersion {
				t.Errorf("Collector.VBulletin() Want version %d but get %d", tt.wantVersion, c.Version)
			}
		})
	}
}
