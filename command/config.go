package command

import (
    "io/ioutil"
    "log"
    "os"
    "strings"
    
    "github.com/jarlex/gommander/plan"
    "github.com/jarlex/gommander/request"
    "github.com/jarlex/gommander/step"
    "github.com/jarlex/gommander/task"
)

type Config struct {
    Plan     *plan.Plan
    Steps    map[string]*step.Step
    Tasks    map[string]*task.Task
    Requests map[string]*request.Request
}

func Read(planFolder string, planFilename ...string) *Config {
    conf := &Config{}
    conf.Requests = make(map[string]*request.Request)
    conf.Steps = make(map[string]*step.Step)
    conf.Tasks = make(map[string]*task.Task)
    
    // Read all Requests
    dir := planFolder + "/requests/"
    files, err := ioutil.ReadDir(dir)
    if err != nil {
        log.Fatal(err)
    }
    
    for _, f := range files {
        aRes, err := request.Read(dir + f.Name())
        if err != nil {
            log.Fatal(err)
        }
        conf.Requests[aRes.Name] = aRes
    }
    
    // Read all Tasks
    dir = planFolder + "/tasks/"
    files, err = ioutil.ReadDir(dir)
    if err != nil {
        log.Fatal(err)
    }
    
    for _, f := range files {
        aTask, err := task.Read(dir+f.Name(), conf.Requests)
        if err != nil {
            log.Fatal(err)
        }
        conf.Tasks[aTask.Name] = aTask
    }
    
    // Read all Steps
    dir = planFolder + "/steps/"
    files, err = ioutil.ReadDir(dir)
    if err != nil {
        log.Fatal(err)
    }
    
    for _, f := range files {
        aStep, err := step.Read(dir+f.Name(), conf.Tasks)
        if err != nil {
            log.Fatal(err)
        }
        conf.Steps[aStep.Name] = aStep
    }
    
    // Read Plan
    planFile := planFolder
    if len(planFilename) > 0 {
        planFile = strings.Join([]string{planFile, planFilename[0]}, string(os.PathSeparator))
    } else {
        planFile = strings.Join([]string{planFile, "plan.json"}, string(os.PathSeparator))
    }
    conf.Plan, err = plan.Read(planFile, conf.Steps)
    if err != nil {
        log.Fatal(err)
    }
    
    return conf
}
