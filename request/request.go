package request

import (
    "encoding/json"
    "errors"
    "fmt"
    "io/ioutil"
    "log"
    "os"
    "strings"
    "time"
    
    "github.com/jarlex/transporter"
)

type Request struct {
    Name       string                 `json:"name"`
    Method     string                 `json:"method"`
    URL        string                 `json:"url"`
    Path       string                 `json:"path"`
    ParamsURL  []string               `json:"paramsURL"`
    ParamsBody []string               `json:"ParamsBody"`
    Body       map[string]interface{} `json:"body"`
}

func Read(filePath string) (*Request, error) {
    logger := log.New(os.Stdout, "", 0)
    raw, err := ioutil.ReadFile(filePath)
    
    if err != nil {
        logger.Fatalf("Error reading the file in %s: %s", filePath, err.Error())
        return nil, err
    }
    
    var r Request
    json.Unmarshal(raw, &r)
    return &r, nil
}

func (r *Request) Execute(tg *transporter.Transporter, base string, callData map[string]interface{}) (map[string]interface{}, int, time.Duration, error) {
    var respJSON map[string]interface{}
    
    directedTg := tg.New()
    // If not Plan URL the task URL is the final path
    if r.URL != "" {
        directedTg = directedTg.Base(r.URL)
    }
    
    finalPath := r.Path
    // Complete request info
    if len(r.ParamsURL) != 0 {
        for _, param := range r.ParamsURL {
            var callDataFinal string
            switch callData[param].(type) {
            case float64:
                callDataFinal = fmt.Sprintf("%f", callData[param].(float64))
            case string:
                callDataFinal = callData[param].(string)
            }
            finalRemplace := fmt.Sprintf("{{%s}}", param)
            if callDataFinal == "" {
                return nil, -1, -1, errors.New("Necesary param not present")
            }
            finalPath = strings.Replace(r.Path, finalRemplace, callDataFinal, 1)
        }
    }
    
    if len(r.ParamsBody) != 0 {
        for _, param := range r.ParamsBody {
            if callData[param] == "" {
                return nil, -1, -1, errors.New("Necesary param not present")
            }
            r.Body[param] = callData[param]
        }
    }
    
    errorJSON := transporter.JSONError{}
    now := time.Now()
    resp, err := directedTg.Path(finalPath).Method(r.Method).BodyJSON(r.Body).Receive(&respJSON, &errorJSON)
    if err != nil {
        return nil, -1, -1, fmt.Errorf("Architecture Error: %s", err.Error())
    }
    
    elapsed := time.Since(now)
    return respJSON, resp.StatusCode, elapsed, nil
}
