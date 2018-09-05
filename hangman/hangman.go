package slacky

import (
    "fmt"
    "log"
    "os"
    "io/ioutil"
    "encoding/json"
    "strings"
    "strconv"
    "regexp"
    "sort"

    globals "github.com/aerophite/slacky/globals"
    logging "github.com/aerophite/slacky/logging"
    "github.com/ejholmes/slash"
)

type Messages struct {
    Begin string `json:"begin"`
    AlreadyBegan string `json:"alreadyBegan"`
    RequiresRunning string `json:"requiresRunning"`
    Won string `json:"won"`
    Lose string `json:"lose"`
    Correct string `json:"correct"`
    Wrong string `json:"wrong"`
    Duplicate string `json:"duplicate"`
    StarterGuess string `json:"starterGuess"`
    OnlyStarter string `json:"onlyStarter"`
    End string `json:"end"`
    InvalidCommand string `json:"invalidCommand"`
    NotEnoughArguments string `json:"notEnoughArguments"`
    Stat string `json:"stat"`
    StarterCantGuess string `json:"starterCantGuess"`
}

type Config struct {
    NumberOfGuesses int `json:"numberOfGuesses"`
    DefaultChannel string `json:"defaultChannel"`
    AllowConcurrentGuesses bool `json:"allowConcurrentGuesses"`
    AllowStarterToGuess bool `json:"allowStarterToGuess"`
    Messages `json:"messages"`
    Log logging.Log `json:"log"`
    GamesLog logging.Log `json:"gamesLog"`
    StatsLog logging.Log `json:"statsLog"`
}

type Game struct {
    Sentence string `json:"sentence"`
    CurrentSentence string `json:"currentSentence"`
    GuessesRemaining int `json:"guessesRemaining"`
    Guesses map[string]bool `json:"remaining"`
    Channel globals.Channel `json:"channel"`
    Starter globals.User `json:"starter"`
    Players []globals.User `json:"players"`
    Status string `json:"status"`
}

func (games Games) FindGame(Identifier string) (Game, int, bool) {
    for i := 0; i < len(games.Games); i++ {
        if (games.Games[i].Channel.ID == Identifier || games.Games[i].Channel.Name == Identifier) {
            return games.Games[i], i, false;
        }
    }
    return Game{}, -1, true
}

func (games Games) AddGame(game Game) (int, Game, bool) {
    games.Games = append(games.Games, game)
    SetGames(games)

    return (len(games.Games)-1), game, false
}

func (games Games) RemoveGame(gameIndex int) (bool) {
    games.Games[len(games.Games)-1], games.Games[gameIndex] = games.Games[gameIndex], games.Games[len(games.Games)-1]
    games.Games = games.Games[:len(games.Games)-1]
    SetGames(games)

    return false
}

type Games struct {
    Games []Game `json:"games"`
}

type HangmanChannel struct {
    Channel globals.Channel `json:"channel"`
    Values map[string]int `json:"values"`
}

type Stat struct {
    User globals.User `json:"user"`
    HangmanChannels []HangmanChannel `json:"hangmanChannel"`
}

func (stats Stats) FindStat(channel globals.Channel, user globals.User) (int, int) {
    for i := 0; i < len(stats.Stats); i++ {
        if (stats.Stats[i].User.ID == user.ID) {
            for j := 0; j < len(stats.Stats[i].HangmanChannels); j++ {
                if (stats.Stats[i].HangmanChannels[j].Channel.ID == channel.ID) {
                    return i, j;
                }
            }

            return i, -1
        }
    }
    return -1, -1
}

func (stats Stats) AddStat(stat Stat) (int, Stat) {
    stats.Stats = append(stats.Stats, stat)
    SetStats(stats)

    return (len(stats.Stats)-1), stat
}

func (stats Stats) AddToStat(statIndex int, hangmanChannelIndex int, field string) {
    stats.Stats[statIndex].HangmanChannels[hangmanChannelIndex].Values[field]++
    SetStats(stats)
}

type Stats struct {
    Stats []Stat `json:"stats"`
}

var (
    config Config
    games Games
    stats Stats
)

