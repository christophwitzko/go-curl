package main

import (
	"fmt"
	"github.com/christophwitzko/go-curl"
	"time"
)

func test1(url string, opts ...interface{}) {
	var st curl.IocopyStat
	err, _ := curl.File(
		url,
		"a.test",
		append(opts, &st)...)
	fmt.Println(err, "size=", st.Size, "average speed=", st.Speed)
}

func test2(url string, opts ...interface{}) {
	curl.File(
		url,
		"a.test",
		append(opts,
			func(st curl.IocopyStat) error {
				fmt.Println(st.Stat, st.Perstr, st.Sizestr, st.Lengthstr, st.Speedstr, st.Durstr)
				return nil
			}, "timeout=10")...,
	)
}
func test3() {
	curl.File(
		"http://de.edis.at/10MB.test",
		"a.test",
		func(st curl.IocopyStat) error {
			//fmt.Println(st.Perstr, st.Sizestr, st.Lengthstr, st.Speedstr, st.Durstr)
			fmt.Println(st.Stat)
			fmt.Println(st.Header["Date"])
			return nil
		},
		"maxspeed=", 30*1000,
		"followredirects=", false,
	)
}

func test4() {
	err, _, resp := curl.String("http://google.com")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(resp.Header["Server"])
}

func test5() {
	var st curl.IocopyStat
	curl.File("http://de.edis.at/10MB.test", "a.test", &st)
	fmt.Println("size=", st.Sizestr, "average speed=", st.Speedstr, "server=", st.Header["Server"][0])
}

func test6() {
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

func test7() {
	cb := func(st curl.IocopyStat) error {
		switch st.Stat {
		case "downloading", "finished":
			fmt.Println("D", st.Perstr, st.Sizestr, st.Lengthstr, st.Speedstr, st.Durstr)
		case "header":
			fmt.Println("Server:", st.Header["Server"][0])
		}

		// return errors.New("I want to stop")
		return nil
	}
	curl.File("http://de.edis.at/10MB.test", "a.test", cb, "cbinterval=", 0.1, "maxspeed=", 2*1024*1024)
}

func main() {
	test7()
}
