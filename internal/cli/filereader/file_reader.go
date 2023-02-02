package filereader

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/goccy/go-json"
	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
)

var (
	ErrExtNotSupported = errors.New("extension not suppoerted")
)

type FileReader struct {
	reader *http.Client
}

func New() (*FileReader, error) {
	c := &http.Client{}
	t := &http.Transport{}
	t.RegisterProtocol("file", http.NewFileTransport(http.Dir("/")))
	c.Transport = t

	return &FileReader{reader: c}, nil
}

// this is taken from internal golang codebase
// if net/url exports this function switch to that one
func urlFromFilePath(path string) (*url.URL, error) {
	if !filepath.IsAbs(path) {
		wd, _ := os.Getwd()
		path = filepath.Clean(filepath.Join(wd, path))
	}

	// If path has a Windows volume name, convert the volume to a host and prefix
	// per https://blogs.msdn.microsoft.com/ie/2006/12/06/file-uris-in-windows/.
	if vol := filepath.VolumeName(path); vol != "" {
		if strings.HasPrefix(vol, `\\`) {
			path = filepath.ToSlash(path[2:])
			i := strings.IndexByte(path, '/')

			if i < 0 {
				// A degenerate case.
				// \\host.example.com (without a share name)
				// becomes
				// file://host.example.com/
				return &url.URL{
					Scheme: "file",
					Host:   path,
					Path:   "/",
				}, nil
			}

			// \\host.example.com\Share\path\to\file
			// becomes
			// file://host.example.com/Share/path/to/file
			return &url.URL{
				Scheme: "file",
				Host:   path[:i],
				Path:   filepath.ToSlash(path[i:]),
			}, nil
		}

		// C:\path\to\file
		// becomes
		// file:///C:/path/to/file
		return &url.URL{
			Scheme: "file",
			Path:   "/" + filepath.ToSlash(path),
		}, nil
	}

	// /path/to/file
	// becomes
	// file:///path/to/file
	return &url.URL{
		Scheme: "file",
		Path:   filepath.ToSlash(path),
	}, nil
}

func (fr *FileReader) ParseFile(raw []byte, data any, ext string) error {
	switch ext {
	case "json":
		if err := json.Unmarshal(raw, data); err != nil {
			return err
		}
	case "yaml":
		if err := yaml.Unmarshal(raw, data); err != nil {
			return err
		}
	case "yml":
		if err := yaml.Unmarshal(raw, data); err != nil {
			return err
		}
	case "toml":
		if err := toml.Unmarshal(raw, data); err != nil {
			return err
		}
	default:
		return ErrExtNotSupported
	}

	return nil
}

func (fr *FileReader) ReadIntoRawBytes(location string) ([]byte, error) {
	// if location is not url convert to proper url with file:// format
	// We use golang http client to get files both in system and from web
	url, err := url.ParseRequestURI(location)
	isValidurl := err == nil && url.Scheme != ""
	if !isValidurl {
		url, err = urlFromFilePath(location)
	}
	// change the location to the url path
	// this also handles conversion from relative to absolute path
	location = url.String()

	if err != nil {
		return nil, err
	}

	resp, err := fr.reader.Get(location)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode > 399 {
		return nil, fmt.Errorf("failed to get the file in %s - status code %d", location, resp.StatusCode)
	}

	return ioutil.ReadAll(resp.Body)
}

// to read a file and parse it
func (fr *FileReader) ReadFile(location string, data any) error {
	if reflect.ValueOf(data).Kind() != reflect.Ptr {
		return errors.New("data must be a pointer")
	}

	raw, err := fr.ReadIntoRawBytes(location)
	if err != nil {
		return err
	}

	ext := filepath.Ext(location)[1:] // .json -> json

	if err := fr.ParseFile(raw, data, ext); err != nil {
		return err
	}
	return nil
}

// this is just for a special case in whic swagger validation requires raw buffer
// for all other usecases use ReadFile to get parsed go structure
// or the GetRaw for getting raw bytes data
func (fr *FileReader) ReadFileReturnRaw(location string, data any) ([]byte, error) {
	if reflect.ValueOf(data).Kind() != reflect.Ptr {
		return nil, errors.New("data must be a pointer")
	}

	raw, err := fr.ReadIntoRawBytes(location)
	if err != nil {
		return nil, err
	}

	ext := filepath.Ext(location)[1:] // .json -> json
	if err := fr.ParseFile(raw, data, ext); err != nil {
		return nil, err
	}

	return raw, nil
}
