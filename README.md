# url_monitor (A Telegraf Input Plugin, Copy from http_response)

This input plugin will test HTTP/HTTPS connections.

### Configuration:
```
# HTTP/HTTPS request given an address a method and a timeout
[[inputs.url_monitor]]
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
  ## Require String 正则表达式用单引号避免转义,如果需要包含单引号，请使用'''。如果字符串末尾有单引号'，需要换行
  require_str = '''
baidu.com
'''
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
```

### Measurements & Fields:

- url_monitor
    - response_time (float, seconds)
    - http_code (int) #The code received
	- data_match
	- time_match
	- code_match
	- require_code
	- require_str
	- require_time


### Tags:

- All measurements have the following tags:
	- app
	- cmdbid
    - url
    - method

### Example Output:

```
# ./telegraf -config url.conf -test
* Plugin: url_monitor, Collection 1
> url_monitor,app=monitor,cmdbid=1701,host=HADOOP-215,method=GET,url=http://www.baidu.com code_match=1i,data_match=1i,failed_threshold="3",http_code=200i,require_code="20\\d",require_str="baidu.com",require_time="0.0005",response_time=0.023157402,time_match=0i 1473150864000000000
```
