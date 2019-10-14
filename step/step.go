package step

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "log"
    "os"
    "sync"
    "time"
    
    "globaldevtools.bbva.com/bitbucket/scm/nbdnt/nbdnt_gommander.git/task"
    "globaldevtools.bbva.com/bitbucket/scm/nbdnt/nbdnt_tongue.git"
)

type Step struct {
    Name            string       `json:"name"`            // Step Name
    NumPetitions    int          `json:"numPetitions"`    // Number of petitions
    ConcurrentUsers int          `json:"concurrentUsers"` // Concurrent users
    TasksNames      []string     `json:"tasks"`           // Concurrent users
    Tasks           []*task.Task // Orderer tasks
}

func Read(filePath string, tasks map[string]*task.Task) (*Step, error) {
    logger := log.New(os.Stdout, "", 0)
    raw, err := ioutil.ReadFile(filePath)
    
    if err != nil {
        logger.Fatalf("Error reading the file in %s: %s", filePath, err.Error())
        return nil, err
    }
    
    var s Step
    json.Unmarshal(raw, &s)
    for _, t := range s.TasksNames {
        s.Tasks = append(s.Tasks, tasks[t])
    }
    return &s, nil
}

func (s *Step) Execute(t *tongue.Tongue, base string) {
    reqEachUser := s.NumPetitions / s.ConcurrentUsers
    var wg sync.WaitGroup
    wg.Add(s.ConcurrentUsers)
    for user := 0; user < s.ConcurrentUsers; user++ {
        go func(user int) {
            defer wg.Done()
            for petition := 0; petition < reqEachUser; petition++ {
                var previousData map[string]interface{}
                var totalTime int64
                totalTime = 0
                for _, tsk := range s.Tasks {
                    var duration time.Duration
                    var err error
                    previousData, duration, err = tsk.Execute(t, base, previousData)
                    if err != nil {
                        fmt.Println(fmt.Sprintf("%s|U%d|FAIL|%s|%d|%s", s.Name, user, tsk.Name, petition, err.Error()))
                        break
                    }
                    fmt.Println(fmt.Sprintf("%s|U%d|%d ns|T%s|%d", s.Name, user, duration.Nanoseconds(), tsk.Name, petition))
                    totalTime = totalTime + duration.Nanoseconds()
                }
                fmt.Println(fmt.Sprintf("T|%s|U%d|%d ns|%d", s.Name, user, totalTime, petition))
            }
        }(user)
    }
    wg.Wait()
}
