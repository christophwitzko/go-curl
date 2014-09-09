package curl

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// IoCopyStat is a struct that contains all information about the IoCopy progress.
type IoCopyStat struct {
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
	Sizestr    string         // pretty format of Size. like: 1.1MB, 3.5GB, 33KB
	Speedstr   string         // pretty format of Speed. like 1.1MB/s
	Lengthstr  string         // pretty format of Length. like: 1.1MB, 3.5GB, 33KB
	Response   *http.Response // response from http request
	Header     http.Header    // response header
	RedirectTo string         // redirect url (only available at Stat == "redirect")
	Intv       int64          //cb interval
}

// Control is a Controller for a curl operation.
type Control struct {
	stop     bool
	maxSpeed int64
	st       *IoCopyStat
}

// Representing the monitoring callback.
type IoCopyCb func(st IoCopyStat) error

func (c *Control) Stop() {
	c.stop = true
}

func (c *Control) Stat() IoCopyStat {
	if c.st == nil {
		return *&IoCopyStat{Stat: "connecting"}
	}
	c.st.update()
	return *c.st
}

// Change the maxSpeed during the download process.
func (c *Control) MaxSpeed(s int64) {
	c.maxSpeed = s
}

func toFloat(o interface{}) (can bool, f float64) {
	var err error
	f, err = strconv.ParseFloat(fmt.Sprintf("%v", o), 64)
	can = (err == nil)
	return
}

func optGet(name string, opts []interface{}) (got bool, val interface{}) {
	for i, o := range opts {
		switch o.(type) {
		case string:
			stro := o.(string)
			if strings.HasPrefix(stro, name) {
				if len(stro) == len(name) {
					if i+1 < len(opts) {
						val = opts[i+1]
					}
				} else {
					val = stro[len(name):]
				}
				got = true
				return
			}
		}
	}
	return
}

func optDuration(name string, opts []interface{}) (got bool, dur time.Duration) {
	var val interface{}
	var f float64
	if got, val = optGet(name, opts); !got {
		return
	}
	if dur, got = val.(time.Duration); got {
		return
	}
	if got, f = toFloat(val); !got {
		return
	}
	dur = time.Duration(float64(time.Second) * f)
	return
}

func optTime(name string, opts []interface{}) (got bool, tm time.Time) {
	var val interface{}
	got, val = optGet(name, opts)
	tm, got = val.(time.Time)
	return
}

func optInt64(name string, opts []interface{}) (got bool, i int64) {
	var val interface{}
	var f float64
	got, val = optGet(name, opts)
	if got, f = toFloat(val); !got {
		return
	}
	i = int64(f)
	return
}

func optIntv(opts []interface{}) (intv time.Duration) {
	var hasintv bool
	hasintv, intv = optDuration("cbinterval=", opts)
	if !hasintv {
		intv = time.Second
	}
	return
}

func optString(name string, opts []interface{}) (got bool, s string) {
	var val interface{}
	got, val = optGet(name, opts)
	if got {
		s = fmt.Sprintf("%v", val)
	}
	return
}

func optBool(name string, opts []interface{}) (bool, bool) {
	got, val := optString(name, opts)
	if got {
		i, err := strconv.ParseBool(val)
		if err == nil {
			return true, i
		}
	}
	return false, false
}

func (st *IoCopyStat) update() {
	if st.Length > 0 {
		st.Per = float64(st.Size) / float64(st.Length)
	}
	st.Speed *= (int64(time.Second) / st.Intv)
	st.Dur = time.Since(st.Begin)
	st.Perstr = PrettyPer(st.Per)
	st.Sizestr = PrettySize(st.Size)
	st.Lengthstr = PrettySize(st.Length)
	st.Speedstr = PrettySpeed(st.Speed)
	st.Durstr = PrettyDur(st.Dur)
}

func (st *IoCopyStat) finish() {
	dur := float64(time.Since(st.Begin)) / float64(time.Second)
	st.Speed = int64(float64(st.Size) / dur)
	st.Per = 1.0
	st.Done = true
	st.Stat = "finished"
	st.update()
}

type mywriter struct {
	io.Writer
	n     int64
	curn  int64
	maxn  int64
	maxtm time.Time
}

