package gestalt

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/fatih/color"
)

type Logger interface {
	Log() logrus.FieldLogger
	CloneFor(string) Logger
	Clone() Logger

	Start()
	Message(string, ...interface{})
	Dump(string)
	Stop(error)
}

type logger struct {
	path   string
	log    logrus.FieldLogger
	out    io.Writer
	logOut io.Writer
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

func (l *logger) Dump(msg string) {
	scanner := bufio.NewScanner(bytes.NewBuffer([]byte(msg)))
	for scanner.Scan() {
		color.New(color.FgWhite, color.Bold).Fprintf(l.logOut, "%v: ", l.path)
		l.logOut.Write(scanner.Bytes())
		l.logOut.Write([]byte("\n"))
	}
}

func (l *logger) Stop(err error) {
	if err == nil {
		fmt.Fprintf(l.out, "%v: [complete]\n", l.path)
	} else {
		fmt.Fprintf(l.out, "%v: [error: %v]\n", l.path, err)
	}
}

func (l *logger) CloneFor(path string) Logger {
	return &logger{path, l.log.WithField("path", path), l.out, l.logOut}
}

func (l *logger) Clone() Logger {
	return &logger{l.path, l.log, l.out, l.logOut}
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

func (lb *logBuilder) WithLevel(level string) *logBuilder {
	lb.log.Level, _ = logrus.ParseLevel(level)
	return lb
}

func (lb *logBuilder) Logger() Logger {
	return &logger{"", lb.log, os.Stdout, lb.log.Out}
}
