package theme

import (
	"fmt"
	"strings"

	"github.com/digitallyserviced/tview"
	"github.com/gdamore/tcell/v2"
	"github.com/gookit/goutil/fsutil"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/file"
)

type Theme struct {
	PrimitiveBackground tcell.Color
	HeaderBackground    tcell.Color
	GrayerBackground    tcell.Color
	SidebarBackground   tcell.Color
	SidebarLines        tcell.Color
	ContentBackground   tcell.Color
	Border              tcell.Color
	Primary             tcell.Color
	Secondary           tcell.Color
	TopbarBorder        tcell.Color
	InfoLabel           tcell.Color
	TagStyles           map[string]TagStyle
	Formats             map[string]ThemeFormatter
	FormatStrings       map[string]string
	Ansi                map[string]TagStyle
	AnsiOverride        map[string]TagStyle
}

const (
	badgeStr    = "[badgeIcon] %s  [-:-:-][badgeText] %s [-:-:-]"
	badgeStrTpl = "[badgeIcon] {{.icon}} [-:-:-][{{.color | contrast}}:{{.color}}]{{ .text | hexless }}[-:-:-]"
	priColor    = "[{{.color | contrast}}:{{.color}}:b] {{ .text | hexless }} [-:-:-]"
	twoColor    = "[{{.fg}}:{{.bg}}:b]{{.text}}[-:-:-]"
)

type TagStyle struct {
	FG, BG, Attributes string
}

var tvtheme *tview.Theme = &tview.Theme{
	PrimitiveBackgroundColor:    tcell.GetColor("#212121"),
	ContrastBackgroundColor:     tcell.ColorBlue,
	MoreContrastBackgroundColor: tcell.ColorGreen,
	BorderColor:                 tcell.ColorWhite,
	BorderFocusColor:            tcell.ColorBlue,
	TitleColor:                  tcell.ColorWhite,
	GraphicsColor:               tcell.ColorWhite,
	PrimaryTextColor:            tcell.ColorWhite,
	SecondaryTextColor:          tcell.ColorYellow,
	TertiaryTextColor:           tcell.ColorGreen,
	InverseTextColor:            tcell.ColorBlue,
	ContrastSecondaryTextColor:  tcell.ColorDarkCyan,
}

var baseXtermAnsiColorNames = []string{
	"black",
	"maroon",
	"green",
	"olive",
	"navy",
	"purple",
	"teal",
	"silver",
	"gray",
	"red",
	"lime",
	"yellow",
	"blue",
	"fuchsia",
	"aqua",
	"white",
}

var theme *Theme

var TagStyler tview.Styler = GetTagStyler(false)

func SetStyler() {
	tview.UpdateCurrentStyler(TagStyler)
}

type SetBackgroundStyler[T any] interface {
	SetBackgroundColor(tcell.Color) *T
}

var (
	merger func(a map[string]interface{}, b map[string]interface{}) ([]string, error)
	bbg    SetBackgroundStyler[tview.Box] = tview.NewBox()
)

