package pkv

import (
	"C"
	"fmt"
	"os"

	"github.com/ebitengine/purego"
)

type PrefixTreeKeyValue struct {
	instance uint64
}

var pkv_new func(filename string, blocksize uint) uint64
var pkv_free func(uint64)
var pkv_set func(libp uint64, key string, data string)
var pkv_get func(libp uint64, key string) string

// var pkv_get func(interface{}, string) string
var libInitialized bool = false
var libc uintptr

func New(filename string, blocksize uint) *PrefixTreeKeyValue {
	var ret = PrefixTreeKeyValue{}

	initLib()

	ret.instance = pkv_new(filename, uint(len(filename)))

	return &ret

}

func initLib() {
	if !libInitialized {
		//display working folder
		fmt.Println("Working folder: ", os.Getenv("PWD"))

		var err error
		libc, err = purego.Dlopen("./pkv.so", purego.RTLD_NOW|purego.RTLD_GLOBAL)

		if err != nil {
			panic(err)
		}

		purego.RegisterLibFunc(&pkv_new, libc, "pkv_new")
		purego.RegisterLibFunc(&pkv_free, libc, "pkv_free")
		purego.RegisterLibFunc(&pkv_set, libc, "pkv_set")
		purego.RegisterLibFunc(&pkv_get, libc, "pkv_get")
		libInitialized = true
	}
}

func (p *PrefixTreeKeyValue) Set(key string, value string) {
	pkv_set(p.instance, key, value)
}

func (p *PrefixTreeKeyValue) Get(key string, defaultValue string) string {
	result := pkv_get(p.instance, key)

	if result == "" {
		return defaultValue
	}

	return result
}
