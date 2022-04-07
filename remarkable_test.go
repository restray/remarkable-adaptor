package remarkableadaptor

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	tablet := new(ReMarkable)

	tablet, err := tablet.Load("10.11.99.1")

	assert.NoError(t, err)
	assert.NotNil(t, tablet)
	assert.IsType(t, *tablet, ReMarkable{})
}

func TestCantLoad(t *testing.T) {
	tablet := new(ReMarkable)

	tablet, err := tablet.Load("unexisting")

	assert.Error(t, err)
	assert.Nil(t, tablet)
}

func TestFetchRootDocuments(t *testing.T) {
	tablet := new(ReMarkable)

	tablet, err := tablet.Load("10.11.99.1")

	assert.NoError(t, err)
	assert.Greater(t, len(tablet.Documents), 0)
	assert.Greater(t, len(tablet.Folders), 0)
	assert.Greater(t, len(tablet.Files), 0)
}

func TestFetchDocumentResetOnRequest(t *testing.T) {
	tablet := new(ReMarkable)

	tablet, err := tablet.Load("10.11.99.1")
	assert.NoError(t, err)

	lenDocuments := len(tablet.Documents)
	lenFolders := len(tablet.Folders)
	lenFiles := len(tablet.Files)

	tablet.FetchDocuments()

	assert.NoError(t, err)
	assert.Equal(t, len(tablet.Documents), lenDocuments)
	assert.Equal(t, len(tablet.Folders), lenFolders)
	assert.Equal(t, len(tablet.Files), lenFiles)
}

func TestMovingFolder(t *testing.T) {
	tablet := new(ReMarkable)
	tablet.Load("10.11.99.1")

	testFolder := tablet.Folders[0]

	tablet.MoveFolder(&testFolder)

	assert.Equal(t, tablet.GetCurrentFolder().ID, testFolder.ID)
}

func getJsonFile(filename string) string {
	// Open our jsonFile
	jsonFile, err := os.Open(filename)
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	respSrv, _ := ioutil.ReadAll(jsonFile)

	return string(respSrv)
}

func TestTree(t *testing.T) {

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", "http://localhost:8080/documents/", httpmock.NewStringResponder(200, getJsonFile("test_set.json")))
	httpmock.RegisterResponder("POST", "http://localhost:8080/documents/test", httpmock.NewStringResponder(200, getJsonFile("test_children_test.json")))
	httpmock.RegisterResponder("POST", "http://localhost:8080/documents/golang", httpmock.NewStringResponder(200, getJsonFile("test_children_golang.json")))

	tablet := new(ReMarkable)
	tablet, err := tablet.Load("localhost:8080")

	assert.NoError(t, err)
	assert.NotNil(t, tablet)

	assert.Equal(t, tablet.PrintTree(), "Root:\n├─ 🗒️  Agenda\n├─ 📂 Test/\n|  ├─ 🗒️  Children File\n├─ 📂 GoLang/\n|  ├─ 🗒️  Golang File\n")
}
