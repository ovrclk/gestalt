package log

import "github.com/Sirupsen/logrus"

func New() logrus.FieldLogger {
	return logrus.New()
}
