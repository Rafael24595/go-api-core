package command_helper

import (
	"fmt"
	"os"
	"regexp"
	"strconv"

	"github.com/Rafael24595/go-collections/collection"
)

const SnpshTimestamp = `^(snpsh_)(\d*)(\.csvt)$`

func FindSnapshots(path string) (*collection.Vector[os.DirEntry], error) {
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return nil, err
	}

	raw, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("an error ocurred during snapshot directory %q reading: %s", path, err.Error())
	}

	re := regexp.MustCompile(SnpshTimestamp)

	return collection.VectorFromList(raw).
		Filter(func(d os.DirEntry) bool {
			return !d.IsDir() && len(re.FindStringSubmatch(d.Name())) == 4
		}).
		Sort(func(a, b os.DirEntry) bool {
			ar := re.FindStringSubmatch(a.Name())[2]
			at, _ := strconv.ParseInt(ar, 10, 64)

			br := re.FindStringSubmatch(b.Name())[2]
			bt, _ := strconv.ParseInt(br, 10, 64)

			return at < bt
		}), nil
}
