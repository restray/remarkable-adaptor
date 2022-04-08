package remarkableadaptor

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type ReFile struct {
	ReDocument

	FileType      string `json:"fileType"`
	FormatVersion int    `json:"formatVersion"`

	Margins            int    `json:"margins"`
	Orientation        string `json:"orientation"`
	RedirectionPageMap []int  `json:"redirectionPageMap"` /* Don't understand this.. */
	SizeInBytes        string `json:"sizeInBytes"`

	FontName      string `json:"fontName"`
	TextAlignment string `json:"textAlignment"`
	TextScale     int    `json:"textScale"`
	LineHeight    int    `json:"lineHeight"`

	Pages             []string `json:"pages"`
	OriginalPageCount int      `json:"originalPageCount"`
	CurrentPage       int      `json:"CurrentPage"`
	PageCount         int      `json:"pageCount"`
	CoverPageNumber   int      `json:"coverPageNumber"`
	DummyDocument     bool     `json:"dummyDocument"`

	ExtraMetadata struct {
		LastBallpointColor     string `json:"LastBallpointColor"`
		LastBallpointSize      string `json:"LastBallpointSize"`
		LastBallpointv2Color   string `json:"LastBallpointv2Color"`
		LastBallpointv2Size    string `json:"LastBallpointv2Size"`
		LastCalligraphyColor   string `json:"LastCalligraphyColor"`
		LastCalligraphySize    string `json:"LastCalligraphySize"`
		LastEraseSectionColor  string `json:"LastEraseSectionColor"`
		LastEraseSectionSize   string `json:"LastEraseSectionSize"`
		LastEraserColor        string `json:"LastEraserColor"`
		LastEraserSize         string `json:"LastEraserSize"`
		LastEraserTool         string `json:"LastEraserTool"`
		LastFinelinerv2Color   string `json:"LastFinelinerv2Color"`
		LastFinelinerv2Size    string `json:"LastFinelinerv2Size"`
		LastHighlighterv2Color string `json:"LastHighlighterv2Color"`
		LastHighlighterv2Size  string `json:"LastHighlighterv2Size"`
		LastMarkerv2Color      string `json:"LastMarkerv2Color"`
		LastMarkerv2Size       string `json:"LastMarkerv2Size"`
		LastPaintbrushv2Color  string `json:"LastPaintbrushv2Color"`
		LastPaintbrushv2Size   string `json:"LastPaintbrushv2Size"`
		LastPen                string `json:"LastPen"`
		LastPencilColor        string `json:"LastPencilColor"`
		LastPencilSize         string `json:"LastPencilSize"`
		LastPencilv2Color      string `json:"LastPencilv2Color"`
		LastPencilv2Size       string `json:"LastPencilv2Size"`
		LastSelectionToolColor string `json:"LastSelectionToolColor"`
		LastSelectionToolSize  string `json:"LastSelectionToolSize"`
		LastSharpPencilv2Color string `json:"LastSharpPencilv2Color"`
		LastSharpPencilv2Size  string `json:"LastSharpPencilv2Size"`
		LastTool               string `json:"LastTool"`
		LastUndefinedColor     string `json:"LastUndefinedColor"`
		LastUndefinedSize      string `json:"LastUndefinedSize"`
	} `json:"extraMetadata"`
}

type ReFolder struct {
	ReDocument
}

type ReDocument struct {
	Bookmarked     bool      `json:"Bookmarked"`
	ID             string    `json:"ID"`
	ModifiedClient time.Time `json:"ModifiedClient"`
	Parent         string    `json:"Parent"`
	Type           string    `json:"Type"`
	Version        int       `json:"Version"`
	VissibleName   string    `json:"VissibleName"`
}

type ReDocuments []ReDocument
type ReFolders []ReFolder
type ReFiles []ReFile

type ReMarkable struct {
	host          string
	Documents     ReDocuments
	Folders       []ReFolder
	FoldersCache  map[string]ReFolder
	Files         []ReFile
	currentFolder *ReFolder
	path          string
}

