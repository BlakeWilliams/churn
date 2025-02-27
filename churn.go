package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var partsRegex = regexp.MustCompile(`\s+`)
var renameRegex = regexp.MustCompile(`(.*){(.+)=>(.+)}(.*)`)

type File struct {
	CurrentName    string `json:"name"`
	MostRecentName string `json:"-"`
	TimesModified  int    `json:"updates"`
	Additions      int    `json:"additions"`
	Deletions      int    `json:"deletions"`
}

func main() {
	date := "2023-02-25"

	var out bytes.Buffer
	cmd := exec.Command("git", "log", "--numstat", "--since", date, "--pretty=format:", "--diff-filter=AMRCD")
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
		if len(strings.TrimSpace(line)) == 0 {
			continue
		}

		parts := partsRegex.Split(line, 3)
		if parts[0] == "-" {
			continue
		}

		additions, err := strconv.Atoi(parts[0])
		if err != nil {
			panic(err)
		}
		deletions, err := strconv.Atoi(parts[1])
		if err != nil {
			panic(err)
		}
		filename := strings.TrimSpace(parts[2])

		if renameRegex.MatchString(filename) {
			renameParts := renameRegex.FindStringSubmatch(filename)

			commonPrefix := renameParts[1]
			left := strings.TrimSpace(renameParts[2])
			right := strings.TrimSpace(renameParts[3])
			commonSuffix := renameParts[4]

			right = fmt.Sprintf("%s%s", commonPrefix, right)

			for k, file := range files {
				if right == file.CurrentName || (strings.HasPrefix(right, file.CurrentName) && strings.TrimPrefix(right, file.CurrentName)[0] == '/') && strings.HasSuffix(right, commonSuffix) {
					file.CurrentName = commonPrefix + left + commonSuffix
					delete(files, k)
					files[file.CurrentName] = file
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

	allFiles := []*File{}

	for _, file := range files {
		allFiles = append(allFiles, file)
	}

	// sort by times changed descending
	sort.Slice(allFiles, func(i, j int) bool {
		return allFiles[i].TimesModified > allFiles[j].TimesModified
	})

	res, err := json.MarshalIndent(allFiles, "", "  ")
	if err != nil {
		panic(err)
	}

	fmt.Println(string(res))
}
