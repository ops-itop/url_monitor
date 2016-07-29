# url_monitor (A Telegraf Input Plugin, Copy from http_response)

This input plugin will test HTTP/HTTPS connections.

### Configuration:


```
# HTTP/HTTPS request given an address a method and a timeout
[[inputs.url_monitor]]
  ## App Name
  app = "monitor"
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
  ## HTTP Request Headers (all values must be strings)
  # [inputs.url_monitor.headers]
  #   Host = "github.com"
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
```

### Measurements & Fields:


- url_monitor
    - response_time (float, seconds)
    - url_monitor_code (int) #The code received

### Tags:

- All measurements have the following tags:
    - server
    - method

### Example Output:

```
# ./telegraf -config url.conf -test
* Plugin: url_monitor, Collection 1
> url_monitor,app=monitor,host=cn.monitor,method=GET,url=http://www.baidu.com http_code=200i,require_code=true,require_match=true,response_time=0.032829802000000005 1469817768000000000
```
