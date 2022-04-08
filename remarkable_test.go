package remarkableadaptor

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
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
	requests     int
	requestsDone int
	tmpFolder    string
}

func getJsonFile(filename string) string {
	// Open our jsonFile
	jsonFile, err := os.Open("test_files/" + filename)
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	respSrv, _ := ioutil.ReadAll(jsonFile)

	return string(respSrv)
}

func (suite *WorkingRemarkableTestSuite) SetupSuite() {
	suite.requests = 0
	suite.requestsDone = 0

	currDir, err := os.Getwd()
	if err != nil {
		log.Panicln("Can't get working dir")
	}
	dir, err := ioutil.TempDir(currDir, "test_tmp")
	if err != nil {
		log.Panicln("Can't create temp dir")
	}

	httpmock.Activate()

	httpmock.RegisterResponder("POST", "http://localhost:8080/documents/", func(req *http.Request) (*http.Response, error) {
		suite.requestsDone++
		return httpmock.NewStringResponse(200, getJsonFile("test_set.json")), nil
	})
	httpmock.RegisterResponder("POST", "http://localhost:8080/documents/test", func(req *http.Request) (*http.Response, error) {
		suite.requestsDone++
		return httpmock.NewStringResponse(200, getJsonFile("test_children_test.json")), nil
	})
	httpmock.RegisterResponder("POST", "http://localhost:8080/documents/golang", func(req *http.Request) (*http.Response, error) {
		suite.requestsDone++
		return httpmock.NewStringResponse(200, getJsonFile("test_children_golang.json")), nil
	})
	httpmock.RegisterResponder("GET", "http://localhost:8080/download/children_file/placeholder", func(req *http.Request) (*http.Response, error) {
		suite.requestsDone++
		return httpmock.NewBytesResponse(200, httpmock.File("test_files/children_file.pdf").Bytes()), nil
	})
	httpmock.RegisterResponder("GET", "http://localhost:8080/download/rootfile/placeholder", func(req *http.Request) (*http.Response, error) {
		suite.requestsDone++
		return httpmock.NewBytesResponse(200, httpmock.File("test_files/rootfile.pdf").Bytes()), nil
	})

	suite.tmpFolder = dir
}

func (suite *WorkingRemarkableTestSuite) TearDownSuite() {
	httpmock.DeactivateAndReset()
	err := os.RemoveAll(suite.tmpFolder)
	if err != nil {
		log.Panicln(suite.tmpFolder, err)
	}
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

	result := tablet.GetTree()
	suite.requests += 2 * 2 // Root is already loaded from "Load", and 2 folders are loaded from "FetchDocuments" and then "MoveParent"
	suite.Equal(result, "📂 Root:\n├─ 🗒️  File On Root\n├─ 📂 Test/\n|  ├─ 🗒️  Children File\n├─ 📂 GoLang/\n|  ├─ 🗒️  Golang File\n")
}

func (suite *WorkingRemarkableTestSuite) TestChildrenTree() {
	tablet := new(ReMarkable)
	tablet, err := tablet.Load("localhost:8080")
	suite.requests++

	suite.NoError(err)
	suite.NotNil(tablet)

	testFolder := tablet.Folders[0]
	tablet.MoveFolder(&testFolder)
	suite.requests++

	result := tablet.GetTree()
	suite.Equal(result, "📂 Test:\n├─ 🗒️  Children File\n")
}

func (suite *WorkingRemarkableTestSuite) TestDownloadFile() {
	tablet := new(ReMarkable)
	tablet, err := tablet.Load("localhost:8080")
	suite.requests++

	suite.NoError(err)
	suite.NotNil(tablet)

	err = tablet.Download(&tablet.Files[0], path.Join(suite.tmpFolder, "test.pdf"))
	suite.requests++
	suite.NoError(err)
	suite.FileExists(path.Join(suite.tmpFolder, "test.pdf"))
}

func (suite *WorkingRemarkableTestSuite) TestDownloadChildrenFile() {
	tablet := new(ReMarkable)
	tablet, err := tablet.Load("localhost:8080")
	suite.requests++

	suite.NoError(err)
	suite.NotNil(tablet)

	testFolder := tablet.Folders[0]
	tablet.MoveFolder(&testFolder)
	suite.requests++

	err = tablet.Download(&tablet.Files[0], path.Join(suite.tmpFolder, "children_file.pdf"))
	suite.requests++
	suite.NoError(err)
	suite.FileExists(path.Join(suite.tmpFolder, "children_file.pdf"))
}

func TestUploadFile(t *testing.T) {
	tablet := new(ReMarkable)
	tablet, err := tablet.Load("10.11.99.1")

	assert.NoError(t, err)
	assert.NotNil(t, tablet)

	currDir, err := os.Getwd()
	err = tablet.Upload(path.Join(currDir, "test_files/test.pdf"), "test.pdf")
	assert.NoError(t, err)
}

func TestUploadFileInFolder(t *testing.T) {
	tablet := new(ReMarkable)
	tablet, err := tablet.Load("10.11.99.1")

	assert.NoError(t, err)
	assert.NotNil(t, tablet)

	testFolder := tablet.Folders[0]
	tablet.MoveFolder(&testFolder)

	currDir, err := os.Getwd()
	err = tablet.Upload(path.Join(currDir, "test_files/test.pdf"), "test.pdf")
	assert.NoError(t, err)
}

func TestWorkingRemarkableTestSuite(t *testing.T) {
	testSuite := new(WorkingRemarkableTestSuite)
	suite.Run(t, testSuite)
	assert.Equal(t, testSuite.requestsDone, testSuite.requests)
}
