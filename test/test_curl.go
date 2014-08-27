package main

import (
	"fmt"
	"github.com/christophwitzko/go-curl"
)

func test1(url string, opts ...interface{}) {
	var st curl.IocopyStat
	err := curl.File(
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
			fmt.Println(st.Perstr, st.Sizestr, st.Lengthstr, st.Speedstr, st.Durstr)
			return nil
		},
		"maxspeed=", 30*1000,
	)
}

func main() {
	test2(
		"http://de.edis.at/10MB.test",
		"maxspeed=", 1000)
}
