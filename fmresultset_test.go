package gofmcon

import (
	"encoding/xml"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

func TestFMResultsetXMLUnmarshal(t *testing.T) {
	testFile, err := os.Open("simple_test_1.xml")
	assert.NoError(t, err)

	b, err := ioutil.ReadAll(testFile)
	assert.NoError(t, err)

	fmResultSet := &FMResultset{}
	err = xml.Unmarshal(b, fmResultSet)
	assert.NoError(t, err)
}
