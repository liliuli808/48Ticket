package model

type TicketType struct {
	Brand     string `yaml:"brand"`
	SeatType  string `yaml:"seatType"`
	TicketID  string `yaml:"ticketId"`
	Cookie    string `yaml:"cookie"`
	StartTime string `yaml:"startTime"`
	Num       string `yaml:"num"`
	LogFile   string `yaml:"logFile"`
	Username  string `yaml:"username"`
	Password  string `yaml:"password"`
}