func defineStyles() {
	bbg.SetBackgroundColor(tcell.Color100)

	if theme != nil {
		return
	}
	SetStyler()
	tview.Styles = *tvtheme
	theme = &Theme{
		PrimitiveBackground: tcell.GetColor("#212121"),
		HeaderBackground:    tcell.GetColor("#1C1C1C"),
		GrayerBackground:    tcell.GetColor("#282c34"),
		SidebarBackground:   tcell.GetColor("#21252B"),
		SidebarLines:        tcell.GetColor("#5c6370"),
		ContentBackground:   tcell.GetColor("#303030"),
		Border:              tcell.GetColor("#1C1C1C"),
		Primary:             tcell.GetColor("#4ed6aa"),
		Secondary:           tcell.GetColor("#b5d1f6"),
		TopbarBorder:        tcell.GetColor("#5c6370"),
		InfoLabel:           tcell.GetColor("#5c6370"),
		TagStyles:           make(map[string]TagStyle),
		Formats:             make(map[string]ThemeFormatter),
		FormatStrings:       make(map[string]string),
		Ansi:                make(map[string]TagStyle),
		AnsiOverride:        make(map[string]TagStyle),
	}
	ko := koanf.NewWithConf(koanf.Conf{
		StrictMerge: false,
	})
	// f := file.Provider(fsutil.Expand("~/.config/coolor/theme.toml"))
	f := file.Provider(fsutil.Expand("/work/coolors/theme.toml"))
	f.Watch(func(event interface{}, err error) {
		fmt.Println(event)
		if err != nil {
			fmt.Printf("watch error: %v", err)
			return
		}

		fmt.Println("config changed. Reloading ...")
		ko.Load(
			f,
			toml.Parser(),
			koanf.WithMergeFunc(func(src, dest map[string]interface{}) error {
				news, e := merger(src, dest)
				if e != nil {
					return e
				}
				fmt.Println("New keys", len(news), news)
				return nil
			}),
		)
		e := ko.Unmarshal("TagStyles", &theme.TagStyles)
		if e != nil {
			panic(e)
		}
		e = ko.Unmarshal("FormatStrings", &theme.FormatStrings)
		if e != nil {
			panic(e)
		}
		fmt.Println(theme.GetTheme().GetFormatString("seedText"))
		if OnConfigReloaded != nil {
			OnConfigReloaded(ko, &theme)
		}
	})
	e := ko.Load(f, toml.Parser())
	if e != nil {
		panic(e)
	}
	e = ko.Unmarshal("TagStyles", &theme.TagStyles)
	if e != nil {
		panic(e)
	}
	e = ko.Unmarshal("FormatStrings", &theme.FormatStrings)
	if e != nil {
		panic(e)
	}
	fmt.Println(theme.GetTheme().GetFormatString("seedText"))

	ResetAnsiOverrides()
}

type ConfigReloadFunc func(k *koanf.Koanf, i ...interface{})

var OnConfigReloaded ConfigReloadFunc

func init() {
	merger = func(a, b map[string]interface{}) ([]string, error) {
		news := make([]string, 0)
		for key, val := range a {
			bVal, ok := b[key]
			if !ok {
				b[key] = val
				news = append(news, key)
				continue
			}

			if _, ok := val.(map[string]interface{}); !ok {
				b[key] = val
				continue
			}

			switch v := bVal.(type) {
			case map[string]interface{}:
				newss, e := merger(val.(map[string]interface{}), v)
				if e != nil {
					return []string{}, e
				}
				news = append(news, newss...)
			default:
				b[key] = val
			}
		}
		return news, nil
	}
	defineStyles()
}

func (t *Theme) NewTagStyle(
	name string,
	args ...string,
) {
	ts := NewStyle(args...)
	t.TagStyles[name] = ts
}

