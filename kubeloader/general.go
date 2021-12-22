package kubeloader

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
)

var contextName string

var LoaderCache Cache

// inject for some UTs
var runGetStdoutReplace func() (*bytes.Buffer, error)

func LoadFromFile(filename string) (interface{}, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	// parse to json
	var data interface{}
	err = json.Unmarshal([]byte(content), &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func runGetStdout(command string, args ...string) (*bytes.Buffer, error) {
	if runGetStdoutReplace != nil {
		return runGetStdoutReplace()
	}

	cmd := exec.Command(command, args...)
	fmt.Printf("Running command: %v\n", cmd)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	return &out, err
}

// objectType should be the kubernetes thing, eg pod, deploy, secret etc
func LoadFromKubectl(objectType string) (interface{}, error) {

	if LoaderCache != nil {
		result, err := LoaderCache.Load(fmt.Sprintf("%vget_%v_-A.json", contextName, objectType))
		if err == nil && result != nil {
			return result, nil
		}
	}

	out, err := runGetStdout("kubectl", "get", objectType, "-A", "-o=json")
	if err != nil {
		return nil, err
	}
	// parse to interface that can be used by jsonata
	var data interface{}
	err = json.Unmarshal(out.Bytes(), &data)
	if err != nil {
		return nil, err
	}

	if LoaderCache != nil {
		LoaderCache.Store(fmt.Sprintf("%vget_%v_-A.json", contextName, objectType), data) // ignore the error
	}
	return data, nil
}

type Cache interface {
	// returns nil if not in cache, key is the command run
	Load(key string) (interface{}, error)
	// stores result in cache
	Store(key string, data interface{}) error
}

type SimpleFileCache struct {
	CachePath string
}

func (sfc *SimpleFileCache) Load(key string) (interface{}, error) {
	// if the file doesnt exist, exit

	filename := key
	if sfc.CachePath != "" {
		filename = fmt.Sprintf("%v/%v", sfc.CachePath, key)
	}
	_, err := os.Stat(filename)
	if err != nil {
		return nil, nil
	}
	data, err := LoadFromFile(filename)
	if err != nil {
		return nil, err
	}
	return data, nil

}
func (sfc *SimpleFileCache) Store(key string, data interface{}) error {
	filename := key
	if sfc.CachePath != "" {
		filename = fmt.Sprintf("%v/%v", sfc.CachePath, key)
	}
	_, err := os.Stat(filename)
	if err == nil {
		err := os.Remove(filename)
		if err != nil {
			return err
		}
	}

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	_, err = f.Write(dataBytes)
	return err
}
