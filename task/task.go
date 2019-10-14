package task

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"globaldevtools.bbva.com/bitbucket/scm/nbdnt/nbdnt_gommander.git/request"
	"globaldevtools.bbva.com/bitbucket/scm/nbdnt/nbdnt_tongue.git"
)

//Task model
type Task struct {
	Name           string   `json:"name"`
	PreviousData   []string `json:"previusData"`
	NextData       []string `json:"nextData"`
	ExpectedStatus int      `json:"expectedStatus"`
	NameRequest    string   `json:"request"`
	Request        *request.Request
}

/*     Public Methods    */
//-------------------------------------------------------------------------------------------------------------------------//

//Read a task in filePath
func Read(filePath string, requests map[string]*request.Request) (*Task, error) {
	logger := log.New(os.Stdout, "", 0)
	raw, err := ioutil.ReadFile(filePath)

	if err != nil {
		logger.Fatalf("Error reading the file in %s: %s", filePath, err.Error())
		return nil, err
	}

	var t Task
	json.Unmarshal(raw, &t)
	t.Request = requests[t.NameRequest]
	return &t, nil
}

//Execute a task
func (t *Task) Execute(tg *tongue.Tongue, base string, previousData map[string]interface{}) (map[string]interface{}, time.Duration, error) {

	if t.PreviousData != nil {
		for _, field := range t.PreviousData {
			if previousData[field] == "" {
				return nil, -1, fmt.Errorf("%s mandatory and not present in previousData", field)
			}
		}
	}

	resp, status, duration, err := t.Request.Execute(tg, base, previousData)
	if err != nil {
		return nil, -1, err
	}

	if status != t.ExpectedStatus {
		return nil, -1, errors.New("Status not expected")
	}

	nextData := make(map[string]interface{})

	if t.NextData != nil {
		for _, field := range t.NextData {
			if resp[field] == "" {
				return nil, -1, fmt.Errorf("%s mandatory and not present in nextData", field)
			}
			nextData[field] = resp[field]
		}
	}

	return nextData, duration, nil
}