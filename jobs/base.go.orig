package jobs

import (
	"strings"

	log "github.com/sirupsen/logrus"

<<<<<<< HEAD:pkg/jobs/base.go
	"github.com/vishhvaan/lab-bot/pkg/functions"
	"github.com/vishhvaan/lab-bot/pkg/logging"
	"github.com/vishhvaan/lab-bot/pkg/scheduling"
	"github.com/vishhvaan/lab-bot/pkg/slack"
=======
	"github.com/vishhvaan/lab-bot/functions"
	"github.com/vishhvaan/lab-bot/logging"
	"github.com/vishhvaan/lab-bot/slack"
>>>>>>> b62817bb8d7ddd9a773824125bdfb1fe97b197ec:jobs/base.go
)

/*
Todo:
schedule jobs with github.com/go-co-op/gocron
keep track of status
struct for each job with parameters
jobs interface with shared commands
add ability to keep track of messageIDs
create processing for message emojis
task done with reactions -> ends interaction for that time
else post on the lab-bot-channel with the info
(or start group conv) -> ends interaction for that time

find members with particular roles
apply jobs to those roles

upload new members.yml via messages
check and rewrite file and map
*/

type labJob struct {
	name      string
	keyword   string
	active    bool
	desc      string
	logger    *log.Entry
	messenger chan slack.MessageInfo
	commander chan slack.CommandInfo
	responses map[string]action
	job
}

type action func(c slack.CommandInfo)

type job interface {
	init()
	enable()
	disable()
	commandProcessor(c slack.CommandInfo)
}

type JobHandler struct {
	jobs      map[string]job
	messenger chan slack.MessageInfo
	commander chan slack.CommandInfo
	logger    *log.Entry
}

func CreateHandler(m chan slack.MessageInfo, c chan slack.CommandInfo) (jh *JobHandler) {
	jobs := make(map[string]job)

	jobLogger := logging.CreateNewLogger("jobhandler", "jobhandler")
	controllerLogger := jobLogger.WithField("jobtype", "controller")

	cC := &controllerJob{
		labJob: labJob{
			name:      "Coffee Controller",
			keyword:   "coffee",
			active:    true,
			desc:      "Power control for the espresso machine in the lab",
			logger:    controllerLogger,
			messenger: m,
			commander: c,
		},
		machineName: "coffee machine",
		powerStatus: false,
		customInit:  pinInit,
		customOn:    pinOn,
		customOff:   pinOff,
		logger:      controllerLogger.WithField("job", "coffeeController"),
		scheduling: scheduling.ControllerSchedule{
			Logger: controllerLogger.WithField("job", "coffeeController").WithField("task", "scheduling"),
		},
	}

	jobs[cC.keyword] = cC

	return &JobHandler{
		jobs:      jobs,
		messenger: m,
		commander: c,
		logger:    jobLogger,
	}
}

func (jh *JobHandler) InitJobs() {
	for job := range jh.jobs {
		jh.jobs[job].init()
		switch j := jh.jobs[job].(type) {
		case *controllerJob:
			j.customInit()
		}
	}
}

func (jh *JobHandler) CommandReceiver() {
	for command := range jh.commander {
		k := strings.ToLower(command.Fields[0])
		if functions.Contains(functions.GetKeys(jh.jobs), k) {
			jh.jobs[k].commandProcessor(command)
		} else {
			jh.messenger <- slack.MessageInfo{
				Text:      "I couldn't find a response to your command.",
				ChannelID: command.Channel,
			}
		}
	}
}

func (lj *labJob) init() {
	lj.active = true
	// lj.messenger <- slack.MessageInfo{
	// 	Text: lj.name + " has been loaded",
	// }
}

func (lj *labJob) enable() {
	lj.active = true
	lj.logger.Info("Enabled job " + lj.name)
}

func (lj *labJob) disable() {
	lj.active = false
	lj.logger.Info("Disabled job " + lj.name)
}

func (lj *labJob) commandProcessor(c slack.CommandInfo) {}

func commandCheck(c slack.CommandInfo, length int, m chan slack.MessageInfo, l *log.Entry) bool {
	if len(c.Fields) > length {
		message := "Your command has more parameters than necessary"
		go l.Info(message)
		m <- slack.MessageInfo{
			Text:      message,
			ChannelID: c.Channel,
		}
		return false
	} else {
		return true
	}
}
