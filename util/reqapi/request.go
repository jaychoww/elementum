package reqapi

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/anacrolix/missinggo/perf"
	"github.com/dustin/go-humanize"
	"github.com/goccy/go-json"
	"github.com/jmcvetta/napping"

	"github.com/elgatito/elementum/config"
	"github.com/elgatito/elementum/util"
	"github.com/elgatito/elementum/util/trace"
)

type Request struct {
	trace.Tracer
	API *API

	Method      string
	URL         string
	Params      url.Values  `msg:"-"`
	Header      http.Header `msg:"-"`
	Payload     *bytes.Buffer
	Description string

	Retry        int
	RetryBackoff time.Duration

	ResponseError      error
	ResponseStatus     string
	ResponseStatusCode int
	ResponseHeader     http.Header
	ResponseBody       *bytes.Buffer
	ResponseSize       uint64

	Result any
}

func (r *Request) Prepare() (err error) {
	r.URL = r.API.GetURL(r.URL)

	if r.Method == "" {
		if r.Payload != nil {
			r.Method = "POST"
		} else {
			r.Method = "GET"
		}
	}

	if r.Header == nil {
		r.Header = http.Header{}
	}

	return nil
}

func (r *Request) Do() (err error) {
	defer perf.ScopeTimer()()

	r.Create()
	if config.Args.EnableRequestTracing {
		defer func() {
			r.Error(err)
			log.Debugf(r.String())
		}()
	}

	if r.API == nil {
		err = errors.New("API not defined")
		return
	}

	// Do internal preparations
	if err = r.Prepare(); err != nil {
		return
	}

	req := &napping.Request{
		Url:                 r.URL,
		Method:              r.Method,
		Params:              &r.Params,
		Header:              &r.Header,
		CaptureResponseBody: true,
	}

	// Apply body payload to request
	if r.Payload != nil {
		req.Payload = r.Payload
		req.RawPayload = true
	}

	var resp *napping.Response

	r.API.RateLimiter.Call(func() error {
		r.Stage("Request")

		for {
			resp, err = r.API.GetSession().Send(req)

			if resp != nil {
				r.ResponseStatusCode = resp.Status()
				r.ResponseStatus = resp.HttpResponse().Status
				r.ResponseHeader = resp.HttpResponse().Header
			}

			if err != nil {
				log.Errorf("Failed to make request to %s for %s with %+v: %s", r.URL, r.Description, r.Params, err)
			} else if r.ResponseStatusCode == 429 {
				log.Warningf("Rate limit exceeded getting %s with %+v on %s, cooling down...", r.Description, r.Params, r.URL)
				r.API.RateLimiter.CoolDown(r.ResponseHeader)
				err = util.ErrExceeded
				return err
			} else if r.ResponseStatusCode == 404 {
				log.Errorf("Bad status getting %s with %+v on %s: %d", r.Description, r.Params, r.URL, r.ResponseStatus)
				err = util.ErrNotFound
				return err
			} else if r.ResponseStatusCode == 403 && r.API.RetriesLeft > 0 {
				r.API.RetriesLeft--
				log.Warningf("Not authorized to get %s with %+v on %s, having %d retries left ...", r.Description, r.Params, r.URL, r.API.RetriesLeft)
				continue
			} else if r.ResponseStatusCode < 200 || r.ResponseStatusCode >= 300 {
				log.Errorf("Bad status getting %s with %+v on %s: %d", r.Description, r.Params, r.URL, r.ResponseStatus)
				err = util.ErrHTTP
				return err
			}

			break
		}

		err = nil
		return nil
	})

	r.Stage("Response")

	if resp != nil && resp.ResponseBody != nil {
		r.Size(uint64(resp.ResponseBody.Len()))
		if r.Result != nil {
			err = json.Unmarshal(resp.ResponseBody.Bytes(), r.Result)
			r.Stage("Unmarshal")
		} else {
			r.ResponseBody = resp.ResponseBody
		}
	}

	r.Complete()
	return
}

func (r *Request) String() string {
	if r.IsComplete() {
		r.Complete()
	}

	params, _ := url.QueryUnescape(r.Params.Encode())
	return fmt.Sprintf(`Trace for request: %s
               URL: %s %s
            Params: %s
            Header: %+v
%s

             Error: %#v
              Size: %s
            Status: %s
        StatusCode: %d
   Response Header: %+v
	`, r.Description, r.Method, r.URL,
		params, r.Header, r.Tracer.String(),
		r.ResponseError, humanize.Bytes(r.ResponseSize), r.ResponseStatus, r.ResponseStatusCode, r.ResponseHeader)
}

func (r *Request) Reset() {
	r.Tracer.Reset()

	r.ResponseSize = 0
	r.ResponseBody = nil
	r.ResponseError = nil
}

func (r *Request) Size(size uint64) {
	r.ResponseSize = size
}

func (r *Request) Error(err error) {
	if err != nil {
		r.ResponseError = err
	}
}
