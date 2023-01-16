package application

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/hkontrol/hkontroller/log"
)

type AccMetadata struct {
	Device    string              `json:"device_id"`
	Accessory uint64              `json:"accessory_id"`
	Data      map[string][]string `json:"metadata"`
}

// AccessoryMetadataStore should store accessories metadata - tags (like room, etc)
type AccessoryMetadataStore struct {
	path string
}

func NewAccessoryMetadataStore(dir string) *AccessoryMetadataStore {
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		log.Info.Panic(err)
	}

	return &AccessoryMetadataStore{
		path: dir,
	}
}

func (c *AccessoryMetadataStore) getPathForAcc(deviceId string, aid uint64) string {
	dd := strings.Replace(deviceId, ":", "", -1)
	filename := fmt.Sprintf("%s_%d.meta", dd, aid)
	return path.Join(c.path, filename)
}

func (c *AccessoryMetadataStore) Save(deviceId string, aid uint64, metadata map[string][]string) error {

	// load and merge
	loaded, err := c.Load(deviceId, aid)
	if err != nil {
		loaded = make(map[string][]string)
	}
	for k, v := range metadata {
		loaded[k] = v
	}

	data := AccMetadata{
		Device:    deviceId,
		Accessory: aid,
		Data:      loaded,
	}
	b, err := json.Marshal(&data)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(c.getPathForAcc(deviceId, aid), os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}

	defer file.Close()

	_, err = file.Write(b)

	return err
}

func (c *AccessoryMetadataStore) Load(deviceId string, aid uint64) (map[string][]string, error) {
	file, err := os.OpenFile(c.getPathForAcc(deviceId, aid), os.O_RDONLY, 0666)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	all, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var data AccMetadata
	err = json.Unmarshal(all, &data)
	if err != nil {
		return nil, err
	}

	return data.Data, nil
}

func (c *AccessoryMetadataStore) GetAll() []AccMetadata {
	files, err := os.ReadDir(c.path)
	if err != nil {
		return nil
	}
	var res []AccMetadata
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		file, err := os.OpenFile(path.Join(c.path, f.Name()), os.O_RDONLY, 0666)
		if err != nil {
			return nil
		}

		defer file.Close()

		all, err := io.ReadAll(file)
		if err != nil {
			return nil
		}

		var data AccMetadata
		err = json.Unmarshal(all, &data)
		if err != nil {
			return nil
		}
		res = append(res, data)
	}
	return res
}

func (c *AccessoryMetadataStore) Remove(deviceId string, aid uint64) error {
	return os.Remove(c.getPathForAcc(deviceId, aid))
}
