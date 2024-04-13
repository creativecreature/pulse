package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/creativecreature/pulse"
	"github.com/creativecreature/pulse/disk"
	"github.com/creativecreature/pulse/logger"
	"github.com/creativecreature/pulse/mongo"
)

// ldflags.
var (
	uri string
	db  string
)

// aggregateByDay takes all the temporary coding sessions, merges
// them by day of occurrence, and moves them to a database. Once
// that is complete it clears the temporary storage of all files.
func aggregateByDay(log *logger.Logger, tempStorage pulse.TemporaryStorage, s pulse.PermanentStorage) {
	log.PrintInfo("Performing aggregation by day", nil)
	tempSessions, err := tempStorage.Read()
	if err != nil {
		log.PrintFatal(err, nil)
	}
	err = s.Write(tempSessions.Aggregate())
	if err != nil {
		log.PrintFatal(err, nil)
	}
	err = tempStorage.Clean()
	if err != nil {
		log.PrintFatal(err, nil)
	}
	log.PrintInfo("Finished aggregation by day", nil)
}

// periodString turns a time period into a readable string.
func periodString(timePeriod pulse.Period) string {
	switch timePeriod {
	case pulse.Day:
		return "day"
	case pulse.Week:
		return "week"
	case pulse.Month:
		return "month"
	case pulse.Year:
		return "year"
	}
	panic("Unknown time period")
}

// aggregateByTimePeriod gathers all daily coding sessions,
// and further consolidates them by week, month, or year.
func aggregateByTimePeriod(log *logger.Logger, tp pulse.Period, s pulse.PermanentStorage) {
	pString := periodString(tp)
	log.PrintInfo(fmt.Sprintf("Performing aggregation by %s", pString), nil)
	err := s.Aggregate(tp)
	if err != nil {
		log.PrintFatal(err, nil)
	}
	log.PrintInfo(fmt.Sprintf("Finished aggregation by %s", pString), nil)
}

func main() {
	log := logger.New(os.Stdout, logger.LevelInfo)
	diskStorage := disk.NewStorage()
	mongoStorage, disconnect := mongo.New(uri, db)
	defer disconnect()

	day := flag.Bool("day", false, "aggregate raw coding sessions by day")
	week := flag.Bool("week", false, "aggregate daily coding sessions into weekly summaries")
	month := flag.Bool("month", false, "aggregate daily coding sessions into monthly summaries")
	year := flag.Bool("year", false, "aggregate daily coding sessions into yearly summaries")
	flag.Parse()

	if *day {
		aggregateByDay(log, diskStorage, mongoStorage)
	}

	if *week {
		aggregateByTimePeriod(log, pulse.Week, mongoStorage)
	}

	if *month {
		aggregateByTimePeriod(log, pulse.Month, mongoStorage)
	}

	if *year {
		aggregateByTimePeriod(log, pulse.Year, mongoStorage)
	}
}
