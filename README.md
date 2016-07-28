# url_monitor (A Telegraf Input Plugin, Copy from http_response)

This input plugin will test HTTP/HTTPS connections.

### Configuration:


```
# HTTP/HTTPS request given an address a method and a timeout

[[inputs.url_monitor]]
  ## Server address (default http://localhost)
  address = "http://github.com"
  ## Set response_timeout (default 5 seconds)
  response_timeout = "5s"
  ## HTTP Request Method
  method = "GET"
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
$ ./telegraf -config telegraf.conf -input-filter url_monitor -test
url_monitor,method=GET,server=http://www.github.com url_monitor_code=200i,response_time=6.223266528 1459419354977857955
```
