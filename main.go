package main

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"sort"
	"time"
)

type FileInfo []fs.FileInfo

func (fileInf FileInfo) Less(i, j int) bool {
	return fileInf[i].Name() < fileInf[j].Name()		
}

func (fileInf FileInfo) Len() int {
	return len(fileInf)
}

func (fileInf FileInfo) Swap(i, j int ) {
	fileInf[i], fileInf[j] = fileInf[j], fileInf[i]
}

func (dir *FileInfo) filter() {
	var newDir FileInfo
	for _, item := range *dir {
		if item.IsDir() {
			newDir = append(newDir, item)
		}
	}
	*dir = newDir
}

func SortDir(dir []fs.FileInfo) FileInfo {
	fileInf := FileInfo(dir)
	sort.Stable(fileInf)
	return fileInf
}

var printFilesFlag bool

func OpenDir(path string) (FileInfo, error) {
	dir, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer dir.Close()

	dirContents, err := dir.Readdir(0)
	if err != nil {
		return nil, err
	}

	resDir := SortDir(dirContents)
	if !printFilesFlag {
		resDir.filter()
	}

	return resDir, err
}

func dirTree(out io.Writer, path string, printFiles bool) error {
	printFilesFlag = printFiles
	rootDir, err := OpenDir(path)
	if err != nil {
		return err
	}

	length := len(rootDir)
	result := ""
	ident := "\n│\t" 
	for i := range rootDir {
		if i == length - 1 {
			ident = "\n\t"
		}
		result += scanDir(path, rootDir, i, ident) + "\n"
	}
	fmt.Fprint(out, result)

	return err
}

func scanDir(path string, dir FileInfo, i int, indent string) string {
	length := len(dir)
	var firstChar string
	if i == length - 1 {
		firstChar = "└"
	} else {
		firstChar = "├" 
	}
	
	branch := ""
	if dir[i].IsDir() {
		pathDirI := fmt.Sprintf("%s%s%s", path, string(os.PathSeparator), dir[i].Name())
		dirContents, _ := OpenDir(pathDirI)
		nextIdent := indent + "│\t"
		for j := range dirContents {
			if j == len(dirContents) - 1{
				nextIdent = indent + "\t"
			} 
			branch += indent + scanDir(pathDirI, dirContents, j, nextIdent)
		}
	} else {
		size := dir[i].Size()
		if size == 0 {
			branch = " (empty)"
		} else {
			branch = fmt.Sprintf(" (%db)", dir[i].Size())
		}
	}
	
	return fmt.Sprintf("%s───%s", firstChar, dir[i].Name()) + branch 

}

func main() {
	out := os.Stdout
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	err := dirTree(out, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
	time.After(time.Second * 2)
}


