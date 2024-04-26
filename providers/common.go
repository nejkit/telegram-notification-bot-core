package providers

import (
	"github.com/sirupsen/logrus"
	"io/fs"
	"io/ioutil"
	"os"
)

type CommonProvider struct {
	filename string
}

func newCommonProvider(dataName string) *CommonProvider {

	if _, err := os.Stat("./storage"); os.IsNotExist(err) {
		logrus.Infoln("Creation a new directory for storage")

		err = os.MkdirAll("./storage", os.ModePerm)

		if err != nil {
			panic(err)
		}
	}

	if _, err := os.Stat("./storage/" + dataName + ".json"); os.IsNotExist(err) {
		logrus.Infoln("Creation a new storage")

		_, err = os.Create("./storage/" + dataName + ".json")
		if err != nil {
			panic(err)
		}
	}

	return &CommonProvider{filename: "./storage/" + dataName + ".json"}
}

func (c *CommonProvider) getAllDataFromStorage() ([]byte, error) {

	text, err := ioutil.ReadFile(c.filename)

	if err != nil {
		return nil, err
	}

	return text, nil
}

func (c *CommonProvider) saveAllDataToStorage(data []byte) error {
	return ioutil.WriteFile(c.filename, data, fs.ModeAppend)
}
