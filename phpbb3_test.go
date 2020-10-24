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

func TestCollector_PHPBB3Post(t *testing.T) {
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

	url := os.Getenv("url_phpbb")
	if url == "" {
		log.Fatalln("Missing ENV: export url like this: 'export url_phpbb=https://somedomain.com'")
	}

	user := os.Getenv("user_phpbb")
	if user == "" {
		log.Fatalln("Missing ENV: export user like this: 'export user_phpbb=someuser'")
	}

	password := os.Getenv("password_phpbb")
	if password == "" {
		log.Fatalln("Missing ENV: export password like this: 'export password_phpbb=someuser'")
	}

	type args struct {
		i PHPBB3InfoSite
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
				PHPBB3InfoSite{
					URL:      url,
					User:     user,
					Password: password,
					T:        "225626",
					F:        "17",
				},
				Payload{
					Title: "asssssss",
					Message: `
					[b]THIS IS [/b] 
					[i]a reply ok?[/i]
					
					[img]someimage[/img]`,
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
			LogLevel("debug")
			var a string
			if a, err = c.PHPBB3Post(tt.args.i, tt.args.p, tt.args.a); (err != nil) != tt.wantErr {
				t.Errorf("Collector.PHPBB3 POST NEW() error = %d, wantErr %v", err, tt.wantErr)
			}
			log.Infoln("Posted on", a)
		})
	}
}

func Test_checkFinalURL(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"post",
			args{
				"https://www.domain.com/viewtopic.php?f=17&t=225622",
			},
			true,
		},
		{
			"reply",
			args{
				"https://www.domain.com/viewtopic.php?f=17&t=225626&p=15076#p15076 ",
			},
			true,
		},
		{
			"some",
			args{
				"https://www.domain.com/viewtopic.php",
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := checkFinalURL(tt.args.url); got != tt.want {
				t.Errorf("checkFinalURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
