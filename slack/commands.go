package slack

var CommandChan = make(chan CommandInfo)

type CommandInfo struct {
	Fields    []string
	Channel   string
	TimeStamp string
	User      string
}

func (sc *slackClient) RunSocketMode() {
	sc.client.Run()
}

func (sc *slackClient) getChannelName(channelID string) (channel string) {
	ch, err := sc.api.GetConversationInfo(channelID, false)
	if err != nil {
		sc.logger.WithField("err", err).Error("Couldn't find conversation info.")
	}
	return ch.Name
}

func (sc *slackClient) getUserName(userID string) (user string) {
	us, err := sc.api.GetUserInfo(userID)
	if err != nil {
		sc.logger.WithField("userID", userID).WithField("err", err).
			Warn("couldn't fetch user info; falling back to mention")
		return "<@" + userID + ">"
	}

	name := us.Profile.DisplayName
	if name == "" {
		name = us.Profile.RealName
	}
	if name == "" {
		name = "<@" + userID + ">"
	}

	return name
}
