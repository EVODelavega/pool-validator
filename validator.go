package validator

import (
	"fmt"
	"strings"
	"sync"
)

// ValidateCallback - the callback to perform type assertions so Validate works as expected
// convenience is the main reason for this type
type ValidateCallback func(v interface{}, args ...interface{}) (interface{}, error)

// Validator - the exposed interface
type Validator interface {
	Validate(...interface{}) (interface{}, error)
	ValidateMultiple([][]interface{}) ([]interface{}, error)
	AddValidators(...interface{})
	ValidateMultipleFullErrStack([][]interface{}) ([]interface{}, ErrStack)
}

// ErrStack - type returned when validating multple data-sets
type ErrStack []error

// internal validator -> contains pool + callback to apply to all relevant objects
type poolValidator struct {
	pool   *sync.Pool
	invoke ValidateCallback
}

// New - Create new pooled validator
func New(n func() interface{}, i ValidateCallback) Validator {
	return &poolValidator{
		pool: &sync.Pool{
			New: n,
		},
		invoke: i,
	}
}

// AddValidators - Puts validators into pool - be careful with this one!
func (pv *poolValidator) AddValidators(vs ...interface{}) {
	for _, v := range vs {
		pv.pool.Put(v)
	}
}

// Validate - Validates given arguments on pooled validator
// the validator is automatically returned to the pool
func (pv *poolValidator) Validate(args ...interface{}) (i interface{}, err error) {
	v := pv.pool.Get()
	i, err = pv.invoke(v, args...)
	pv.pool.Put(v)
	return
}

// ValidateMultiple - Validate multiple data-sets. This is roughly equivalent of calling Validate
// in a loop, but it doens't get a new validator from the pool each time.
func (pv *poolValidator) ValidateMultiple(margs [][]interface{}) ([]interface{}, error) {
	ret, err := pv.multi(margs, false)
	if len(err) > 0 {
		return ret, ErrStack(err)
	}
	return ret, nil
}

// ValidateMultipleFullErrStack - Same as ValidateMultiple, only this time the full ErrStack is returned, including nil errors
// this allows you to easily work out which data-set caused the error
func (pv *poolValidator) ValidateMultipleFullErrStack(margs [][]interface{}) ([]interface{}, ErrStack) {
	ret, err := pv.multi(margs, true)
	return ret, ErrStack(err)
}

// actual implementation of ValidateMultiple
func (pv *poolValidator) multi(margs [][]interface{}, allErrs bool) ([]interface{}, []error) {
	v := pv.pool.Get()
	res := make([]interface{}, 0, len(margs))
	errs := make([]error, 0, len(margs))
	for _, args := range margs {
		i, err := pv.invoke(v, args...)
		res = append(res, i)
		if allErrs || err != nil {
			errs = append(errs, err)
		}
	}
	pv.pool.Put(v)
	return res, errs
}

// Error - implement built-in error interface on ErrStack
func (es ErrStack) Error() string {
	str := make([]string, 0, len(es))
	for _, e := range es {
		str = append(str, fmt.Sprintf("%+v", e))
	}
	return strings.Join(str, "\n")
}
