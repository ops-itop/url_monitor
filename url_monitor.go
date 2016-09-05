package url_monitor

import (
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
	"regexp"
	"strconv"
	"fmt"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// HTTPResponse struct
type HTTPResponse struct {
	Cmdbid          string
	App				string
	Address         string
	Body            string
	Method          string
	ResponseTimeout internal.Duration
	Headers         map[string]string
	FollowRedirects bool
	RequireStr		string
	RequireCode     string
	FailedCount     int
	FailedTimeout   float64

	// Path to CA file
	SSLCA string `toml:"ssl_ca"`
	// Path to host cert file
	SSLCert string `toml:"ssl_cert"`
	// Path to cert key file
	SSLKey string `toml:"ssl_key"`
	// Use SSL but skip chain & host verification
	InsecureSkipVerify bool
}

// Description returns the plugin Description
func (h *HTTPResponse) Description() string {
	return "HTTP/HTTPS request given an address a method and a timeout"
}

var sampleConfig = `
  ## App Name
  app = "monitor"
  ## CMDB ID
  cmdbid = "1701"
  ## Server address (default http://localhost)
  address = "http://www.baidu.com"
  ## Set response_timeout (default 5 seconds)
  response_timeout = "5s"
  ## HTTP Request Method
  method = "GET"
  ## Require String 正则表达式用单引号避免转义
  require_str = 'baidu.com'
  require_code = '20\d'
  failed_count = 3
  failed_timeout = 0.5
  ## Whether to follow redirects from the server (defaults to false)
  follow_redirects = true
  ## Optional HTTP Request Body
  # body = '''
  # {'fake':'data'}
  # '''

  ## Optional SSL Config
  # ssl_ca = "/etc/telegraf/ca.pem"
  # ssl_cert = "/etc/telegraf/cert.pem"
  # ssl_key = "/etc/telegraf/key.pem"
  ## Use SSL but skip chain & host verification
  # insecure_skip_verify = false

  ## HTTP Request Headers (all values must be strings)
  ## 表格名下，直到下一个表格名或文件尾，均为当前表格的内容 所以Headers应该放在最后
  # [inputs.url_monitor.headers]
  #   Host = "github.com"
`

// SampleConfig returns the plugin SampleConfig
func (h *HTTPResponse) SampleConfig() string {
	return sampleConfig
}

// ErrRedirectAttempted indicates that a redirect occurred
var ErrRedirectAttempted = errors.New("redirect")

// CreateHttpClient creates an http client which will timeout at the specified
// timeout period and can follow redirects if specified
func (h *HTTPResponse) createHttpClient() (*http.Client, error) {
	tlsCfg, err := internal.GetTLSConfig(
		h.SSLCert, h.SSLKey, h.SSLCA, h.InsecureSkipVerify)
	if err != nil {
		return nil, err
	}
	tr := &http.Transport{
		ResponseHeaderTimeout: h.ResponseTimeout.Duration,
		TLSClientConfig:       tlsCfg,
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   h.ResponseTimeout.Duration,
	}

	if h.FollowRedirects == false {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return ErrRedirectAttempted
		}
	}
	return client, nil
}