func init() {
    file, err := ioutil.ReadFile("./hangman/config.json")

    if err != nil {
        log.Fatal("File doesn't exist")
    }

    if err := json.Unmarshal(file, &config); err != nil {
        log.Fatal("Cannot parse hangman config.json: " + err.Error())
    }

    if (config.Log.Directory == "default") {
        dir, err := os.Getwd()
        if err != nil {
            fmt.Println("[Fatal Error] " + err.Error())
            log.Fatal(err)
        }

        config.Log.Directory = dir + "/hangman/logs/"
    }

    if (config.GamesLog.Directory == "default") {
        dir, err := os.Getwd()
        if err != nil {
            fmt.Println("[Fatal Error] " + err.Error())
            log.Fatal(err)
        }

        config.GamesLog.Directory = dir + "/hangman/logs/"
    }

    if (config.GamesLog.File == "default") {
        config.GamesLog.File = "games.json"
    }

    if (config.StatsLog.Directory == "default") {
        dir, err := os.Getwd()
        if err != nil {
            fmt.Println("[Fatal Error] " + err.Error())
            log.Fatal(err)
        }

        config.StatsLog.Directory = dir + "/hangman/logs/"
    }

    if (config.StatsLog.File == "default") {
        config.StatsLog.File = "stats.json"
    }

    GetGames()
    GetStats()
}

func SetGames(changedGames Games) {
    games = changedGames
    b, err := json.Marshal(games)
    if err != nil {
        fmt.Println(err)
        return
    }
    // fmt.Println(string(b))
    ioutil.WriteFile(config.GamesLog.Directory + config.GamesLog.File, b, 0666)
}

func GetGames() {
    file, err := ioutil.ReadFile(config.GamesLog.Directory + config.GamesLog.File)

    if err == nil {
        if err := json.Unmarshal(file, &games); err != nil {
            log.Fatal(fmt.Sprintf("Cannot parse : %s", config.GamesLog.File) + err.Error())
        }
    }
}

func SetStats(changedStats Stats) {
    stats = changedStats
    b, err := json.Marshal(stats)
    if err != nil {
        fmt.Println(err)
        return
    }
    // fmt.Println(string(b))
    ioutil.WriteFile(config.StatsLog.Directory + config.StatsLog.File, b, 0666)
}

func GetStats() {
    file, err := ioutil.ReadFile(config.StatsLog.Directory + config.StatsLog.File)

    if err == nil {
        if err := json.Unmarshal(file, &stats); err != nil {
            log.Fatal(fmt.Sprintf("Cannot parse : %s", config.StatsLog.File) + err.Error())
        }
    }
}

func hasRequiredSize(min int, size int, message globals.Message) (bool) {
    if (size >= min) {
        return true
    }

    message.Responder.Respond(slash.Reply(generateReply(config.Messages.NotEnoughArguments, map[string]string {"<min>": strconv.Itoa(min)})))
    return false
}

func hasRequiredGame(gameIndex int, flipped bool, message globals.Message) (bool) {
    response := config.Messages.RequiresRunning
    if (flipped == false && gameIndex != -1) {
        return true
    } else if (flipped == true) {

        response = config.Messages.AlreadyBegan

        if (gameIndex == -1) {
            return true
        }
    }

    message.Responder.Respond(slash.Reply(response))
    return false
}

func generateHangmanString(sentence string, replace string) (string) {
    sentence = strings.Replace(sentence, " ", "   ", -1)
    re := regexp.MustCompile(replace)
    sentence = strings.TrimSpace(re.ReplaceAllString(sentence, " _"))
    re = regexp.MustCompile("([a-z])")
    sentence = re.ReplaceAllString(sentence, " $1")
    sentence = strings.Replace(sentence, "    ", "   ", -1)

    return strings.TrimSpace(sentence)
}

func generateReply(sentence string, replacements map[string]string) string {
    // name, min, sentence, guess, currentSentence

    for k, v := range replacements {
        sentence = strings.Replace(sentence, k, v, -1)
    }

    return sentence
}

func Hangman(message globals.Message) error {
    _, gameIndex, _ := games.FindGame(message.Channel.ID)
    statIndex, hangmanChannelIndex := stats.FindStat(message.Channel, message.User)

    if (statIndex == -1) {
        stats.AddStat(Stat{
            message.User,
            []HangmanChannel{
                message.Channel,
                map[string]int{"started": 0, "games": 0, "wins": 0, "guesses": 0, "correct": 0}}})
    }

    numOfFields := len(message.Fields)

    if (strings.Replace(message.Text, "_", "", -1) != message.Text) {
        if err := message.Responder.Respond(slash.Reply("Underscores are not allowed")); err != nil {
            return err
        }

        return nil
    }

    switch message.Command {
        case "ping":
            ping(message)
        case "start":
            if (hasRequiredGame(gameIndex, true, message) && hasRequiredSize(1, numOfFields, message)) {
                if err := start(message); err != nil {
                    stats.AddToStat(statIndex, hangmanChannelIndex, "Games")
                }
            }
        case "guess":
            if (hasRequiredGame(gameIndex, false, message) && hasRequiredSize(1, numOfFields, message)) {
                guess(message, gameIndex)
            }
        case "stop":
            if (hasRequiredGame(gameIndex, false, message)) {
                stop(message, gameIndex, false)
            }
        case "stat":
            if (hasRequiredGame(gameIndex, false, message)) {
                status(message, gameIndex, true)
            }
        case "help":
            help(message)
        default:
            if err := message.Responder.Respond(slash.Reply(config.Messages.InvalidCommand)); err != nil {
                return err
            }
            help(message)
    }

    return nil
}

