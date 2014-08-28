package curl

import (
	"testing"
)

func TestOptDur(t *testing.T) {
	opts := []interface{}{
		"timeout=", 10,
	}
	has, dur := optDuration("timeout=", opts)
	if !has {
		t.Fail()
	}
	if dur.Seconds() != 10 {
		t.Fail()
	}
}

func TestToFloat(t *testing.T) {
	got, f := toFloat("0.321")
	if !got {
		t.Fail()
	}
	if f != 0.321 {
		t.Fail()
	}
}

func TestOptBool(t *testing.T) {
	optsArr := make([][]interface{}, 3)
	optsArr[0] = []interface{}{
		"followredirects=", "true",
	}
	optsArr[1] = []interface{}{
		"followredirects=", 0,
	}
	optsArr[2] = []interface{}{
		"notfound=", 0,
	}
	for i, v := range optsArr {
		got, v := optBool("followredirects=", v)
		switch i {
		case 0:
			if !got {
				t.Fail()
			}
			if !v {
				t.Fail()
			}
		case 1:
			if !got {
				t.Fail()
			}
			if v {
				t.Fail()
			}
		case 2:
			if got {
				t.Fail()
			}
		}

	}
}
