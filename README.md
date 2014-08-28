# go-curl

> Fork from https://github.com/go-av/curl

* WITHOUT libcurl.so just using "net/http"
* Monitoring progress
* Timeouts and deadline
* Speed limit

Examples: test/test_curl.go

## Simple Usage

### Curl string or bytes
```go
import "github.com/christophwitzko/go-curl"

err, str, resp := curl.String("http://google.com")
// access response header: resp.Header["Server"]
err, b, _ := curl.Bytes("http://google.com")
```
### Save to file or writer
```go
err, resp := curl.File("http://google.com", "/tmp/index.html")

var w io.Writer
err, resp := curl.Write("http://google.com", w)
```
### With timeout (both dial timeout and read timeout set)
```go
curl.String("http://google.com", "timeout=10")
curl.String("http://google.com", "timeout=", 10)
curl.String("http://google.com", "timeout=", time.Second*10)
```
### With different dial timeout and read timeout
```go
curl.String("http://google.com", "dialtimeout=10", "readtimeout=20")
curl.String("http://google.com", "dialtimeout=", 10, "readtimeout=", time.Second*20)
```
### With deadline (if cannot download in 10s then die)
```go
curl.File("http://google.com", "index.html", "deadline=", time.Now().Add(time.Second*10))
curl.File("http://google.com", "index.html", "deadline=10")
curl.File("http://google.com", "index.html", "deadline=", 10.0)
curl.File("http://google.com", "index.html", "deadline=", time.Second*10)
```
### With speed limit
```go
curl.File("http://google.com", "index.html", "maxspeed=", 30*1024)
```
### With custom http header
```go
header := http.Header {
  "User-Agent" : {"curl/7.29.0"},
}
curl.File("http://google.com", "file", header)
```
### These params can be use in any function and in any order
```go
curl.File("http://google.com", "index.html", "timeout=", 10, header)
curl.String("http://google.com", index.html", timeout=", 10)
```
## Advanced Usage

### Get detail info
```go
var st curl.IocopyStat
curl.File("http://de.edis.at/10MB.test", "a.test", &st)
fmt.Println("size=", st.Sizestr, "average speed=", st.Speedstr, "server=", st.Header["Server"][0])
```
#### Outputs:
```
size= 9.5MB average speed= 1.9MB/s server= nginx/0.7.67
```

#### The IocopyStat struct:
```go
type IocopyStat struct {
  Stat       string         // connecting, redirect, header, downloading, finished
  Done       bool           // download is done
  Begin      time.Time      // download begin time
  Dur        time.Duration  // download elapsed time
  Per        float64        // complete percent. range 0.0 ~ 1.0
  Size       int64          // bytes downloaded
  Speed      int64          // bytes per second
  Length     int64          // content length
  Durstr     string         // pretty format of Dur. like: 10:11
  Perstr     string         // pretty format of Per. like: 3.9%
  Sizestr    string         // pretty format of Size. like: 1.1M, 3.5G, 33K
  Speedstr   string         // pretty format of Speed. like 1.1M/s
  Lengthstr  string         // pretty format of Length. like: 1.1M, 3.5G, 33K
  Response   *http.Response // response from http request
  Header     http.Header    // response header
  RedirectTo string         // redirect url (only available at Stat == redirect)
}
```
### Monitor progress
```go
curl.File(
  "http://de.edis.at/10MB.test",
  "a.test",
  func(st curl.IocopyStat) error {
    fmt.Println(st.Stat, st.Perstr, st.Sizestr, st.Lengthstr, st.Speedstr, st.Durstr)
    // return errors.New("I want to stop")
    return nil
  },
)
```
#### Outputs:
```
connecting     
header     
downloading 17.2% 1.6MB 9.5MB 1.6MB/s 0:01
downloading 38.3% 3.6MB 9.5MB 2.0MB/s 0:02
downloading 59.5% 5.7MB 9.5MB 2.0MB/s 0:03
downloading 83.9% 8.0MB 9.5MB 2.3MB/s 0:04
finished 100.0% 9.5MB 9.5MB 2.0MB/s 0:04
```
### Set monitor callback interval
```go
curl.File("xxxx", "xxx", cb, "cbinterval=", 0.5) // 0.5 second
```
### Curl in goroutine
```go
con := &curl.Control{}
go curl.File("xxx", "xxx", con)
// and then get stat
st := con.Stat()
// or stop
con.Stop()
// set max speed
con.MaxSpeed(1024*10)
// cancel max speed
con.MaxSpeed(0)
```
### Just dial
```go
err, resp := curl.Dial("http://de.edis.at/10MB.test", "timeout=11")
fmt.Println("contentLength=", resp.ContentLength)
```
## Useful Functions

### Functions format size, speed pretty
```go
curl.PrettySize(100000000)) // 95.4MB
curl.PrettyPer(0.345) // 34.5%
curl.PrettySpeed(1200) // 1.2KB/s
curl.PrettyDur(time.Second*66) // 1:06
```
### Progressed io.Copy
```go
r, _ := os.Open("infile")
w, _ := os.Create("outfile")
length := 1024*888
cb := func (st curl.IocopyStat) error {
  fmt.Println(st.Perstr, st.Sizestr, st.Lengthstr, st.Speedstr, st.Durstr)
  return nil
}
curl.IoCopy(r, length, w, "readtimeout=12", cb)
```
