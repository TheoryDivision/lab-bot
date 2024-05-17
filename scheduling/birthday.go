package scheduling

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/go-co-op/gocron"
	crondesc "github.com/lnquy/cron"
	log "github.com/sirupsen/logrus"

	"github.com/vishhvaan/lab-bot/db"
	"github.com/vishhvaan/lab-bot/slack"
)

type BirthdaySchedule struct {
	BirthdayMessageChannel string
	scheduler              *gocron.Scheduler
	CronExp                string
	dbPath                 []string
	Logger                 *log.Entry
	sched                  map[string]*Schedule
}

func (bs *BirthdaySchedule) Init(dbPath []string, logger *log.Entry) {
	bs.sched = make(map[string]*Schedule)
	bs.dbPath = dbPath

	bs.scheduler = gocron.NewScheduler(time.Now().Local().Location())
	bs.scheduler.Cron(bs.CronExp).Do(func() {
		bs.Logger.Info("running daily birthday congratulate job")
		bs.congratulate(bs.BirthdayMessageChannel)
	})

	exprDesc, err := crondesc.NewDescriptor()
	if err != nil {
		bs.Logger.Error("cannot create cron descriptor")
	}
	scheduledText, err := exprDesc.ToDescription(bs.CronExp, crondesc.Locale_en)
	if err != nil {
		bs.Logger.Error("cannot convert cron exp to description")
	}

	bs.scheduler.StartAsync()
	slack.Message("Scheduling daily birthday messages " + scheduledText)
	bs.Logger.Info("daily birthday messages " + scheduledText)
}

func (bs *BirthdaySchedule) congratulate(channel string) {
	upcomingBirthdays, err := bs.readUpcomingBirthdays(false)
	if err != nil {
		go bs.Logger.WithError(err).Warn("cannot run daily birthday checks")
		slack.Message("Cannot run daily birthday checks.")
		return
	}

	var todayBDs []string
	for u := range upcomingBirthdays["todayBDs"] {
		todayBDs = append(todayBDs, "<@"+u+">")
	}

	if len(todayBDs) > 0 {
		birthdayMessage := "Happy Birthday " + strings.Join(todayBDs, ", ") + "! :tada:"
		bs.Logger.Info("birthdays found for today")
		slack.SendMessage(channel, birthdayMessage)
	}
	bs.Logger.Info("no birthdays found for today")
}

func (bs *BirthdaySchedule) UpcomingBirthdays(c slack.CommandInfo) {
	var force bool
	if len(c.Fields) > 2 {
		if c.Fields[2] == "force" {
			force = true
		} else {
			slack.SendMessage(c.Channel, "only the force flag is supported")
			return
		}
	}

	upcomingBirthdays, err := bs.readUpcomingBirthdays(force)
	if err != nil {
		bs.errorMsg(c, err, "cannot read upcoming birthdays")
		return
	}

	message, err := bs.formatUpcomingBirthdays(upcomingBirthdays)

	if err != nil {
		bs.errorMsg(c, err, "cannot format upcoming birthdays")
	}
	slack.PostMessage(c.Channel, message)

}

func (bs *BirthdaySchedule) readUpcomingBirthdays(force bool) (upcomingBirthdays map[string]map[string]time.Time, err error) {
	lU, err := db.ReadValue(append(bs.dbPath, "upcoming"), "lastUpdated")
	if err != nil {
		return nil, err
	}

	if lU == nil {
		upcomingBirthdays, err = bs.generateUpcomingBirthdays()
		if err != nil {
			return nil, err
		}
		return upcomingBirthdays, nil
	}

	var lastUpdated time.Time
	err = lastUpdated.UnmarshalJSON(lU)
	if err != nil {
		return nil, err
	}

	if lastUpdated.Truncate(24*time.Hour) == time.Now().Truncate(24*time.Hour) && !force {
		uB, err := db.ReadValue(append(bs.dbPath, "upcoming"), "birthdays")
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(uB, &upcomingBirthdays)
		if err != nil {
			return nil, err
		}
		return upcomingBirthdays, nil
	}

	upcomingBirthdays, err = bs.generateUpcomingBirthdays()
	if err != nil {
		return nil, err
	}
	return upcomingBirthdays, nil
}

