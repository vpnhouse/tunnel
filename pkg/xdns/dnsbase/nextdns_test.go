package dnsbase

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	domainRules = map[string]string{
		"*.a":          "cat1",
		"*.dom.a":      "cat2",
		"sub1.dom.a":   "cat3",
		"sub2.dom.a":   "cat4",
		"*.b":          "cat5",
		"*.dom.b":      "cat6",
		"sub1.dom.b":   "cat7",
		"sub2.dom.b":   "cat8",
		"foo.bar.net.": "cat777",
	}

	invalidRecordsOverride = map[string][]string{
		// try to override existing wildcard rec
		"*.a": {"cat1new"},
		// try to override existing exact-match record
		"sub1.dom.a": {"cat3new"},
		// must never insert a record without categories
		"ru.com": nil,
		// must never insert a record with invalid domain
		"ru": {"cat"},
		// must never block the whole internet I suppose...
		"*": {"cat"},
	}

	checkResults = map[string]string{
		"sub33.a":                   "cat1",
		"sub1.sub2.sub3.sub4.a":     "cat1",
		"sub1.sub2.sub3.sub4.dom.a": "cat2",
		"sub33.dom.a":               "cat2",
		"subsub.sub33.dom.a":        "cat2",
		"sub1.dom.a":                "cat3",
		"sub2.dom.a":                "cat4",
		"sub44.b":                   "cat5",
		"sub44.dom.b":               "cat6",
		"sub1.dom.b":                "cat7",
		"sub2.dom.b":                "cat8",
		"foo.bar.net":               "cat777",
		"foo.bar.net.":              "cat777",
	}

	nonExistentDomains = []string{
		"a",
		"b",
		"dom.a",
		"dom.b",
		"sub1.dom.c",
		"sub2.dom.d",
	}
)

func dbPath() (string, func()) {
	dir, err := ioutil.TempDir(os.TempDir(), "dnsbase")
	if err != nil {
		panic(err)
	}

	closer := func() {
		_ = os.RemoveAll(dir)
	}

	return filepath.Join(dir, "db.sqlite3"), closer
}

func TestDuplicates_longFirst(t *testing.T) {
	p, closer := dbPath()
	defer closer()

	wr, err := NewWriter(p)
	require.NoError(t, err)

	err = wr.Write("pre.foo.bar", []string{"a"})
	require.NoError(t, err)
	err = wr.Write("foo.bar", []string{"b"})
	require.NoError(t, err)
}

func TestDuplicates_shortFirst(t *testing.T) {
	p, closer := dbPath()
	defer closer()

	wr, err := NewWriter(p)
	require.NoError(t, err)

	err = wr.Write("foo.bar", []string{"b"})
	require.NoError(t, err)
	err = wr.Write("pre.foo.bar", []string{"a"})
	require.NoError(t, err)
}

func TestEnd2End(t *testing.T) {
	p, closer := dbPath()
	defer closer()

	// XXX writer side
	wr, err := NewWriter(p)
	require.NoError(t, err)

	for rule, category := range domainRules {
		err = wr.Write(rule, []string{category})
		assert.Nil(t, err, "rule = %s, cat = %v", rule, category)
	}

	// nasty operations must never pass
	for dom, cats := range invalidRecordsOverride {
		err = wr.Write(dom, cats)
		assert.Error(t, err)
	}

	err = wr.Close()
	require.NoError(t, err)

	// XXX reader side
	rd, err := NewReader(p)
	require.NoError(t, err)

	for domain, expectedCategory := range checkResults {
		dom, err := rd.Lookup(domain)
		assert.Nil(t, err)
		assert.NotNil(t, dom)
		assert.Contains(t, dom.Categories, expectedCategory)
	}

	for _, dom := range nonExistentDomains {
		res, err := rd.Lookup(dom)
		assert.Error(t, err, "given %s", dom)
		fmt.Println(res)
	}
}
