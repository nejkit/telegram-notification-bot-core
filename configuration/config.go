package configuration

import "time"

type Configuration struct {
	Security struct {
		AllowedAccountIds []int `yaml:"allowed-account-ids" env:"ALLOWED_ACCOUNT_IDS"` //todo: for allowed talks with bot and receiving pushes
		TrustedAccountIds []int `yaml:"trusted-account-ids" env:"TRUSTED_ACCOUNT_IDS"` // for admin operations
	} `envPrefix:"SECURITY_"`

	ScheduleSettings struct {
		TimeSlotsConfiguration map[int]struct { // configure a schedule
			StartTime time.Duration `yaml:"start-time" env:"START_TIME"`
			EndTime   time.Duration `yaml:"end-time" env:"END_TIME"`
		} `yaml:"time-slots-configuration" envPrefix:"TIMESLOTS_"`

		ScheduleRefreshInterval time.Duration `yaml:"schedule-refresh-interval" env:"REFRESH_INTERVAL"`
		ReminderIntervals       []int         `yaml:"reminder-intervals" env:"REMINDER_INTERVALS"` // in minutes
	} `yaml:"schedule-settings" envPrefix:"SCHEDULE_"`

	TelegramTokenBot string `yaml:"telegram-token-bot"`
}
