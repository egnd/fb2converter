package processor

import (
	"strings"

	"github.com/rupor-github/fb2converter/etree"
	"github.com/rupor-github/fb2converter/utils"
)

var attr = etree.NewAttr

// getAttrValue returns value  of requested attribute or empty string.
func getAttrValue(e *etree.Element, key string) string {
	a := e.SelectAttr(key)
	if a == nil {
		return ""
	}
	return a.Value
}

func extractText(e *etree.Element, head bool) string {
	res := e.Text()
	for _, c := range e.ChildElements() {
		if utils.IsOneOfIgnoreCase(c.Tag, []string{"p", "div"}) {
			res += "\n" + extractText(c, false)
		} else {
			res += extractText(c, false)
		}
	}
	res += e.Tail()
	if !head {
		return res
	}
	return strings.TrimSpace(res)
}

func getTextFragment(e *etree.Element) string {
	return extractText(e, true)
}

//nolint:deadcode,unused
func getXMLFragment(d *etree.Document) string {
	d.IndentTabs()
	s, err := d.WriteToString()
	if err != nil {
		return err.Error()
	}
	return s
}

func getXMLFragmentFromElement(e *etree.Element) string {
	d := etree.NewDocument()
	d.WriteSettings = etree.WriteSettings{CanonicalText: true, CanonicalAttrVal: true}
	d.SetRoot(e.Copy())
	d.IndentTabs()
	s, err := d.WriteToString()
	if err != nil {
		return err.Error()
	}
	return s
}