func (m *mywriter) Write(p []byte) (n int, err error) {
	n, err = m.Writer.Write(p)
	m.n += int64(n)
	m.curn += int64(n)
	if m.maxn != 0 && m.curn > m.maxn {
		time.Sleep(m.maxtm.Sub(time.Now()))
	}
	return
}

func IoCopy(r io.ReadCloser, length int64, w io.Writer, opts ...interface{}) (err error) {
	var st *IoCopyStat
	var cb IoCopyCb
	var ct *Control
	var resp *http.Response

	for _, o := range opts {
		switch o.(type) {
		case *IoCopyStat:
			st = o.(*IoCopyStat)
		case *Control:
			ct = o.(*Control)
		case *http.Response:
			resp = o.(*http.Response)
		case func(IoCopyStat) error:
			cb = o.(func(IoCopyStat) error)
		}
	}

	myw := &mywriter{Writer: w}
	if st == nil {
		st = &IoCopyStat{}
	}

	if ct == nil {
		ct = &Control{st: st}
	}

	if ct.st == nil {
		ct.st = st
	}

	var rto time.Duration
	var hasrto bool
	hasrto, rto = optDuration("readtimeout=", opts)
	if !hasrto {
		hasrto, rto = optDuration("timeout=", opts)
	}

	var deadtm time.Time
	var deaddur time.Duration
	var hasdeadtm bool
	var hasdeaddur bool
	hasdeadtm, deadtm = optTime("deadline=", opts)
	if !hasdeadtm {
		hasdeaddur, deaddur = optDuration("deadline=", opts)
	}
	if hasdeaddur {
		hasdeadtm = true
		deadtm = time.Now().Add(deaddur)
	}

	intv := optIntv(opts)

	_, ct.maxSpeed = optInt64("maxspeed=", opts)

	st.Stat = "downloading"
	if resp != nil {
		st.Response = resp
		st.Header = resp.Header
	}
	st.Begin = time.Now()
	st.Length = length
	st.Intv = int64(intv)

	done := make(chan int, 0)
	go func() {
		if ct.maxSpeed == 0 {
			_, err = io.Copy(myw, r)
		} else {
			tm := time.Now()
			for {
				var nn int64
				nn, err = io.CopyN(myw, r, ct.maxSpeed)
				dur := time.Since(tm)
				if dur < time.Second {
					time.Sleep(time.Second - dur)
				}
				tm = time.Now()
				if nn != ct.maxSpeed || err != nil {
					break
				}
			}
		}
		if err == io.EOF {
			err = nil
		}
		done <- 1
	}()

	defer r.Close()

	var n, idle int64

	myw.maxn = ct.maxSpeed * int64(intv) / int64(time.Second)
	for {
		myw.maxtm = time.Now().Add(intv)
		myw.curn = 0
		select {
		case <-done:
			st.Size = myw.n
			st.Speed = myw.n - n
			st.finish()
			if cb != nil {
				err = cb(ct.Stat())
			}
			if err != nil {
				return
			}
			return
		case <-time.After(intv):
			if ct.stop {
				err = errors.New("user stops")
				return
			}
			st.Size = myw.n
			st.Speed = myw.n - n
			if cb != nil {
				err = cb(ct.Stat())
			}
			if err != nil {
				return
			}
			if myw.n != n {
				n = myw.n
				idle = 0
			} else {
				idle++
			}
			if hasrto && time.Duration(idle)*intv > rto {
				err = errors.New("read timeout")
				return
			}
			if hasdeadtm && time.Now().After(deadtm) {
				err = errors.New("deadline reached")
				return
			}
		}
	}

	return
}

