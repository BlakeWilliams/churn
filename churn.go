package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

var partsRegex = regexp.MustCompile(`\s+`)
var renameRegex = regexp.MustCompile(`{(.+)=>(.+)}`)

type File struct {
	CurrentName    string
	MostRecentName string

	TimesModified int

	Additions int
	Deletions int
}

func main() {
	date := "2025-02-25"

	var out bytes.Buffer
	cmd := exec.Command("git", "log", "--since", date, "--pretty=format:", "--diff-filter=AMRCD")
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		fmt.Println(out.String())
		panic(err)
	}

	files := map[string]*File{}

	reader := bufio.NewReader(&out)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		if len(line) == 0 {
			continue
		}

		parts := partsRegex.Split(line, 3)
		additions, err := strconv.Atoi(parts[0])
		if err != nil {
			panic(err)
		}
		deletions, err := strconv.Atoi(parts[1])
		if err != nil {
			panic(err)
		}
		filename := parts[2]

		if renameRegex.MatchString(filename) {
			renameParts := renameRegex.FindStringSubmatch(filename)
			left := renameParts[1]
			right := renameParts[2]

			for _, file := range files {
				if left == file.CurrentName || (strings.HasPrefix(left, file.CurrentName) && strings.TrimPrefix(left, file.CurrentName)[0] == '/') {
					file.CurrentName = right
				}
			}

			continue
		}

		if files[filename] == nil {
			files[filename] = &File{
				CurrentName:    filename,
				MostRecentName: filename,
			}
		}

		files[filename].Additions += additions
		files[filename].Deletions += deletions
		files[filename].TimesModified++
	}
}
