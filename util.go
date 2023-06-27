package theme

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/digitallyserviced/tview"
	"github.com/gdamore/tcell/v2"
	"github.com/gookit/color"
)

func ResetAnsiOverrides() {
	theme.AnsiOverride = make(map[string]TagStyle)
}
func NewStyle(
	args ...string,
) (ts TagStyle) {
	ts = TagStyle{
		FG:         "",
		BG:         "",
		Attributes: "",
	}

	if len(args) > 0 {
		ts.FG = args[0]
	}
	if len(args) > 1 {
		ts.BG = args[1]
	}
	if len(args) > 2 {
		ts.Attributes = args[2]
	}
	return
}

func Jright(s string, n int) string {
	if n < 0 {
		n = 0
	}
	return strings.Repeat(" ", n) + s
}

func Jleft(s string, n int) string {
	if n < 0 {
		n = 0
	}
	return s + strings.Repeat(" ", n)
}

func Jcenter(s string, n int) string {
	if n < 0 {
		n = 0
	}
	rem := n - len(s)
	rem = ((rem * 2) / 2) / 2
	rem = rem
	return strings.Repeat(" ", rem) + s + strings.Repeat(" ", rem)
}


func GetTagStyle(fg string, ansi ...bool) (TagStyle, bool) {
	if len(ansi) > 0 && ansi[0] {
		if sty, ok := theme.AnsiOverride[fg]; ok {
			return sty, true
		}
	}
	if sty, ok := theme.TagStyles[fg]; ok {
		return sty, true
	}
	return TagStyle{}, false
}

func GetTagStyler(ansi bool) tview.Styler {
	return func(fg, bg, attr string) (newFgColor string, newBgColor string, newAttributes string) {
		newFgColor = fg
		newBgColor = bg
		newAttributes = attr
		if sty, ok := GetTagStyle(fg, ansi); ok {
			if sty.FG != "" {
				newFgColor = sty.FG
			}
			if sty.BG != "" {
				newBgColor = sty.BG
			}
			if sty.Attributes != "" {
				newAttributes = sty.Attributes
			}
		}
		if sty, ok := GetTagStyle(bg, ansi); ok {
			if sty.FG != "" {
				newBgColor = sty.FG
			}
			if sty.Attributes != "" {
				newAttributes = sty.Attributes
			}
		}
		return
	}
}

func GetTagFromStyle(sty tcell.Style) (tag string) {
	fg, bg, attr := GetTagStyleArgsFromStyle(sty)
	var colon string
	if len(attr) > 0 {
		colon = ":"
	}
	if len(fg) > 0 || len(bg) > 0 || len(attr) > 0 {
		tag = fmt.Sprintf("[%s:%s%s%s]", fg, bg, colon, attr)
	}
	return
}

func GetTagStyleArgsFromStyle(sty tcell.Style) (fg, bg, attrs string) {
	sfg, sbg, sattr := sty.Decompose()

	fg = sfg.String()
	bg = sbg.String()

	for i, v := range tcell.Attrs {
		if sattr&v > 0 {
			attrs += string(tcell.AttrsTagChar[i])
		}
	}
	return
}

func GetTheme() *Theme {
	SetStyler()

	return theme
}

const (
	TplFgRGB = "38;2;%d;%d;%d"
	TplBgRGB = "48;2;%d;%d;%d"
	FgRGBPfx = "38;2;"
	BgRGBPfx = "48;2;"
  TplFg256 = "38;5;%d"
  TplBg256 = "48;5;%d"
  Fg256Pfx = "38;5;"
  Bg256Pfx = "48;5;"
)

var (
	rxNumStr  = regexp.MustCompile("^[0-9]{1,3}$")
	rxHexCode = regexp.MustCompile("^#?([0-9a-fA-F]{3}|[0-9a-fA-F]{6})$")
)

func RgbHex256toCode(val string, isBg bool) (code string) {
	if len(val) == 6 && rxHexCode.MatchString(val) { // hex: "fc1cac"
		code = color.HEX(val, isBg).String()
	} else if strings.ContainsRune(val, ',') { // rgb: "231,178,161"
		code = strings.Replace(val, ",", ";", -1)
		if isBg {
			code = BgRGBPfx + code
		} else {
			code = FgRGBPfx + code
		}
	} else if len(val) < 4 && rxNumStr.MatchString(val) { // 256 code
		if isBg {
			code = Bg256Pfx + val
		} else {
			code = Fg256Pfx + val
		}
	}
	return
}
