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

func TestCollector_ForumophiliaPost(t *testing.T) {
	type fields struct {
		UserAgent string
		Context   context.Context
		LogLevel  string
		LogFile   bool
		Cookie    *cookiejar.Jar
		Client    *http.Client
	}
	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Fatalf("Failed to set cookie: %s", err)
	}

	client := &http.Client{
		Jar: jar,
	}

	var url = "https://www.forumophilia.com"

	user := os.Getenv("user_forumophilia")
	if user == "" {
		log.Fatalln("Missing ENV: export user like this: 'export user_forumophilia=someuser'")
	}

	password := os.Getenv("password_forumophilia")
	if password == "" {
		log.Fatalln("Missing ENV: export password like this: 'export password_forumophilia=someuser'")
	}

	type args struct {
		i ForumophiliaInfoSite
		p Payload
		a string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		/*{
			"post new",
			fields{
				"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/71.0.3578.98 Safari/537.36",
				nil,
				"debug",
				false,
				jar,
				client,
			},
			args{
				PHPBB3InfoSite{
					URL:      url,
					User:     user,
					Password: password,
					F:        "17",
				},
				Payload{
					Title: "asssssss",
					Message: `
					[b]aa[/b]
					[i]sdfsd[/i]

					[img]someimage[/img]`,
				},
				"post",
			},
			false,
		},*/
		{
			"reply",
			fields{
				"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/71.0.3578.98 Safari/537.36",
				nil,
				"debug",
				false,
				jar,
				client,
			},
			args{
				ForumophiliaInfoSite{
					URL:      url,
					User:     user,
					Password: password,
					T:        "767242",
					F:        "9",
				},
				Payload{
					Title: "I'm vanino1 cool",
					Message: `
					[b]Can you [/b] 
					[i]read me ok?[/i]
					
					`,
				},
				"reply",
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
			LogLevel("trace")
			var a string
			if a, err = c.ForumophiliaPost(tt.args.i, tt.args.p, tt.args.a); (err != nil) != tt.wantErr {
				t.Errorf("Collector.PHPBB3 POST NEW() error = %s, wantErr %v", err, tt.wantErr)
			}
			log.Infoln("Posted on", a)
		})
	}
}
