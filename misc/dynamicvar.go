package misc

import (
	"strconv"
	"strings"
)

type DynamicVar struct {
	data string
}

// Creates a new DynamicVar with a initial data. You can use the functions
// 'WithString', 'WithInt', 'WitFloat' and 'WithBool' to specify the initial
// value
func NewDynamicVar(option func(*DynamicVar)) DynamicVar {
	ret := DynamicVar{}
	option(&ret)
	return ret
}

func NewEmptyDynamicVar() DynamicVar {
	ret := DynamicVar{}
	return ret
}

func WithString(value string) func(*DynamicVar) {
	return func(dn *DynamicVar) {
		dn.SetString(value)
	}
}

func WithInt(value int64) func(*DynamicVar) {
	return func(dn *DynamicVar) {
		dn.SetInt64(value)
	}
}

func WithFloat(value float64) func(*DynamicVar) {
	return func(dn *DynamicVar) {
		dn.SetFloat64(value)
	}
}

func WithBool(value bool) func(*DynamicVar) {
	return func(dn *DynamicVar) {
		dn.SetBool(value)
	}
}

func (this *DynamicVar) SetString(value string) {
	this.data = value
}

func (this *DynamicVar) GetString() string {
	return this.data
}

func (this *DynamicVar) SetInt64(value int64) {
	this.SetString(strconv.Itoa(int(value)))
}

func (this *DynamicVar) GetInt64e() (int64, error) {
	result, err := strconv.Atoi(this.data)
	if err != nil {
		return 0, err
	}

	return int64(result), nil

}

func (this *DynamicVar) GetInt64() int64 {
	r, _ := this.GetInt64e()
	return r
}

func (this *DynamicVar) SetFloat64(value float64) {
	//strconv.FormatFloat(value, 'g', -1, 64);
	this.data = strconv.FormatFloat(value, 'f', -1, 64)
}

func (this *DynamicVar) GetFloat64e() (float64, error) {
	result, err := strconv.ParseFloat(this.data, 64)
	if err != nil {
		return 0, err
	}

	return result, nil
}

func (this *DynamicVar) GetFloat64() float64 {
	r, _ := this.GetFloat64e()
	return r

}

func (this *DynamicVar) SetBool(value bool) {
	if value {
		this.data = "1"
	} else {
		this.data = "0"
	}
}

func (this *DynamicVar) GetBool() bool {
	ret := strings.Contains("1trueyesok", strings.ToLower(this.data))
	return ret
}
