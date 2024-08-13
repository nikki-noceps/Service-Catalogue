package logger

type Logger interface {
	DEBUG(msg string, tags ...Tag)
	INFO(msg string, tags ...Tag)
	WARN(msg string, tags ...Tag)
	ERROR(msg string, tags ...Tag)
	FATAL(msg string, tags ...Tag)
	WITH(tags ...Tag) Logger
}
