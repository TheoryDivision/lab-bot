package scheduling

import (
	"math/rand"

	log "github.com/sirupsen/logrus"

	"github.com/go-co-op/gocron"
	"github.com/vishhvaan/lab-bot/logging"
	"github.com/vishhvaan/lab-bot/slack"
)

const idLength = 5
const idLetters = "abcdefghijklmnopqrstuvwxyz0123456789"

type scheduleRecord struct {
	id      string
	name    string
	cronExp string
	command slack.CommandInfo
}

type Schedule struct {
	scheduleRecord
	command   slack.CommandInfo
	scheduler *gocron.Scheduler
	logger    *log.Entry
	schedule
}

var schedChan chan *Schedule

// // type SchedJobs struct {
// // 	name      string
// // 	status    string
// // 	desc      string
// // 	frequency string
// // 	sched     *gocron.Scheduler
// // 	logger    *log.Entry
// // }

// // type SchedBirthdays struct {
// // 	enabled bool

// // }

type schedule interface {
	init()
	enable()
	disable()
	commandProcessor(c slack.CommandInfo)
}

type ScheduleTracker struct {
	schedules map[string]*Schedule
	messenger chan slack.MessageInfo
	logger    *log.Entry
}

func CreateScheduleTracker() (st *ScheduleTracker) {
	schedules := make(map[string]*Schedule)

	schedLogger := logging.CreateNewLogger("scheduling", "scheduling")

	return &ScheduleTracker{
		schedules: schedules,
		logger:    schedLogger,
	}
}

// func (jh *SchedHandler) InitScheds() {
// 	for job := range jh.jobs {
// 		jh.jobs[job].init()
// 		switch j := jh.jobs[job].(type) {
// 		case *controllerJob:
// 			j.customInit()
// 		}
// 	}
// }

func (st *ScheduleTracker) Reciever() {
	for sched := range schedChan {
		if sched.scheduler.IsRunning() {
			st.schedules[sched.id] = sched
		} else {
			delete(st.schedules, sched.id)
		}
	}
}

func GenerateID() string {
	b := make([]byte, idLength)
	for i := range b {
		b[i] = idLetters[rand.Intn(len(idLetters))]
	}
	return string(b)
}