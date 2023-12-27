package main

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/sirupsen/logrus"
)

func loadSpec(cPath string) (spec *specs.Spec, err error) {
	cf, err := os.Open(cPath)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	defer cf.Close()

	if err = json.NewDecoder(cf).Decode(&spec); err != nil {
		logrus.Error("err")
		return nil, err
	}
	if spec == nil {
		return nil, errors.New("config error")
	}
	return spec, nil
}
