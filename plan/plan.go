package plan

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "log"
    "os"
    "time"
    
    "github.com/jarlex/gommander/step"
    "github.com/jarlex/transporter"
)

type Plan struct {
    Type         string   `json:"type"`
    Name         string   `json:"name"`
    AuthType     string   `json:"authType"`
    AuthUser     string   `json:"authUser"`
    AuthPass     string   `json:"authPass"`
    AuthEndpoint string   `json:"authEndpoint"`
    URL          string   `json:"url"`
    Path         string   `json:"path"`
    StepsNames   []string `json:"steps"`
    Steps        []*step.Step
}

func Read(filePath string, steps map[string]*step.Step) (*Plan, error) {
    logger := log.New(os.Stdout, "", 0)
    raw, err := ioutil.ReadFile(filePath)
    
    if err != nil {
        logger.Fatalf("Error reading the file in %s: %s", filePath, err.Error())
        return nil, err
    }
    
    var p Plan
    json.Unmarshal(raw, &p)
    for _, step := range p.StepsNames {
        if steps[step] == nil {
            logger.Fatalf("Error reading the step %s", step)
            return nil, err
        }
        p.Steps = append(p.Steps, steps[step])
    }
    return &p, nil
}

func (p *Plan) Execute() {
    logger := log.New(os.Stdout, "", 0)
    t := transporter.New()
    switch p.AuthType {
    case "basic":
        t.SetBasicAuth(p.AuthUser, p.AuthPass)
    default:
        break
    }
    t.Base(p.URL)
    t.Path(p.Path)
    now := time.Now()
    for _, s := range p.Steps {
        s.Execute(t, p.URL)
    }
    elapsed := time.Since(now)
    logger.Println(fmt.Sprintf("Full Plan: %d", elapsed))
}
