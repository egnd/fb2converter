package processor

import (
	"testing"
)

type testCase struct {
	in  string
	m   map[string]string
	out string
}

var cases = []testCase{
	{
		in: `#l #f #m`,
		m: map[string]string{
			"#l": "Last_Name",
			"#f": "First_Name",
			"#m": "Middle_Name",
		},
		out: `Last_Name First_Name Middle_Name`,
	},
	{
		in: `#l #f{ #m}`,
		m: map[string]string{
			"#l": "Last_Name",
			"#f": "First_Name",
			"#m": "Middle_Name",
		},
		out: `Last_Name First_Name Middle_Name`,
	},
	{
		in: `#l #f{ #m}`,
		m: map[string]string{
			"#l": "Last_Name",
			"#f": "First_Name",
			"#m": "",
		},
		out: `Last_Name First_Name`,
	},
	{
		in: `#l #f{ aaaaaaaaaa }`,
		m: map[string]string{
			"#l": "Last_Name",
			"#f": "First_Name",
			"#m": "Middle_Name",
		},
		out: `Last_Name First_Name`,
	},
	{
		in: `#l{ #f{ #m}}`,
		m: map[string]string{
			"#l": "Last_Name",
			"#f": "First_Name",
			"#m": "",
		},
		out: `Last_Name First_Name`,
	},
	{
		in: `#l{ \{mm\} #f{ #m}}`,
		m: map[string]string{
			"#l": "Last_Name",
			"#f": "First_Name",
			"#m": "",
		},
		out: `Last_Name {mm} First_Name`,
	},
	{
		in: `#authors #title #author`,
		m: map[string]string{
			"#author":  "_single_author_",
			"#authors": "_multiple_authors_",
			"#title":   "book-title",
		},
		out: `_multiple_authors_ book-title _single_author_`,
	},
	{
		in: `#abbrseries #ABBRseries`,
		m: map[string]string{
			"#abbrseries": "_a_b_c_",
			"#ABBRseries": "_A_B_C_",
		},
		out: `_a_b_c_ _A_B_C_`,
	},
}

func TestReplaceKeywords(t *testing.T) {

	for i, c := range cases {
		res := ReplaceKeywords(c.in, c.m)
		if res != c.out {
			t.Fatalf("BAD RESULT for case %d\nEXPECTED:\n[%s]\nGOT:\n[%s]", i+1, c.out, res)
		}
	}
	t.Logf("OK - %s: %d cases", t.Name(), len(cases))
}