func (bs *BirthdaySchedule) generateUpcomingBirthdays() (upcomingBirthdays map[string]map[string]time.Time, err error) {
	upcomingBirthdays = make(map[string]map[string]time.Time)
	upcomingBirthdays["todayBDs"] = make(map[string]time.Time)
	upcomingBirthdays["nextDayBDs"] = make(map[string]time.Time)
	upcomingBirthdays["nextWeekBDs"] = make(map[string]time.Time)
	upcomingBirthdays["nextMonthBDs"] = make(map[string]time.Time)

	now := time.Now()
	today := now.Truncate(24 * time.Hour)
	tomorrow := today.AddDate(0, 0, 1)
	nextWeek := today.AddDate(0, 0, 7)
	nextMonth := today.AddDate(0, 1, 0)

	updateUpcoming := func(userKey []byte, birthdayValue []byte) error {
		var birthday time.Time
		err := birthday.UnmarshalJSON(birthdayValue)
		if err != nil {
			return err
		}

		birthday = time.Date(time.Now().Year(), birthday.Month(), birthday.Day(), birthday.Hour(), birthday.Minute(), birthday.Second(), birthday.Nanosecond(), birthday.Location())

		if birthday == today {
			upcomingBirthdays["todayBDs"][string(userKey)] = birthday
		} else if birthday == tomorrow {
			upcomingBirthdays["nextDayBDs"][string(userKey)] = birthday
		} else if birthday.Before(nextWeek) {
			upcomingBirthdays["nextWeekBDs"][string(userKey)] = birthday
		} else if birthday.Before(nextMonth) {
			upcomingBirthdays["nextMonthBDs"][string(userKey)] = birthday
		}
		return nil
	}

	err = db.RunCallbackOnEachKey(append(bs.dbPath, "records"), updateUpcoming)
	if err != nil {
		return upcomingBirthdays, err
	}
	n, err := now.MarshalJSON()
	if err != nil {
		return upcomingBirthdays, err
	}
	b, err := json.Marshal(upcomingBirthdays)
	if err != nil {
		return upcomingBirthdays, err
	}
	err = db.AddValue(append(bs.dbPath, "upcoming"), "birthdays", b)
	if err != nil {
		return upcomingBirthdays, err
	}
	err = db.AddValue(append(bs.dbPath, "upcoming"), "lastUpdated", n)
	if err != nil {
		return upcomingBirthdays, err
	}

	return upcomingBirthdays, nil
}

func (bs *BirthdaySchedule) formatUpcomingBirthdays(upcomingBirthdays map[string]map[string]time.Time) (message string, err error) {
	var m strings.Builder

	genUsers := func(users map[string]time.Time) (s string) {
		if len(users) == 0 {
			return "none\n"
		} else {
			var items []string
			for user, birthday := range users {
				items = append(items, slack.GetUserName(user)+" ["+birthday.UTC().Format("Jan 02")+"]")
			}
			return strings.Join(items, ", ") + "\n"
		}
	}

	m.WriteString("Upcoming Birthdays:\n")
	m.WriteString("Today: ")
	m.WriteString(genUsers(upcomingBirthdays["todayBDs"]))
	m.WriteString("Tomorrow: ")
	m.WriteString(genUsers(upcomingBirthdays["nextDayBDs"]))
	m.WriteString("Next 7 Days: ")
	m.WriteString(genUsers(upcomingBirthdays["nextWeekBDs"]))
	m.WriteString("Next 30 Days: ")
	m.WriteString(genUsers(upcomingBirthdays["nextMonthBDs"]))

	return m.String(), nil
}

func (bs *BirthdaySchedule) errorMsg(c slack.CommandInfo, err error, message string) {
	go bs.Logger.WithField("fields", c.Fields).WithError(err).Warn(message)
	slack.PostMessage(c.Channel, message)
}
