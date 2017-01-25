package gestalt

import (
	"fmt"
	"io"
	"os"

	"github.com/Sirupsen/logrus"
)

type Logger interface {
	Log() logrus.FieldLogger
	CloneFor(string) Logger

	Start()
	Message(string, ...interface{})
	Stop(error)
}

type logger struct {
	path string
	log  logrus.FieldLogger
	out  io.Writer
}

func (l *logger) Log() logrus.FieldLogger {
	return l.log
}

func (l *logger) Start() {
	fmt.Fprintf(l.out, "%v [start]\n", l.path)
}

func (l *logger) Message(msg string, args ...interface{}) {
	fmt.Fprintf(l.out, "%v: %v\n", l.path, fmt.Sprintf(msg, args...))
}

func (l *logger) Stop(err error) {
	if err == nil {
		fmt.Fprintf(l.out, "%v: [complete]\n", l.path)
	} else {
		fmt.Fprintf(l.out, "%v: [error: %v]\n", l.path, err)
	}
}

func (l *logger) CloneFor(path string) Logger {
	return &logger{path, l.log.WithField("path", path), l.out}
}

type logBuilder struct {
	log *logrus.Logger
}

func newLogBuilder() *logBuilder {
	l := logrus.New()
	l.Level = logrus.PanicLevel
	return &logBuilder{
		log: l,
	}
}

func (lb *logBuilder) WithLogOut(o io.Writer) *logBuilder {
	lb.log.Out = o
	return lb
}

func (lb *logBuilder) WithLevel(level logrus.Level) *logBuilder {
	lb.log.Level = level
	return lb
}

func (lb *logBuilder) Logger() Logger {
	return &logger{"", lb.log, os.Stdout}
}
