package remarkableadaptor

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type WorkingRemarkableTestSuite struct {
	suite.Suite
	requests int
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

func (suite *WorkingRemarkableTestSuite) SetupAllSuite() {
	suite.requests = 0
}

func (suite *WorkingRemarkableTestSuite) TestLoad() {
	tablet := new(ReMarkable)
	tablet, err := tablet.Load("localhost:8080")
	suite.requests++

	suite.NoError(err)
	suite.NotNil(tablet)
	suite.IsType(*tablet, ReMarkable{})
}

func (suite *WorkingRemarkableTestSuite) TestCantLoad() {
	tablet := new(ReMarkable)
	tablet, err := tablet.Load("unexisting")

	suite.Error(err)
	suite.Nil(tablet)
}

func (suite *WorkingRemarkableTestSuite) TestFetchRootDocuments() {
	tablet := new(ReMarkable)
	tablet, err := tablet.Load("localhost:8080")
	suite.requests++

	suite.NoError(err)
	suite.Equal(len(tablet.Documents), 3)
	suite.Equal(len(tablet.Folders), 2)
	suite.Equal(len(tablet.Files), 1)
}

func (suite *WorkingRemarkableTestSuite) TestFetchDocumentResetOnRequest() {
	tablet := new(ReMarkable)
	tablet, err := tablet.Load("localhost:8080")
	suite.requests++

	suite.NoError(err)

	lenDocuments := len(tablet.Documents)
	lenFolders := len(tablet.Folders)
	lenFiles := len(tablet.Files)

	tablet.FetchDocuments()
	suite.requests++

	suite.NoError(err)
	suite.Equal(len(tablet.Documents), lenDocuments)
	suite.Equal(len(tablet.Folders), lenFolders)
	suite.Equal(len(tablet.Files), lenFiles)
}

func (suite *WorkingRemarkableTestSuite) TestMovingFolder() {
	tablet := new(ReMarkable)
	tablet, _ = tablet.Load("localhost:8080")
	suite.requests++

	testFolder := tablet.Folders[0]
	tablet.MoveFolder(&testFolder)
	suite.requests++

	suite.Equal(tablet.GetCurrentFolder().ID, testFolder.ID)
}

func (suite *WorkingRemarkableTestSuite) TestTree() {
	tablet := new(ReMarkable)
	tablet, err := tablet.Load("localhost:8080")
	suite.requests++

	suite.NoError(err)
	suite.NotNil(tablet)

	result := tablet.PrintTree()
	suite.requests += 2 * 2 // Root is already loaded from "Load", and 2 folders are loaded from "FetchDocuments" and then "MoveParent"
	suite.Equal(result, "Root:\n├─ 🗒️  Agenda\n├─ 📂 Test/\n|  ├─ 🗒️  Children File\n├─ 📂 GoLang/\n|  ├─ 🗒️  Golang File\n")
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestWorkingRemarkableTestSuite(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	requestsDone := 0

	httpmock.RegisterResponder("POST", "http://localhost:8080/documents/", func(req *http.Request) (*http.Response, error) {
		fmt.Println("[POST] Request documents root")
		requestsDone++
		return httpmock.NewStringResponse(200, getJsonFile("test_set.json")), nil
	})
	httpmock.RegisterResponder("POST", "http://localhost:8080/documents/test", func(req *http.Request) (*http.Response, error) {
		fmt.Println("[POST] Request test/ documents")
		requestsDone++
		return httpmock.NewStringResponse(200, getJsonFile("test_children_test.json")), nil
	})
	httpmock.RegisterResponder("POST", "http://localhost:8080/documents/golang", func(req *http.Request) (*http.Response, error) {
		fmt.Println("[POST] Request golang/ documents")
		requestsDone++
		return httpmock.NewStringResponse(200, getJsonFile("test_children_golang.json")), nil
	})

	testSuite := new(WorkingRemarkableTestSuite)
	suite.Run(t, testSuite)
	fmt.Println(testSuite.requests, requestsDone)
	assert.Equal(t, requestsDone, testSuite.requests)
}
