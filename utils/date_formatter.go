package utils

import (
	"discord-bot/constants"
	"time"
)

// FormatDate 단일 날짜를 포맷팅합니다
func FormatDate(date time.Time) string {
	return date.Format(constants.DateFormat)
}

// FormatDateTime 날짜와 시간을 포맷팅합니다
func FormatDateTime(dateTime time.Time) string {
	return dateTime.Format(constants.DateTimeFormat)
}