func ping(message globals.Message) error {
    logging.WriteToLog("ping", config.Log);

    if err := message.Responder.Respond(slash.Reply("Pong!")); err != nil {
        return err
    }

    return nil
}

func start(message globals.Message) error { // sentence, number of guesses
    logging.WriteToLog("start", config.Log);

    guessesRemaining := config.NumberOfGuesses
    sentence := strings.Join(strings.Fields(strings.Join(message.Fields, " ")), " ")

    gameIndex, _, _ := games.AddGame(Game{
            sentence,
            generateHangmanString(strings.ToLower(sentence), "[abcdefghijklmnopqrstuvwxyz]"),
            guessesRemaining,
            map[string]bool {"a": false, "b": false, "c": false, "d": false, "e": false, "f": false, "g": false, "h": false, "i": false, "j": false, "k": false, "l": false, "m": false, "n": false, "o": false, "p": false, "q": false, "r": false, "s": false, "t": false, "u": false, "v": false, "w": false, "x": false, "y": false, "z": false},
            message.Channel,
            message.User,
            []globals.User{},
            "in-process"})

    // Notify channel that a game has started
    if err := message.Responder.Respond(slash.Say(generateReply(config.Messages.Begin, map[string]string {"<name>": message.User.Name}))); err != nil {
        return err
    }

    status(message, gameIndex, false)

    return nil
}

func guess(message globals.Message, gameIndex int) error { // guess
    logging.WriteToLog("guess", config.Log);

    if (message.Channel.Name != "directmessage" && config.AllowStarterToGuess == false && message.User.ID == games.Games[gameIndex].Starter.ID) {
        if err := message.Responder.Respond(slash.Reply(config.Messages.StarterCantGuess)); err != nil {
            return err
        }

        return nil
    }

    actualGuess := strings.Join(strings.Fields(strings.Join(message.Fields, " ")), " ")
    actualSentence := ""
    currentSentence := ""
    re := regexp.MustCompile("([^a-zA-Z0-9_ ])")
    lowerGuess := re.ReplaceAllString(strings.ToLower(actualGuess), "")

    if (games.Games[gameIndex].Guesses[lowerGuess] == true) {
        if err := message.Responder.Respond(slash.Reply(generateReply(config.Messages.Duplicate, map[string]string {"<guess>": actualGuess}))); err != nil {
            return err
        }

        status(message, gameIndex, true)
        return nil
    } else {
        replace := ""
        if (len(actualGuess) == 1) {
            games.Games[gameIndex].Guesses[lowerGuess] = true

            replace = "["
            for k, v := range games.Games[gameIndex].Guesses {
                if (v == false && len(k) == 1) {
                    replace = replace + k + strings.ToUpper(k)
                }
            }
            replace = replace + "]"

            currentSentence = generateHangmanString(games.Games[gameIndex].Sentence, replace)
            re2 := regexp.MustCompile("((\\w) )")
            actualSentence = re2.ReplaceAllString(currentSentence, "$2")
            actualSentence = strings.Replace(actualSentence, "  ", " ", -1);
            re = regexp.MustCompile("([^a-zA-Z0-9_])")
            actualSentence = re.ReplaceAllString(strings.ToLower(currentSentence), "")
        } else {
            games.Games[gameIndex].Guesses[strings.ToLower(actualGuess)] = true

            currentSentence = games.Games[gameIndex].CurrentSentence
            actualSentence = lowerGuess
        }

        reply := ""
        if (re.ReplaceAllString(strings.ToLower(games.Games[gameIndex].Sentence), "") == actualSentence) {
            if err := message.Responder.Respond(slash.Say(generateReply(config.Messages.Won, map[string]string {"<name>": message.User.Name, "<sentence>": games.Games[gameIndex].Sentence}))); err != nil {
                return err
            }

            stop(message, gameIndex, true)
            return nil
        } else if (games.Games[gameIndex].CurrentSentence == currentSentence) {
            reply = generateReply(config.Messages.Wrong, map[string]string {"<name>": message.User.Name, "<guess>": actualGuess})

            games.Games[gameIndex].GuessesRemaining = games.Games[gameIndex].GuessesRemaining - 1
            if (games.Games[gameIndex].GuessesRemaining == 0) {
                if err := message.Responder.Respond(slash.Say(generateReply(config.Messages.Lose, map[string]string {"<sentence>": games.Games[gameIndex].Sentence}))); err != nil {
                    return err
                }

                stop(message, gameIndex, true)
                return nil
            }
        } else {
            games.Games[gameIndex].CurrentSentence = currentSentence
            reply = generateReply(config.Messages.Correct, map[string]string {"<name>": message.User.Name, "<guess>": actualGuess})
        }

        if err := message.Responder.Respond(slash.Say(reply)); err != nil {
            return err
        }

        status(message, gameIndex, false)
    }

    SetGames(games)
    return nil

    // If guess is 1 character
        // If in Game.Remaining
            // Remove from Game.Remaining
            // Notify channel that there was a correct guess
            // Show game stats
        // Else
            // Notify channel that there was a wrong guess
            // Show game stats
    // Else
        // Someone is making a full guess
        // If guess equals Game.Sentence
            // Notify the channel that the game has been won
            // Trigger end, overriding need to be the starter
}

