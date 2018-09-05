package slacky

import (
    "fmt"
    "log"
    "os"
    "io/ioutil"
    "encoding/json"
)

type Log struct {
    Enabled bool `json:"enabled"` // Whether to write a log
    Directory string `json:"directory"` // default = "[Working Directory]/logs/"
    File string `json:"file"` // default = "tracking.log"
}

type Config struct {
    Log `json:"log"`
}

var (
    config Config
)

func init() {
    file, err := ioutil.ReadFile("./logging/config.json")

    if err != nil {
        log.Fatal("File doesn't exist")
    }

    if err := json.Unmarshal(file, &config); err != nil {
        log.Fatal("Cannot parse config.json: " + err.Error())
    }

    if (config.Log.Directory == "default") {
        dir, err := os.Getwd()
        if err != nil {
            fmt.Println("[Fatal Error] " + err.Error())
            log.Fatal(err)
        }

        config.Log.Directory = dir + "/logging/logs/"
    }
}

func WriteToLog(message string, logging Log) {
    if (logging.Enabled == true) {
        configDirectory := logging.Directory
        if (configDirectory == "default") {
            dir, err := os.Getwd()
            if err != nil {
                fmt.Println("[Fatal Error] " + err.Error())
                log.Fatal(err)
            }

            configDirectory = dir + "/logs/"
        }

        configFile := logging.File
        if (configFile == "default") {
            configFile = "tracking.log"
        }

        f, err := os.OpenFile(configDirectory + configFile, os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
        if err != nil {
            fmt.Println("[Fatal Error] Log file update failed!")
            log.Fatal("Log file update failed!")
        }

        log.SetOutput(f)
        log.Println(">>> " + message)

        f.Close()
    }
}