func (tablet *ReMarkable) setHost(providedHost string) {
	tablet.host = fmt.Sprintf("http://%s", providedHost)
}

func (folder ReDocument) String() string {
	return fmt.Sprintf("name: %s", folder.VissibleName)
}

func (tablet *ReMarkable) MoveToRoot() {
	tablet.currentFolder = nil
	tablet.path = ""
}

func (tablet *ReMarkable) MoveFolder(folder *ReFolder) error {
	if folder == nil {
		return errors.New("folder is nil")
	}
	tablet.currentFolder = folder
	tablet.path = tablet.currentFolder.ID
	if _, err := tablet.FetchDocuments(); err != nil {
		return err
	}
	return nil
}

func (tablet *ReMarkable) MoveParent() error {
	if tablet.currentFolder == nil {
		return errors.New("no parent folder")
	}

	if tablet.currentFolder.Parent == "" {
		tablet.MoveToRoot()
		if _, err := tablet.FetchDocuments(); err != nil {
			return err
		}
		return nil
	}

	if tablet.currentFolder == nil {
		return errors.New("no current folder")
	}

	cacheIndex := tablet.currentFolder.Parent
	cache := tablet.FoldersCache[cacheIndex]
	if err := tablet.MoveFolder(&cache); err != nil {
		return err
	}

	return nil
}

func (tablet *ReMarkable) GetCurrentFolder() ReFolder {
	return *tablet.currentFolder
}

func (tablet *ReMarkable) resetDocuments() {
	tablet.Documents = ReDocuments{}
	tablet.Folders = ReFolders{}
	tablet.Files = ReFiles{}
}

func (tablet *ReMarkable) appendToCache(folder ReFolder) {
	tablet.FoldersCache[folder.ID] = folder
}

func (tablet *ReMarkable) request() ([]byte, error) {
	resp, err := http.Post(tablet.host+"/documents/"+tablet.path, "text/plain", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Decode the folders
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (tablet *ReMarkable) FetchDocuments() (*ReDocuments, error) {
	// Fill the ReMarkable struct with the data from the API
	b, err := tablet.request()
	if err != nil {
		return nil, err
	}

	tablet.resetDocuments()

	var tmpFiles []ReFile
	var tmpFolders []ReFolder

	json.Unmarshal(b, &tablet.Documents)
	json.Unmarshal(b, &tmpFiles)
	json.Unmarshal(b, &tmpFolders)

	for i, v := range tablet.Documents {
		if v.Type == "CollectionType" {
			tablet.appendToCache(tmpFolders[i])
			tablet.Folders = append(tablet.Folders, tmpFolders[i])
		} else if v.Type == "DocumentType" {
			tablet.Files = append(tablet.Files, tmpFiles[i])
		}
	}

	return &tablet.Documents, nil
}

func (tablet *ReMarkable) printTree(tab int, append string) string {
	for _, file := range tablet.Files {
		append += fmt.Sprintf("%s├─ 🗒️  %s\n", strings.Repeat("|  ", tab), file.VissibleName)
	}
	for _, folder := range tablet.Folders {
		append += fmt.Sprintf("%s├─ 📂 %s/\n", strings.Repeat("|  ", tab), folder.VissibleName)
		tablet.MoveFolder(&folder)
		append = tablet.printTree(tab+1, append)
	}
	tablet.MoveParent()
	return append
}

func (tablet *ReMarkable) PrintTree() string {
	return tablet.printTree(0, "Root:\n")
}

func (tablet *ReMarkable) Load(providedHost string) (*ReMarkable, error) {
	tablet.setHost(providedHost)
	tablet.path = ""
	tablet.FoldersCache = make(map[string]ReFolder)

	if _, err := tablet.FetchDocuments(); err != nil {
		return nil, err
	}

	return tablet, nil
}