func Dial(url string, opts ...interface{}) (err error, retResp *http.Response) {
	var req *http.Request
	var cb IoCopyCb

	hasmet, method := optString("method=", opts)
	if !hasmet {
		method = "GET"
	}
	var reqBody io.Reader = nil
	hasdata, rbdy := optGet("data=", opts)
	if hasdata {
		reqBody = rbdy.(io.Reader)
	}
	method = strings.ToUpper(method)
	req, err = http.NewRequest(method, url, reqBody)
	if err != nil {
		return
	}

	for _, o := range opts {
		switch o.(type) {
		case func(IoCopyStat) error:
			cb = o.(func(IoCopyStat) error)
		}
	}

	hasdto, dto := optDuration("dialtimeout=", opts)
	if !hasdto {
		hasdto, dto = optDuration("timeout=", opts)
	}

	intv := optIntv(opts)

	var header http.Header
	hasheader, iheader := optGet("header=", opts)
	if oheader, tcok := iheader.(http.Header); hasheader && tcok {
		header = oheader
	}
	req.Header = header

	var resp *http.Response

	callcb := func(st IoCopyStat) bool {
		if cb != nil {
			err = cb(st)
		}
		return err != nil
	}

	hasdiscomp, disablecompression := optBool("disablecompression=", opts)
	if !hasdiscomp {
		disablecompression = false
	}
	tr := &http.Transport{
		DisableCompression: disablecompression,
		Dial: func(network, addr string) (c net.Conn, e error) {
			if hasdto {
				c, e = net.DialTimeout(network, addr, dto)
			} else {
				c, e = net.Dial(network, addr)
			}
			return
		},
	}

	hasfolred, followredirects := optBool("followredirects=", opts)
	if !hasfolred {
		followredirects = true
	}
	client := &http.Client{
		Transport: tr,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if followredirects {
				if len(via) >= 10 {
					return errors.New("stopped after 10 redirects")
				}
				if callcb(IoCopyStat{Stat: "redirect", RedirectTo: req.URL.String()}) {
					return errors.New("user aborted")
				}
				return nil
			}
			return errors.New("following redirects not allowed")
		},
	}

	done := make(chan int, 1)
	go func() {
		defer func() {
			recover()
		}()
		resp, err = client.Do(req)
		done <- 1
	}()

	starttm := time.Now()

	if callcb(IoCopyStat{Stat: "connecting"}) {
		return
	}
out:
	for {
		select {
		case <-done:
			break out
		case <-time.After(intv):
			if hasdto && time.Now().After(starttm.Add(dto)) {
				err = errors.New("dial timeout")
				return
			}
			if callcb(IoCopyStat{Stat: "connecting"}) {
				return
			}
		}
	}

	if err != nil {
		return
	}

	if callcb(IoCopyStat{Stat: "header", Response: resp, Header: resp.Header}) {
		return
	}
	retResp = resp
	return
}

func String(url string, opts ...interface{}) (err error, body string, resp *http.Response) {
	var b bytes.Buffer
	err, resp = Write(url, &b, opts...)
	body = string(b.Bytes())
	return
}

func Bytes(url string, opts ...interface{}) (err error, body []byte, resp *http.Response) {
	var b bytes.Buffer
	err, resp = Write(url, &b, opts...)
	body = b.Bytes()
	return
}

func File(url string, path string, opts ...interface{}) (err error, resp *http.Response) {
	var w io.WriteCloser
	w, err = os.Create(path)
	if err != nil {
		return
	}
	defer w.Close()
	err, resp = Write(url, w, opts...)
	return
}

func Write(url string, w io.Writer, opts ...interface{}) (err error, resp *http.Response) {
	err, resp = Dial(url, opts...)
	if err != nil {
		return
	}
	err = IoCopy(resp.Body, resp.ContentLength, w, append(opts, resp)...)
	return
}

func PrettyDur(dur time.Duration) string {
	d := float64(dur) / float64(time.Second)
	if d < 3600 {
		return fmt.Sprintf("%d:%.2d", int(d/60), int(d)%60)
	}
	return fmt.Sprintf("%d:%.2d:%.2d", int(d/3600), int(d/60)%60, int(d)%60)
}

func PrettyPer(f float64) string {
	return fmt.Sprintf("%.1f%%", f*100)
}

func prettySize(_size interface{}, mul float64, tag []string) string {
	if len(tag) < 4 {
		return ""
	}
	var size float64
	switch _size.(type) {
	case int64:
		size = float64(_size.(int64))
	case int:
		size = float64(_size.(int))
	case float64:
		size = _size.(float64)
	default:
		return ""
	}
	size *= mul

	isc := 0
	for {
		if isc > 2 || size < 1024 {
			break
		}
		size = size / 1024.0
		isc++
	}
	return fmt.Sprintf("%.1f%s", size, tag[isc])

}

func PrettySize2(size interface{}) string {
	return prettySize(size, 8, []string{
		"Bits", "KBits", "MBits", "GBits",
	})
}

func PrettySize(size interface{}) string {
	return prettySize(size, 1, []string{
		"B", "KB", "MB", "GB",
	})
}

func PrettySpeed(s int64) string {
	return fmt.Sprintf("%s/s", PrettySize(s))
}
