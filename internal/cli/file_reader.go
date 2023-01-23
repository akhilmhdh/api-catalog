package cli

import (
	"errors"
	"io/ioutil"
	"path/filepath"

	"github.com/goccy/go-json"
)

var (
	ErrExtNotSupported = errors.New("extension not suppoerted")
)

type FileReader struct {
	fileName string
	ext      string
	data     map[string]interface{}
}

func NewFileReader(path string) (*FileReader, error) {

	file, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// get the extension of file
	ext := filepath.Ext(path)[1:] // .json -> json
	// set the reader
	fr := &FileReader{fileName: filepath.Base(path), ext: ext}

	switch ext {
	case "json":
		if err = json.Unmarshal(file, &fr.data); err != nil {
			return nil, err
		}
	default:
		return nil, ErrExtNotSupported
	}

	return fr, nil
}
