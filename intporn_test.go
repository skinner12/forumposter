package forumposter

import (
	"context"
	"net/http"
	"net/http/cookiejar"
	"os"
	"testing"

	log "github.com/sirupsen/logrus"
)

func TestCollector_IntPorn(t *testing.T) {
	type fields struct {
		UserAgent string
		Context   context.Context
		LogLevel  string
		LogFile   bool
		Cookie    *cookiejar.Jar
		Client    *http.Client
	}
	type args struct {
		i IntPornInfoSite
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

	url := os.Getenv("url_intporn")
	if url == "" {
		log.Fatalln("Missing ENV: export url like this: 'export url_intporn=https://somedomain.com'")
	}

	user := os.Getenv("user_intporn")
	if user == "" {
		log.Fatalln("Missing ENV: export user like this: 'export user_intporn=someuser'")
	}

	password := os.Getenv("password_intporn")
	if password == "" {
		log.Fatalln("Missing ENV: export password like this: 'export password_intporn=someuser'")
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
				IntPornInfoSite{
					URL:      url,
					User:     user,
					Password: password,
					T:        "1874299",
				},
				Payload{
					Title: "GUY",
					Message: `
					<h2>Hi</h2>
					<p>im happy to see you good porn sex fisting guy ok?</p>

					`,
				},
				"reply",
			},
			false,
		},
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
			a, err := c.IntPorn(tt.args.i, tt.args.p, tt.args.a)
			if (err != nil) != tt.wantErr {
				t.Errorf("Collector.IntPorn() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			log.Infoln("Final Response:", a)

		})
	}
}