func stop(message globals.Message, gameIndex int, override bool) error {
    logging.WriteToLog("stop", config.Log);

    if (override == true) {
        // If overriding, just remove the game
        games.RemoveGame(gameIndex)
    } else if (games.Games[gameIndex].Starter.ID == message.User.ID) {
        // If starter, remove game and notify channel
        games.RemoveGame(gameIndex)

        if err := message.Responder.Respond(slash.Say(generateReply(config.Messages.End, map[string]string {"<name>": message.User.Name}))); err != nil {
            return err
        }
    } else {
        // Notify sender that they cannot end the game since they didn't start it
        if err := message.Responder.Respond(slash.Reply(config.Messages.OnlyStarter)); err != nil {
            return err
        }
    }

    return nil
}

func status(message globals.Message, gameIndex int, reply bool) error {
    logging.WriteToLog("stat", config.Log);

    remaining := ""
    guessedChars := ""
    guessedSentence := ""

    for k, v := range games.Games[gameIndex].Guesses {
        if (v == false) {
            remaining = remaining + "   " + strings.ToUpper(k)
        } else {
            if (len(k) == 1) {
                guessedChars = guessedChars + "   " + strings.ToUpper(k)
            } else {
                guessedSentence = guessedSentence + "   \"" + k + "\""
            }
        }
    }

    remainingSlice := strings.Fields(strings.TrimSpace(remaining))
    sort.Strings(remainingSlice)
    remaining = strings.Join(remainingSlice, "  ")

    guessedSlice := strings.Fields(strings.TrimSpace(guessedChars))
    sort.Strings(guessedSlice)
    guessed := strings.TrimSpace(strings.Join(guessedSlice, "  ") + guessedSentence)

    newMessage := generateReply(config.Messages.Stat, map[string]string {"<currentSentence>": "```" + games.Games[gameIndex].CurrentSentence + "```", "<guessed>": guessed, "<remaining>": remaining, "<guessesRemaining>": strconv.Itoa(games.Games[gameIndex].GuessesRemaining)})

    if (reply) {
        if err := message.Responder.Respond(slash.Reply(newMessage)); err != nil {
            return err
        }
    } else {
        if err := message.Responder.Respond(slash.Say(newMessage)); err != nil {
            return err
        }
    }

    return nil
}

func help(message globals.Message) error {
    reply := "\n*/hangman help* : This screen, just showing you some helpful commands.\n\n*/hangman start (word|sentence)* : Start a game in the current channel. Your word or sentence will not be shown to anyone.\n\n*/hangman stop* : Stops a game. Can only be ran by the person that started the game.\n\n*/hangman guess (character|sentence)* : Make a guess at the word or sentence. If guessing a sentence, it will attempt an exact match (minus punctuation and spaces).\n\n*/hangman stat* : Get the current status of the game.\n"

    if err := message.Responder.Respond(slash.Reply(reply)); err != nil {
        return err
    }

    return nil
}