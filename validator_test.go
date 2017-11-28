package validator

import (
	"fmt"
	"testing"
)

type simpleVal struct {
	min, max int
}

func getSimpleVal() Validator {
	new := func() interface{} {
		return simpleVal{
			min: 1,
			max: 10,
		}
	}
	invoke := func(v interface{}, args ...interface{}) (interface{}, error) {
		sv := v.(simpleVal)
		i := args[0].(int)
		if sv.inRange(i) {
			return i, nil
		}
		return i, fmt.Errorf("%d not in range", i)
	}
	return New(new, invoke)
}

func TestValidData(t *testing.T) {
	data := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	v := getSimpleVal()
	for _, arg := range data {
		_, err := v.Validate(arg)
		if err != nil {
			t.Fatalf("Unexpected error %+v validating value %d", err, arg)
		}
	}
}

func TestValidMulti(t *testing.T) {
	data := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	multi := [][]interface{}{}
	for _, i := range data {
		multi = append(multi, []interface{}{i})
	}
	v := getSimpleVal()
	r, err := v.ValidateMultiple(multi)
	if err != nil {
		t.Fatalf("Unexpected error %+v validating bulk: %#v", err, data)
	}
	for k, val := range r {
		ri := val.(int)
		if ri != data[k] {
			t.Fatalf("Expected %d to equal %d", ri, data[k])
		}
	}
}

func TestValidMultiWithErrs(t *testing.T) {
	data := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	multi := [][]interface{}{}
	for _, i := range data {
		multi = append(multi, []interface{}{i})
	}
	v := getSimpleVal()
	_, err := v.ValidateMultipleFullErrStack(multi)
	if len(err) != len(multi) {
		t.Fatalf("Unexpected number of error values: Expected %d, got %d", len(multi), len(err))
	}
	t.Logf("%s\n", err.Error())
}

func TestError(t *testing.T) {
	v := getSimpleVal()
	data := []int{-1, 100}
	for _, i := range data {
		if r, err := v.Validate(i); err == nil {
			t.Fatalf("Expected validation to fail for %d -> instead returned %v, %v", i, r, err)
		}
	}
}

func TestSetWithError(t *testing.T) {
	v := getSimpleVal()
	data := []int{1, 2, 3, 44, 9} // 44 is invalid
	multi := [][]interface{}{}
	for _, i := range data {
		multi = append(multi, []interface{}{i})
	}
	r, err := v.ValidateMultiple(multi)
	if err == nil {
		t.Fatal("Expected an error")
	}
	for k, val := range r {
		ri := val.(int)
		if ri != data[k] {
			t.Fatalf("Expected %d to equal %d", ri, data[k])
		}
	}
}

func (sv simpleVal) inRange(i int) bool {
	if i < sv.min || i > sv.max {
		return false
	}
	return true
}
