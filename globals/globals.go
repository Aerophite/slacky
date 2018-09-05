package slacky

import(
	"github.com/ejholmes/slash"
	"net/url"
)

type Channels []Channel
func (channels Channels) FindChannel(Identifier string) (Channel, int, bool) {
    for i := 0; i < len(channels); i++ {
        if (channels[i].ID == Identifier || channels[i].Name == Identifier) {
            return channels[i], i, false;
        }
    }
    return Channel{}, -1, true
}

type Channel struct {
    ID string `json:"id"`
    Name string `json:"name"`
}

type Users []User
func (users Users) FindUser(Identifier string) (User, int, bool) {
    for i := 0; i < len(users); i++ {
        if (users[i].ID == Identifier || users[i].Name == Identifier) {
            return users[i], i, false;
        }
    }
    return User{}, -1, true
}

type User struct {
	ID string `json:"id"`
    Name string `json:"name"`
	ResponseURL *url.URL `json:"responseUrl"`
}

type Teams []Team
func (teams Teams) FindTeam(Identifier string) (Team, int, bool) {
    for i := 0; i < len(teams); i++ {
        if (teams[i].ID == Identifier) {
            return teams[i], i, false;
        }
    }
    return Team{}, -1, true
}

type Team struct {
	ID string `json:"id"`
    Name string `json:"name"`
}

type Message struct {
    Token string
    Team Team
    Command string
    Fields []string
    Channel Channel
    User User
    Responder slash.Responder
}