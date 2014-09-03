package curl_test

import (
	"fmt"
	"github.com/christophwitzko/go-curl"
	"io"
	"net/http"
	"strings"
	"time"
)

// Curl string or bytes.
func Example_simple() {
	err, str, resp := curl.String("http://google.com")
	if err != nil {
		return
	}
	fmt.Println(str)
	fmt.Println(resp.Header["Server"]) // access response header
	err, _, _ = curl.Bytes("http://google.com")
	if err != nil {
		return
	}
}

func ExampleIoCopyStat() {
	var st curl.IoCopyStat
	curl.File("http://de.edis.at/10MB.test", "a.test", &st)
	fmt.Println("size=", st.Sizestr, "average speed=", st.Speedstr, "server=", st.Header.Get("Server"))
	// Output: size= 9.5MB average speed= 1.9MB/s server= nginx/0.7.67
}

func ExampleIoCopyCb() {
	curl.File(
		"http://de.edis.at/10MB.test",
		"a.test",
		func(st curl.IoCopyStat) error {
			fmt.Println(st.Stat, st.Perstr, st.Sizestr, st.Lengthstr, st.Speedstr, st.Durstr)
			//fmt.Println(st.Header["Date"])
			return nil
		},
		"maxspeed=", 3*1024*1024,
		"followredirects=", false,
		"cbinterval=", 0.5, // call the callback 0.5 second
		"header=", http.Header{"User-Agent": {"curl/7.29.0"}},
	)
	// Output: connecting
	// Output: header
	// Output: downloading 17.2% 1.6MB 9.5MB 1.6MB/s 0:01
	// Output: downloading 38.3% 3.6MB 9.5MB 2.0MB/s 0:02
	// Output: downloading 59.5% 5.7MB 9.5MB 2.0MB/s 0:03
	// Output: downloading 83.9% 8.0MB 9.5MB 2.3MB/s 0:04
	// Output: finished 100.0% 9.5MB 9.5MB 2.0MB/s 0:04
}

func ExampleControl() {
	con := &curl.Control{}
	go curl.File("http://de.edis.at/10MB.test", "a.test", con)
	for {
		st := con.Stat()
		fmt.Println(st.Stat, st.Perstr)
		if st.Done {
			return
		}
		time.Sleep(1000 * time.Millisecond)
	}
}

// Save the response data to a file or a io.Writer
func Example_writer() {
	err, _ := curl.File("http://google.com", "/tmp/index.html")
	if err != nil {
		return
	}
	var w io.Writer
	err, _ = curl.Write("http://google.com", w)
	if err != nil {
		return
	}
}

// Set a timeout (both dial timeout and read timeout set)
func Example_timeout() {
	curl.String("http://google.com", "timeout=10")
	curl.String("http://google.com", "timeout=", 10)
	curl.String("http://google.com", "timeout=", time.Second*10)
}

// Set a different dial timeout and read timeout
func Example_timeout2() {
	curl.String("http://google.com", "dialtimeout=10", "readtimeout=20")
	curl.String("http://google.com", "dialtimeout=", 10, "readtimeout=", time.Second*20)
}

// Set a deadline (if cannot download in 10s then die)
func Example_deadline() {
	curl.File("http://google.com", "index.html", "deadline=", time.Now().Add(time.Second*10))
	curl.File("http://google.com", "index.html", "deadline=10")
	curl.File("http://google.com", "index.html", "deadline=", 10.0)
	curl.File("http://google.com", "index.html", "deadline=", time.Second*10)
}

// Set a speed limit
func Example_maxspeed() {
	curl.File("http://google.com", "index.html", "maxspeed=", 30*1024)
}

// Set custom HTTP header
func Example_header() {
	header := http.Header{
		"User-Agent": {"curl/7.29.0"},
	}
	curl.File("http://google.com", "file", "header=", header)
}

// All params can be use in any function and in any order.
func Example_params() {
	curl.File("http://google.com", "index.html", "timeout=", 10)
	curl.String("http://google.com", "index.html", "maxspeed=", 30*1024, "timeout=", 10)
}

// With go-curl you can use every request methode and the "data=" option takes a io.Reader.
// So you can send very easy JSON, XML, etc with your request.
func Example_post() {
	cb := func(st curl.IoCopyStat) error {
		fmt.Println(st.Stat)
		if st.Response != nil {
			fmt.Println(st.Response.Status)
		}
		return nil
	}
	err, str, resp := curl.String(
		"http://httpbin.org/post", cb, "method=", "POST",
		"data=", strings.NewReader("{\"asd\": \"test\"}"),
		"disablecompression=", true,
		"header=", http.Header{"X-My-Header": {"Gopher"}},
	)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(resp.Header)
	fmt.Println(str)
}
