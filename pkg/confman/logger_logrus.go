package confman

import "github.com/sirupsen/logrus"

type LogrusWrapper struct {
	*logrus.Logger
}

func (lw LogrusWrapper) WithField(key string, value interface{}) Logger {
	return LogrusEntryWrapper{lw.Logger.WithField(key, value)}
}

type LogrusEntryWrapper struct {
	*logrus.Entry
}

func (lw LogrusEntryWrapper) WithField(key string, value interface{}) Logger {
	return LogrusEntryWrapper{lw.Entry.WithField(key, value)}
}