// HTTPGather gathers all fields and returns any errors it encounters
func (h *HTTPResponse) HTTPGather() (map[string]interface{}, error) {
	// Prepare fields
	fields := make(map[string]interface{})

	// 经常被用户更改配置的变量不在作为tag，作为fields
	fields["require_code"] = h.RequireCode
	fields["require_str"] = h.RequireStr
	fields["require_time"] = strconv.FormatFloat(h.FailedTimeout, 'g', 1, 64)
	fields["failed_threshold"] = strconv.Itoa(h.FailedCount)

	client, err := h.createHttpClient()
	if err != nil {
		return nil, err
	}

	var body io.Reader
	address := h.Address
	if h.Body != "" {
		body = strings.NewReader(h.Body)
		if h.Method == "GET" {
			address = h.Address + "?" + h.Body
			body = nil
		}
	}
	request, err := http.NewRequest(h.Method, address, body)
	if err != nil {
		fields["msg"] = err
		//return fields,nil
		//return nil, err
	}

	content_type := 0
	for key, val := range h.Headers {
		request.Header.Add(key, val)
		if key == "Host" {
			request.Host = val
		}
		
		//如果指定了content-type, content_type设置为1
		if strings.ToLower(key) == "content-type" {
			content_type = 1
		}
	}
	//如果没有指定content-type，则设置默认值为application/x-www-form-urlencoded
	if content_type == 0 {
		request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}

	// Start Timer
	start := time.Now()
	resp, err := client.Do(request)
	if err != nil {
		if h.FollowRedirects {
			fields["msg"] = err
			//return fields,nil
			//return nil, err
		}
		if urlError, ok := err.(*url.Error); ok &&
			urlError.Err == ErrRedirectAttempted {
			err = nil
		} else {
			fields["msg"] = err
			//err = nil
			//return fields,nil
		}
	}

	_, ok := fields["msg"] 
	if ok {
		fields["data_match"] = 0
		fields["code_match"] = 0
		fields["time_match"] = 0
		fields["response_time"] = time.Since(start).Seconds()
		fields["http_code"] = 000
		return fields, nil
	}
	// require string
	if h.RequireStr == "" {
		fields["data_match"] = 1
	}else{
		body,_ := ioutil.ReadAll(resp.Body)
		bodystr := string(body)

		r,err := regexp.Compile(h.RequireStr)
		//r,_ := regexp.CompilePOSIX(h.RequireStr)

		if err != nil {
			if strings.Contains(bodystr, h.RequireStr) {
				fields["data_match"] = 1
			}else{
				fields["data_match"] = 0
				// fields['msg']中文unicode转字符串，并截取超长的内容
				fields["msg"] = suberrmsg(bodystr)
			}
		}else{
			if r.FindString(bodystr) != ""{
				fields["data_match"] = 1
			}else {
				fields["data_match"] = 0
				fields["msg"] = suberrmsg(bodystr)
			}
		}
		
	}

	// require http code
	if h.RequireCode == "" {
		fields["code_match"] = 1
	}else {
		status_code :=  strconv.Itoa(resp.StatusCode)

		r,err := regexp.Compile(h.RequireCode)
		//r,_ := regexp.CompilePOSIX(h.RequireCode)
		
		if err != nil {
			if strings.Contains(status_code, h.RequireCode) {
				fields["code_match"] = 1
			}else{
				fields["code_match"] = 0
			}

		}else{
			if r.FindString(status_code) != "" {
				fields["code_match"] = 1
			}else {
				fields["code_match"] = 0
			}
		}
	}
	fields["response_time"] = time.Since(start).Seconds()
	fields["http_code"] = resp.StatusCode

	// require response time
	fields["time_match"] = 1

	rt := fields["response_time"].(float64)
	if rt > h.FailedTimeout {
		fields["time_match"] = 0
	}
	if fields["time_match"] == 0 && rt < h.FailedTimeout * 0.7 {
		fields["time_match"] = 1
	}

	return fields, nil
}

// Gather gets all metric fields and tags and returns any errors it encounters
func (h *HTTPResponse) Gather(acc telegraf.Accumulator) error {
	// Set default values
	if h.ResponseTimeout.Duration < time.Second {
		h.ResponseTimeout.Duration = time.Second * 5
	}
	// Check send and expected string
	if h.Method == "" {
		h.Method = "GET"
	}
	if h.Address == "" {
		h.Address = "http://localhost"
	}
	addr, err := url.Parse(h.Address)
	if err != nil {
		return err
	}
	if addr.Scheme != "http" && addr.Scheme != "https" {
		return errors.New("Only http and https are supported")
	}
	// Prepare data
	tags := map[string]string{"cmdbid": h.Cmdbid, "app": h.App, "url": h.Address, "method": h.Method}
	var fields map[string]interface{}
	// Gather data
	fields, err = h.HTTPGather()
	if err != nil {
		return err
	}
	// Add metrics
	acc.AddFields("url_monitor", fields, tags)
	return nil
}

func init() {
	inputs.Add("url_monitor", func() telegraf.Input {
		return &HTTPResponse{}
	})
}

// fields['msg']最长不超过1kb
func suberrmsg(errmsg string) (submsg string) {
	limit := 1250
	u2s,_ := unicode2str(errmsg)
	if len(u2s) > limit {
		submsg = u2s[0:limit]
	}else{
		submsg = u2s
	}
	return
}

// convert unicode chinese char to string
func unicode2str(u string) (context string, err error) {
    sUnicodev := strings.Split(u,"\\u")
    for _, v := range sUnicodev {
        if len(v) < 1 {
            continue
        }
		if len(v) > 4 {
			v1 := v[0:4]
			v2 := v[4:len(v)]
			temp, err := strconv.ParseInt(v1, 16, 32)
			if err != nil {
				context += v
			}else{
				context += fmt.Sprintf("%c", temp)
				context += v2
			}
		}else{
			temp, err := strconv.ParseInt(v, 16, 32)
			if err != nil {
				context += v
			}else{
				context += fmt.Sprintf("%c", temp)
			}
		}
    }
	return
}
