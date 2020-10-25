package forumposter

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

//Payload is the content of post to send to forum
type Payload struct {
	Title   string
	Message string
}

// A CollectorOption sets an option on a Collector.
type CollectorOption func(*Collector)

// Collector provides the scraper instance for a scraping job
type Collector struct {
	// UserAgent is the User-Agent string used by HTTP requests
	UserAgent string
	// Context is the context that will be used for HTTP requests. You can set this
	// to support clean cancellation of scraping.
	Context context.Context
	// LogLevel set the level of logging. You can set this to
	// info, debug or trace
	// Default is INFO
	LogLevel string
	// LogFile set the log to print to display or save to file too
	// Default is set to false
	LogFile bool
	//Cookie is the cookie from session
	Cookie *cookiejar.Jar
	//Client is the http client
	Client *http.Client
	//Sid is the SID extracted from cookie
	Sid string
	// Is the URL after the redirect
	FinalURL string
}

var (
	// ErrForbiddenDomain is the error thrown if visiting
	// a domain which is not allowed in AllowedDomains
	ErrForbiddenDomain = errors.New("Forbidden domain")
	// ErrMissingURL is the error type for missing URL errors
	ErrMissingURL = errors.New("Missing URL")
	// ErrMaxDepth is the error type for exceeding max depth
	ErrMaxDepth = errors.New("Max depth limit reached")
	// ErrForbiddenURL is the error thrown if visiting
	// a URL which is not allowed by URLFilters
	ErrForbiddenURL = errors.New("ForbiddenURL")

	// ErrNoURLFiltersMatch is the error thrown if visiting
	// a URL which is not allowed by URLFilters
	ErrNoURLFiltersMatch = errors.New("No URLFilters match")
	// ErrAlreadyVisited is the error type for already visited URLs
	ErrAlreadyVisited = errors.New("URL already visited")
	// ErrRobotsTxtBlocked is the error type for robots.txt errors
	ErrRobotsTxtBlocked = errors.New("URL blocked by robots.txt")
	// ErrNoCookieJar is the error type for missing cookie jar
	ErrNoCookieJar = errors.New("Cookie jar is not available")
	// ErrNoPattern is the error type for LimitRules without patterns
	ErrNoPattern = errors.New("No pattern defined in LimitRule")
	// ErrEmptyProxyURL is the error type for empty Proxy URL list
	ErrEmptyProxyURL = errors.New("Proxy URL list is empty")
	// ErrAbortedAfterHeaders is the error returned when OnResponseHeaders aborts the transfer.
	ErrAbortedAfterHeaders = errors.New("Aborted after receiving response headers")
	// ErrQueueFull is the error returned when the queue is full
	ErrQueueFull = errors.New("Queue MaxSize reached")
)

// Init initializes the Collector's private variables and sets default
// configuration for the Collector
func (c *Collector) Init() {

	c.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169 Safari/537.362s"
	c.Context = context.Background()

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

	// set cookie
	c.Cookie, err = cookiejar.New(nil)
	if err != nil {
		log.Fatalf("Failed to set cookie: %s", err)
	}

	// Initialize client with cookie shared
	c.Client = &http.Client{
		Jar: c.Cookie,
	}

}

// NewCollector creates a new Collector instance with default configuration
func NewCollector(options ...CollectorOption) *Collector {
	c := &Collector{}
	c.Init()

	for _, f := range options {
		f(c)
	}

	return c
}

// LogLevel sets the log level used by the Collector.
func LogLevel(ll string) {
	switch ll {
	case "debug":
		log.SetLevel(log.DebugLevel)
		log.Infoln("DebugLevel LOG Set")
	case "trace":
		log.SetLevel(log.TraceLevel)
		log.Infoln("TraceLevel LOG Set")
	default:
		log.SetLevel(log.InfoLevel)
		log.Infoln("InfoLevel LOG Set")
	}
}

