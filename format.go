package theme

import (
	"fmt"
	"reflect"
	"text/template"

	"github.com/gookit/goutil/strutil"
)

var FormatFuncs = template.FuncMap{}

type ThemeFormatter interface {
	Formatt() string
}

func (tf *ThemeFormat) Valid() (valid bool) {
	valid = true
	for i, v := range append(tf.BaseValues, tf.Values...) {
		if i >= len(tf.Required)-1 {
			continue
		}
		if tf.Required[i] != v.Kind() {
			valid = false
		}
	}
	return
}

var FormatStrings map[string]string

type ThemeFormatterFunc func(format string, i ...interface{}) string

type (
	FormatFn      func(tf *ThemeFormatt, i ...interface{}) string
	FormatParseFn func(tf *ThemeFormat, i ...interface{})
)

type ThemeFormatt struct {
	formatStr string
	PreFn     FormatParseFn
	formatFn  FormatFn
	Required  []reflect.Kind
	Keys      []string
}

type ThemeFormat struct {
	*ThemeFormatt
	BaseValues []reflect.Value
	Values     []reflect.Value
}

func (tf *ThemeFormat) formatFn() (s string) {
	vvals := append(tf.BaseValues, tf.Values...)
	vals := make(map[string]string)
	for i, v := range tf.Keys {
		str, _ := strutil.AnyToString(vvals[i].Interface(), false)
		vals[v] = str
	}
	str := strutil.RenderText(tf.formatStr, vals, FormatFuncs)
	fmt.Println("HOTSHIT", str)
	s = str
	return
}

func (tf *ThemeFormat) parseFn(i ...interface{}) {
	for n, v := range i {
		x := reflect.ValueOf(v)
		if x.Kind() == tf.Required[n+len(tf.BaseValues)-1] {
			tf.Values = append(tf.Values, x)
		}
	}
	return
}

func (tf *ThemeFormat) AddValues(i ...interface{}) {
	tf.parseFn(i...)
}

func (tf *ThemeFormat) Formatt(i ...interface{}) (s string) {
	fmt.Println("FUCKeR")
	tf.parseFn(i...)
	if tf.Valid() {
		if tf.PreFn != nil {
			tf.PreFn(tf, i...)
		}
		s = tf.formatFn()
		tf.Values = make([]reflect.Value, 0)
	}
	return
}
var (
	ColorSet = &ThemeFormatt{
		formatStr: twoColor,
		Required: []reflect.Kind{
			reflect.String,
			reflect.String,
			reflect.String,
		},
		Keys: []string{"text", "fg", "bg"},
	}
	BadgeFormat = &ThemeFormatt{
		formatStr: priColor,
		Required: []reflect.Kind{
			reflect.String,
			reflect.String,
			reflect.String,
		},
		Keys: []string{"icon", "color", "text"},
	}
	TwoColorBar = &ThemeFormat{
		ThemeFormatt: ColorSet,
		BaseValues:   []reflect.Value{reflect.ValueOf("▀▀▀▀")},
	}
	CSSColorBadge = &ThemeFormat{
		ThemeFormatt: BadgeFormat,
		BaseValues:   []reflect.Value{reflect.ValueOf("#")},
	}
)
