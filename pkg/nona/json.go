package nona

import (
	"encoding/json"
	"os"
)

type JSON struct {
	Data map[string]string `json:"data"`
}

type jsonConfig struct {
	Path string `json:"path"`
}

var _ Store = (*JSON)(nil)

func NewJSON(config json.RawMessage) (s Store, err error) {
	var data jsonConfig
	if err := json.Unmarshal(config, &data); err != nil {
		return nil, err
	}
	var j JSON
	f, err := os.Open(data.Path)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err2 := f.Close(); err2 != nil {
			err = err2
		}
	}()
	if err := json.NewDecoder(f).Decode(&j.Data); err != nil {
		return nil, err
	}
	return &j, nil
}

func (j *JSON) Get(key string) (url string, exists bool, err error) {
	url, exists = j.Data[key]
	return url, exists, nil
}