//LogFile save to file logs: true save file, false only print to screen
func LogFile(f bool) {
	//Logging CONFIG
	if f == true {
		runID := time.Now().Format("run-02-01-2006--15-04-05")
		logLocation := filepath.Join(os.Getenv("logFolder"), runID+".log")
		logFile, err := os.OpenFile(logLocation, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalf("Failed to open log file %s for output: %s", logLocation, err)
		}
		log.SetOutput(io.MultiWriter(os.Stderr, logFile))
		log.RegisterExitHandler(func() {
			if logFile == nil {
				return
			}
			logFile.Close()
		})
		log.WithFields(log.Fields{"at": "start", "log-location": logLocation}).Info()
		// perform actions
		//log.Exit(0)

		logrus.SetReportCaller(true)

		//log.SetFormatter(&log.TextFormatter{})
		log.SetFormatter(&logrus.JSONFormatter{CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			s := strings.Split(f.Function, ".")
			funcName := s[len(s)-1]
			return funcName, fmt.Sprintf("%s:%d", path.Base(f.File), f.Line)
		}})
	} else {
		log.SetFormatter(&log.TextFormatter{CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			s := strings.Split(f.Function, ".")
			funcName := s[len(s)-1]
			return funcName, fmt.Sprintf("%s:%d", path.Base(f.File), f.Line)
		}})
	}
}

func (c *Collector) fetch(r *Request) ([]byte, error) {

	var payload io.Reader

	log.WithFields(log.Fields{
		"payload": r.Body,
		"writer":  r.Writer,
		"URL":     r.URL,
		"Method":  r.Method,
	}).Debug("[Forum-Poster] - Value Request")

	if r.Method != "GET" {
		payload = r.Body
	} else {
		payload = nil
	}
	req, err := http.NewRequest(r.Method, r.URL, payload)

	if err != nil {
		log.WithFields(log.Fields{
			"payload": r.Body,
			"writer":  r.Writer,
			"URL":     r.URL,
			"Method":  r.Method,
			"Error":   err,
		}).Error("[Forum-Poster] - Make Request")
		//return nil, fmt.Errorf("[Forum-Poster] - Error in request: %s", err)
	}
	//req.Header.Add("Cookie", c.Cookie)

	if r.Writer != nil {
		req.Header.Set("Content-Type", r.Writer.FormDataContentType())
	}

	req.Header.Set("User-Agent", c.UserAgent)

	res, err := c.Client.Do(req)

	if err != nil {
		log.WithFields(log.Fields{
			"payload": r.Body,
			"writer":  r.Writer,
			"URL":     r.URL,
			"Method":  r.Method,
			"Error":   err,
		}).Error("[Forum-Poster] - Get Response")
		//return nil, fmt.Errorf("[Forum-Poster] - Error in response: %s", err)
	}

	if res.StatusCode == 302 {
		log.Debugln("Redirect to:", res.Header.Get("Location"))
	}

	log.Debugln("[Forum-Poster] - Satus Response", res.StatusCode)

	if res.StatusCode != 200 {
		log.WithFields(log.Fields{
			"payload":    r.Body,
			"writer":     r.Writer,
			"URL":        r.URL,
			"Method":     r.Method,
			"StatusCode": res.StatusCode,
		}).Error("[Forum-Poster] - Get Response")
		//return nil, fmt.Errorf("[Forum-Poster] - Response not valid: %s", err)
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)

	// Your magic function. The Request in the Response is the last URL the
	// client tried to access.
	c.FinalURL = res.Request.URL.String()

	log.Debugf("The URL you ended up at is: %v\n", c.FinalURL)

	log.Debugln("Cookie from", r.URL, "are:", res.Cookies())

	log.Traceln("[Forum-Poster] - HTML response: ", string(body))

	for _, cookie := range res.Cookies() {

		log.Tracef("  %s: %s\n", cookie.Name, cookie.Value)
		if strings.Contains(cookie.Name, "sid") {
			c.Sid = cookie.Value
		}

	}

	// Sleep every request so POST will work
	time.Sleep(2 * time.Second)

	return body, nil
}

// UserAgent sets the user agent used by the Collector.
func UserAgent(ua string) CollectorOption {
	return func(c *Collector) {
		c.UserAgent = ua
	}
}
