package forumposter

import (
	"context"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"os"
	"testing"

	log "github.com/sirupsen/logrus"
)

func TestCollector_VBulletin(t *testing.T) {
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
					T:        "3774608",
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
			LogLevel("debug")
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
					T:        "994063",
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
