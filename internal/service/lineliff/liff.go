package lineliff

import "fmt"

var lineLiffIdMap = map[string]string{
	"booking":       "2008253569-vD39MBzW",
	"checkin":       "2008253569-y2o46v1g",
	"training_data": "2008253569-ERDR2wvq",
}
var lineLiffUrlMap = map[string]string{
	"booking":       fmt.Sprintf("https://liff.line.me/%s", lineLiffIdMap["booking"]),
	"checkin":       fmt.Sprintf("https://liff.line.me/%s", lineLiffIdMap["checkin"]),
	"training_data": fmt.Sprintf("https://liff.line.me/%s", lineLiffIdMap["training_data"]),
}

func GetTrainingDataLiffId() string {
	return lineLiffIdMap["training_data"]
}

func GetTrainingDataLiffUrl() string {
	return lineLiffUrlMap["training_data"]
}

func GetBookingLiffId() string {
	return lineLiffIdMap["booking"]
}

func GetBookingLiffUrl() string {
	return lineLiffUrlMap["booking"]
}

func GetCheckinLiffId() string {
	return lineLiffIdMap["checkin"]
}

func GetCheckinLiffUrl() string {
	return lineLiffUrlMap["checkin"]
}