func WriteDefaultStyles() {
	theme.NewTagStyle("paletteTagsIcon", "yellow", "#5c6370")
	theme.NewTagStyle("paletteTagsInfo", "black", "yellow")
	theme.NewTagStyle("fieldLabel", "#303030", "blue")
	theme.NewTagStyle("fieldInput", "blue", "#303030")
	theme.NewTagStyle("badgeIcon", "yellow", "#303030")
	theme.NewTagStyle("foreground", "white")
	theme.NewTagStyle("cursor", "orange")
	theme.NewTagStyle("background", "white", "#303030", "r")
	theme.NewTagStyle("badgeText", "blue", "black")
	theme.NewTagStyle("qcStatusInfo", "green", "black", "b")
	theme.NewTagStyle("tblSortAsc", "blue", "black", "rb")
	theme.NewTagStyle("tblSortDesc", "yellow", "black", "rb")

	theme.NewTagStyle("shortcut", "", "#343434")
	theme.NewTagStyle("shortcutIcon", "teal", "#343434")
	theme.NewTagStyle("shortcutModifier", "#343434", "teal")
	theme.NewTagStyle("shortcutLink", "teal", "red")
	theme.NewTagStyle("shortcutKeys", "", "#343434")
	theme.NewTagStyle("shortcutKey", "red", "#343434")
	theme.NewTagStyle("shortcutAction", "yellow", "#343434")

	theme.NewTagStyle("titleText", "red", "#343434")
	theme.NewTagStyle("titleIcon", "#343434", "red")

	theme.NewTagStyle("consoleIcon", "blue", "#303030")
	theme.NewTagStyle("consoleMsg", "blue", "black", "r")
	theme.NewTagStyle("consoleMsgPlugin", "green", "black", "r")
	theme.NewTagStyle("consoleMsgErr", "red", "black", "r")
	theme.NewTagStyle("consoleMsgPluginErr", "red", "black", "r")
	theme.NewTagStyle("consoleMsgWarn", "yellow", "black", "r")
	theme.NewTagStyle("consoleMsgDebug", "pink", "black", "r")
	theme.NewTagStyle("scriptConsoleMsg", "green", "black", "r")
	theme.NewTagStyle("scriptConsoleErr", "red", "black", "r")
	theme.NewTagStyle("palette_name", "white", "red", "b")
	theme.NewTagStyle("action", "red", "yellow")
	theme.NewTagStyle("list_main", "green")
	theme.NewTagStyle("list_second", "blue")
	theme.NewTagStyle("input_placeholder", "yellow", "#2c3139")
	theme.NewTagStyle("input_field", "blue", "#373e48")
	theme.NewTagStyle("input_autocomplete", "red", "#373e48")
	for _, v := range baseXtermAnsiColorNames {
		theme.NewTagStyle(fmt.Sprintf("baseColorTag%s", v), v, "", "r")
	}
}

func (t *Theme) AddFormatString(name, format string) {
	t.FormatStrings[name] = format
}

func (t *Theme) GetFormatString(name string) string {
	if str, ok := t.FormatStrings[name]; ok {
		return str
	}
	return ""
}

func (t *Theme) Get(name string) *tcell.Style {
	if sty, ok := t.TagStyles[name]; ok {
		style := tcell.StyleDefault
		if sty.FG != "" {
			style = style.Foreground(tcell.GetColor(sty.FG))
		}
		if sty.BG != "" {
			style = style.Background(tcell.GetColor(sty.BG))
		}
		if sty.Attributes != "" {
			for _, flag := range sty.Attributes {
				switch flag {
				case 'l':
					style = style.Blink(true)
				case 'b':
					style = style.Bold(true)
				case 'i':
					style = style.Italic(true)
				case 'd':
					style = style.Dim(true)
				case 'r':
					style = style.Reverse(true)
				case 'u':
					style = style.Underline(true)
				case 's':
					style = style.StrikeThrough(true)
				}
			}
		}
		return &style
	}
	return &tcell.Style{}
}

func (t *Theme) GetTheme() *Theme {
	return t
}

func (t *Theme) FixedSize(w int) string {
	return strings.Repeat(" ", w)
}

