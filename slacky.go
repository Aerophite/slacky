package main

import (
	"fmt"
	"net/http"
	"os"
	"log"
	"io/ioutil"
    "encoding/json"
    "strings"

    globals "github.com/aerophite/slacky/globals"
	logging "github.com/aerophite/slacky/logging"
	hangman "github.com/aerophite/slacky/hangman"
	"github.com/ejholmes/slash"
	"golang.org/x/net/context"
)

type Config struct {
    Log logging.Log `json:"log"`
}

var (
    config Config
)

func init() {
    file, err := ioutil.ReadFile("./config.json")

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

        config.Log.Directory = dir + "/logs/"
    }
}

func main() {
	h := slash.HandlerFunc(Handle)
	s := slash.NewServer(h)
	logging.WriteToLog("Slack Thingy Started!", config.Log);
	http.ListenAndServe(":8082", s)
}

func Handle(ctx context.Context, r slash.Responder, command slash.Command) error {

	fields := strings.Fields(strings.TrimSpace(command.Text))
    field, fields := fields[0], fields[1:]

	message := globals.Message{
		command.Token,
		globals.Team{command.TeamID, command.TeamDomain},
		field,
		command.Text,
		fields,
		globals.Channel{command.ChannelID, command.ChannelName},
		globals.User{command.UserID, command.UserName, command.ResponseURL},
		r}

	switch command.Command {
        case "/hangman":
            hangman.Hangman(message)
    }

	return nil
}

func printErrors(h slash.Handler) slash.Handler {
	return slash.HandlerFunc(func(ctx context.Context, r slash.Responder, command slash.Command) error {
		if err := h.ServeCommand(ctx, r, command); err != nil {
			fmt.Printf("error: %v\n", err)
		}
		return nil
	})
}