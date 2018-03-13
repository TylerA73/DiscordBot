package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

// Variables used for command line parameters
var (
	Token string
)

type joke struct {
	Joke string `json:"joke"`
}

func init() {

	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.Parse()
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	//s.StateEnabled = true

	userID := m.Author.ID
	username := m.Author.Mention()
	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if userID == s.State.User.ID {
		return
	} else {

	}

	switch strings.ToLower(m.Content) {
	case "joke":
		dadJoke(s, m)
		break
	case "roll":
		rollDie(s, m, username)
		break
	case "two teams":
		teamGenerator(s, m)
	}
}

func dadJoke(s *discordgo.Session, m *discordgo.MessageCreate) {
	client := http.Client{}

	req, err := http.NewRequest(http.MethodGet, "https://icanhazdadjoke.com", nil)
	if err != nil {
		return
	}
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	theJoke := joke{}

	jsonErr := json.Unmarshal(body, &theJoke)
	if jsonErr != nil {
		return
	}

	s.ChannelMessageSend(m.ChannelID, theJoke.Joke)
}

func rollDie(s *discordgo.Session, m *discordgo.MessageCreate, username string) {
	max := 6
	min := 1
	rand.Seed(time.Now().Unix())
	number := rand.Intn(max-min) + min
	s.ChannelMessageSend(m.ChannelID, username+" rolled a "+strconv.Itoa(number))
}

func teamGenerator(s *discordgo.Session, m *discordgo.MessageCreate) {
	channel, err := s.State.Channel(m.ChannelID)
	if err != nil {
		return
	}

	guild, err := s.Guild(channel.GuildID)
	if err != nil {
		return
	}

	members := guild.Members
	undivided := make([]string, 0)
	list := make([]int, 0)

	for i := 0; i < len(members); i++ {
		if !members[i].User.Bot {
			undivided = append(undivided, members[i].User.Username)
			list = append(list, i)
		}
	}

	used := make([]int, len(list))
	team1 := make([]string, 0)
	team2 := make([]string, 0)
	rand.Seed(time.Now().Unix())

	for i := 0; i < len(undivided); i++ {
		number := rand.Intn(len(list))
		for checkValid(number, used) {
			number = rand.Intn(len(list))
		}
		if i%2 == 0 {
			team2 = append(team2, undivided[i])
		} else {
			team1 = append(team1, undivided[i])
		}

		used[i] = number

	}

	s.ChannelMessageSend(m.ChannelID, "TEAM 1")
	for i := 0; i < len(team1); i++ {
		s.ChannelMessageSend(m.ChannelID, team1[i])
	}

	s.ChannelMessageSend(m.ChannelID, "TEAM 2")
	for i := 0; i < len(team2); i++ {
		s.ChannelMessageSend(m.ChannelID, team2[i])
	}

}

func checkValid(val int, numbers []int) bool {
	for i := 0; i < len(numbers); i++ {
		if numbers[i] == val {
			return false
		}
	}

	return true
}

func main() {

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(messageCreate)

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}