// "github.com/knadh/koanf/providers/rawbytes"
// type BoxBg SetBackgroundStyler[tview.Box]
// func (bg *BoxBg) SetBackgroundColor(tcell.Color) *tview.Box {
//   return tview.NewBox()
// }
// Throw away the old config and load a fresh copy.
// koanf.WithMergeFunc(merge func(src, dest map[string]interface{}) error)
// k := koanf.NewWithConf(koanf.Option)
// koanf.New(".").Load(toml, pa koanf.Parser, opts ...koanf.Option)
// themePath := fsutil.Expand("~/.config/coolor/theme.toml")
// if fsutil.FileExists(themePath){
//   tree, err := toml.LoadFile(themePath)
//   if err != nil {
//     panic(err)
//   }
//   // styles := tree.GetPath([]string{"TagStyles"})
//   err = tree.Unmarshal(theme)
//   if err != nil {
//     panic(err)
//   }
//
//   fmt.Println(theme)
// } else {
//   // WriteDefaultStyles()
// }
// Does the key exist in the target map?
// If no, add it and move on.
// If the incoming val is not a map, do a direct merge.
// news = append(news, key)
// The source key and target keys are both maps. Merge them.
// sb := &strings.Builder{}
// tomlenc := toml.NewEncoder(sb)
// err := tomlenc.Encode(theme)
// if err != nil {
// 	panic(err)
// }
// err = ioutil.WriteFile(fsutil.ExpandPath("~/.config/coolor/theme.toml"), []byte(sb.String()), fsutil.DefaultFilePerm)
// if err != nil {
// 	panic(err)
// }
// sb.Reset()
// "reflect"
// "gopkg.in/yaml.v2"
// "github.com/knadh/koanf"
// Styles            map[string]tcell.Style
// var theme koanf.Option
// Styles:            make(map[string]tcell.Style),
// yamlenc := yaml.NewEncoder(sb)
// err = yamlenc.Encode(theme)
// if err != nil {
// 	panic(err)
// }
// err = ioutil.WriteFile(fsutil.ExpandPath("~/.config/coolor/theme.yaml"), []byte(sb.String()), fsutil.DefaultFilePerm)
// if err != nil {
// 	panic(err)
// }
//
// func (t *Theme) SetStyleFgBgAttr(
// 	name string,
// 	fg, bg tcell.Color,
// 	attr tcell.AttrMask,
// ) *tcell.Style {
// 	sty := t.SetStyle(name)
// 	t.Styles[name] = sty.Foreground(fg).Background(bg).Attributes(attr)
// 	return sty
// }
//
// func (t *Theme) SetStyleFg(name string, fg tcell.Color) *tcell.Style {
// 	sty := t.SetStyle(name)
// 	t.Styles[name] = sty.Foreground(fg)
// 	return sty
// }
//
// func (t *Theme) SetStyleFgBg(name string, fg, bg tcell.Color) *tcell.Style {
// 	sty := t.SetStyle(name)
// 	t.Styles[name] = sty.Foreground(fg).Background(bg)
// 	return sty
// }
//
// func (t *Theme) SetStyle(name string) *tcell.Style {
// 	sty := &tcell.Style{}
// 	t.Styles[name] = *sty
// 	return sty
// }
// if sty, ok := t.Styles[name]; ok {
// 	return &sty
// }
// "fmt"
// "fmt"
// "github.com/digitallyserviced/coolors/coolor"
// "contrast": ,
// Formatt implements ThemeFormatter
// rv := append([]reflect.Value{reflect.ValueOf(tf.formatStr)}, tf.Values...)
// str := reflect.ValueOf(fmt.Sprintf).Call(rv)
// s = "FUCKeR"
// // parseFn: func(tf *ThemeFormatt, i ...interface{}) {
// // },
// formatFn: func(tf *ThemeFormatt, i ...interface{}) string {
// 	return fmt.Sprintf(tf.formatStr, i...)
// },
// Values: []reflect.Value{},
// // parseFn: func(tf *ThemeFormatt, i ...interface{}) {
// // },
// formatFn: func(tf *ThemeFormatt, i ...interface{}) string {
// 	return fmt.Sprintf(tf.formatStr, i...)
// },
// Values: []reflect.Value{},
// TwoColorBar.P
//	Theme{
//		// PrimitiveBackgroundColor:    tcell.GetColor("#101010").TrueColor(),
//	  PrimitiveBackgroundColor:    tcell.ColorBlack,
//		ContrastBackgroundColor:     tcell.ColorBlue,
//		MoreContrastBackgroundColor: tcell.ColorGreen,
//		BorderColor:                 tcell.ColorWhite,
//		BorderFocusColor:            tcell.ColorBlue,
//		TitleColor:                  tcell.ColorWhite,
//		GraphicsColor:               tcell.ColorWhite,
//		PrimaryTextColor:            tcell.ColorWhite,
//		SecondaryTextColor:          tcell.ColorYellow,
//		TertiaryTextColor:           tcell.ColorGreen,
//		InverseTextColor:            tcell.ColorBlue,
//		ContrastSecondaryTextColor:  tcell.ColorDarkCyan,
//	}
// TitleColor:                  tcell.GetColor("#5c6370"),
// ContrastBackgroundColor:     0,
// MoreContrastBackgroundColor: 0,
// BorderColor:                 0,
// BorderFocusColor:            0,
// GraphicsColor:               0,
// PrimaryTextColor:            0,
// SecondaryTextColor:          0,
// TertiaryTextColor:           0,
// InverseTextColor:            0,
// ContrastSecondaryTextColor:  0,
// func NewHexColor() tcell.Color {
//
// }
// if sty.fg != "" {
// 	newFgColor = sty.fg
// }
// var tff ThemeFormatter = BadgeFormat
// tcell.GetColor("#890a37"),
// tcell.ColorGreen,
// theme.SetStyleFgBg("action", tcell.ColorBlack, tcell.ColorYellow)
// theme.NewTagStyle("sectionSelectSelected", selCol)
// theme.NewTagStyle("consoleMsgPrefix", "", "", "r")
// theme.SetStyleFgBg("paletteTagsIcon", tcell.GetColor("red"), tcell.ColorBlack)
// theme.SetStyleFgBg("paletteTagsInfo", tcell.ColorBlack, tcell.ColorRed)
// style = style.Normal()
// div := ((2 * (n-len(s))) / 2) + 1
// tags := color.GetColorTags()
// tags["infolabel"] = RgbHex256toCode("5c6370", false)
// tags["sckey"] = RgbHex256toCode("fda47f", false)
// tags["scicon"] = RgbHex256toCode("7aa4a1", false)
// tags["scname"] = RgbHex256toCode("7aa4a1", false)
// tags["scdesc"] = RgbHex256toCode("5a93aa", false)
// tags["colorinfolabel"] = RgbHex256toCode("7aa4a1", false)
// tags["colorinfovalue"] = RgbHex256toCode("fda47f", false)
//"#cb7985"
//"#ff8349"
//"#2f3239", "#e85c51", "#7aa4a1", "#fda47f", "#5a93aa", "#ad5c7c", "#a1cdd8", "#ebebeb"
//"#4e5157", "#eb746b", "#8eb2af", "#fdb292", "#73a3b7", "#b97490", "#afd4de", "#eeeeee"
// _ = sty
// dfg, dbg, dattr := sty.Decompose()
// if dfg != 0 {
//   newFgColor = fmt.Sprintf("#%06X", dfg.Hex())
// }
// if dbg != 0 {
//   newBgColor = fmt.Sprintf("#%06X", dbg.Hex())
// }
// attrs := make([]string, 0)
// attrs = append(attrs, lo.Ternary(dattr&tcell.AttrBold > 0, "b", ""))
// attrs = append(attrs, lo.Ternary(dattr&tcell.AttrItalic > 0, "i", ""))
// attrs = append(attrs, lo.Ternary(dattr&tcell.AttrBlink > 0, "l", ""))
// attrs = append(attrs, lo.Ternary(dattr&tcell.AttrReverse > 0, "r", ""))
// attrs = append(attrs, lo.Ternary(dattr&tcell.AttrDim > 0, "d", ""))
// attrs = append(attrs, lo.Ternary(dattr&tcell.AttrUnderline > 0, "u", ""))
// attrs = append(attrs, lo.Ternary(dattr&tcell.AttrStrikeThrough > 0, "s", ""))
// newAttributes = strings.Join(attrs, "")
// fmt.Println(newFgColor, newBgColor, newAttributes)
