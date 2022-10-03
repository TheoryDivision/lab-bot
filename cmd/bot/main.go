package main

import (
	"flag"
	"fmt"
	"math/rand"
	"path"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/vishhvaan/lab-bot/config"
	"github.com/vishhvaan/lab-bot/db"
	"github.com/vishhvaan/lab-bot/files"
	"github.com/vishhvaan/lab-bot/jobs"
	"github.com/vishhvaan/lab-bot/logging"
	"github.com/vishhvaan/lab-bot/scheduling"
	"github.com/vishhvaan/lab-bot/slack"
)

var (
	membersFile string
	secretsFile string
	botName     string
	botChannel  string
)

func init() {
	rand.Seed(time.Now().UnixNano())
	logging.Setup()
	exePath := logging.FindExeDir()
	flag.StringVar(&membersFile, "members", path.Join(exePath, "members.yml"), "Location of the members file")
	flag.StringVar(&secretsFile, "secrets", path.Join(exePath, "secrets.yml"), "Location of the secrets file")
	flag.StringVar(&botChannel, "channel", "lab-bot-channel", "Name of the bot channel")
}

func main() {
	fmt.Println("::: Lab Bot :::")
	log.Info("Program Starting...")
	flag.Parse()

	log.Info("Checking config files.")
	files.CheckFile(membersFile)
	files.CheckFile(secretsFile)

	log.Info("Loading config files.")
	members := config.ParseMembers(membersFile)
	secrets := config.ParseSecrets(secretsFile)
	slack.CheckSecrets(secrets)

	db.Open()

	slackClient := slack.CreateClient(secrets, members, botChannel)
	go slackClient.MessageProcessor()
	go slackClient.EventProcessor()
	go slackClient.RunSocketMode()

	scheduleTracker := scheduling.CreateScheduleTracker()
	go scheduleTracker.Reciever()

	jobHandler := jobs.CreateHandler()
	jobHandler.InitJobs()
	jobHandler.CommandReceiver()
}
