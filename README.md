# go-curl
[![No Maintenance Intended](http://unmaintained.tech/badge.svg)](http://unmaintained.tech/)

for a maintained alternative, try [andelf/go-curl](https://github.com/andelf/go-curl)
## Documentation:

http://godoc.org/github.com/christophwitzko/go-curl

## Options

| Option              | Type                        | Default       |
|-------------------- |---------------------------- |-------------  |
| method=             | string                      | GET           |
| data=               | io.Reader                   | nil           |
| dialtimeout=        | time.Duration               | 0             |
| readtimeout=        | time.Duration               | 0             |
| timeout=            | time.Duration               | 0             |
| cbinterval=         | time.Duration               | time.Second   |
| followredirects=    | bool                        | true          |
| deadline=           | time.Time or time.Duration  | 0             |
| maxspeed=           | int64                       | 0             |
| header=             | http.Header                 | nil           |
| disablecompression= | bool                        | false         |

## Advanced Usage

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
cb := func (st curl.IoCopyStat) error {
  fmt.Println(st.Perstr, st.Sizestr, st.Lengthstr, st.Speedstr, st.Durstr)
  return nil
}
curl.IoCopy(r, length, w, "readtimeout=12", cb)
```
