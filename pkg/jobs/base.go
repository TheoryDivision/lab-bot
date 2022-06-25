package jobs

import (
	// "github.com/go-co-op/gocron"
	log "github.com/sirupsen/logrus"

	"github.com/vishhvaan/lab-bot/pkg/logging"
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
	name   string
	status bool
	desc   string
	logger *log.Entry
	job
}

type job interface {
	init()
	enable()
	disable()
}

type controllerJob struct {
	labJob
	powerStatus bool
	controller
	logger *log.Entry
}

type controller interface {
	init()
	turnOn()
	turnOff()
}

type JobHandler struct {
	jobs   []*job
	logger *log.Entry
}

func CreateHandler() (jh *JobHandler) {
	var jobs []*job

	jobLogger := logging.CreateNewLogger("jobhandler", "jobhandler")
	controllerLogger := jobLogger.WithField("jobtype", "controller")

	cC := &controllerJob{
		labJob: labJob{
			name:   "Coffee Controller",
			status: false,
			desc:   "Power control for the espresso machine in the lab",
			logger: controllerLogger,
		},
		powerStatus: false,
		logger:      controllerLogger.WithField("job", "coffeeController"),
	}

	jobs = append(jobs, cC)

	return &JobHandler{
		jobs:   jobs,
		logger: jobLogger,
	}
}

func (lj *labJob) init() {
	lj.status = true
}

func (cj *controllerJob) init() {
	cj.labJob.init()
}
