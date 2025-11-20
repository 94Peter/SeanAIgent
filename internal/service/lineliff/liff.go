package lineliff

import "fmt"

func InitLineLiff(data map[string]string) {
	// default
	if len(data) == 0 {
		data = map[string]string{
			"training_data": "2008253569-ERDR2wvq",
			"booking":       "2008253569-vD39MBzW",
			"checkin":       "2008253569-y2o46v1g",
		}
	}
	// init line liff
	lineLiffIdMap = make(map[string]*LineLiff)
	for key, value := range data {
		lineLiffIdMap[key] = &LineLiff{
			LiffId:  value,
			LiffUrl: fmt.Sprintf("https://liff.line.me/%s", value),
		}
	}
}

var lineLiffIdMap map[string]*LineLiff

type LineLiff struct {
	LiffId  string
	LiffUrl string
}

var emptyLineLiff = &LineLiff{}

func getLineLiff(key string) *LineLiff {
	v, ok := lineLiffIdMap[key]
	if !ok {
		return emptyLineLiff
	}
	return v
}

func GetTrainingDataLiffId() string {
	return getLineLiff("training_data").LiffId
}

func GetTrainingDataLiffUrl() string {
	return getLineLiff("training_data").LiffUrl
}

func GetBookingLiffId() string {
	return getLineLiff("booking").LiffId
}

func GetBookingLiffUrl() string {
	return getLineLiff("booking").LiffUrl
}

func GetCheckinLiffId() string {
	return getLineLiff("checkin").LiffId
}

func GetCheckinLiffUrl() string {
	return getLineLiff("checkin").LiffUrl
}
