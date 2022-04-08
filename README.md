# 📘 A ReMarkable Adaptor in Go
## 😎 Easily manage your ReMarkable

To begin, this project is a simple adaptor for ReMarkable. It provides an easy to understand interface to manage your ReMarkable Documents from your tablet!

I'm maintaining it on my free time, so it's not perfect, but it's a good start.

---
## Installation

Very easy to install:
```shell
$ go get github.com/restray/remarkable-adaptor
```

You must put your enable USB Web Interface on your ReMarkable! 
> 🗒️ Settings --> Storage --> USB Web App

---
## Usage

I personally use it to backup my tablet into my computer, to use my tablet offline and to use a Terminal CLI to push/pull files when I'm programming!
I'll add in the README the projects that use this Lib 😎

### 🤖 Go Methods

You can watch the `remarkable_test.go` to get a working example!
What you can do basically:
- Available types: 
   - File types: `ReFile ReFolder` that extend `ReDocument`
   - Files List: `ReDocuments ReFolders ReFiles`
   - Global interface: `ReMarkable`

### 🗒️ Example

Connect to the tablet API:
> ⚠️ By default, when you call `Load`, that will fetch root documents

```golang
tablet := new(ReMarkable)
tablet, err := tablet.Load("10.11.99.1")
```

Getting Documents:

```golang
/* ONLY NEEDED TO FORCE SYNC! */
tablet.FetchDocuments() // Will put current folder documents in the tablet struct

documents := tablet.Documents // Is a ReDocuments type
folders := tablet.Folders // Is a ReFolders type
files := tablet.Files // Is a ReFiles type
```

Moving to Folder:
> ⚠️ By default, when you call `MoveFolder`, that will fetch root documents

```golang
folderToMove := tablet.Folders[0]
tablet.MoveFolder(&folderToMove)
```

Getting a document tree:

```golang
fmt.Println(tablet.GetTree())
/* Will output something like
📂 Root:
├─ 🗒️  File On Root
├─ 📂 Test/
|  ├─ 🗒️  Children File
├─ 📂 GoLang/
|  ├─ 🗒️  Golang File\n
*/
```

### 🧭 REST API

⚙️ I'm currently working on it, should be release soon!
