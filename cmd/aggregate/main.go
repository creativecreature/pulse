package main

import (
	"flag"
	"fmt"
	"os"

	"code-harvest.conner.dev/domain"
	"code-harvest.conner.dev/logger"
	"code-harvest.conner.dev/storage"
)

// ldflags
var (
	uri string
	db  string
)

// aggregateByDay takes all the temporary coding sessions, merges them by day
// of occurrence, and moves them to a database. Once that is complete it clears
// the temporary storage of all files.
func aggregateByDay(log *logger.Logger, tempStorage storage.TemporaryStorage, permStorage storage.PermanentStorage) {
	log.PrintInfo("Performing aggregation by day", nil)
	tempSessions, err := tempStorage.Read()
	if err != nil {
		log.PrintFatal(err, nil)
	}
	err = permStorage.Write(tempSessions.Aggregate())
	if err != nil {
		log.PrintFatal(err, nil)
	}
	err = tempStorage.Clean()
	if err != nil {
		log.PrintFatal(err, nil)
	}
	log.PrintInfo("Finished aggregation by day", nil)
}

// periodString turns a time period into a readable string
func periodString(timePeriod domain.TimePeriod) string {
	switch timePeriod {
	case domain.Day:
		return "day"
	case domain.Week:
		return "week"
	case domain.Month:
		return "month"
	case domain.Year:
		return "year"
	}
	panic("Unknown time period")
}

// aggregateByTimePeriod gathers all daily coding sessions, and further
// consolidates them by week, month, or year.
func aggregateByTimePeriod(log *logger.Logger, timePeriod domain.TimePeriod, permStorage storage.PermanentStorage) {
	pString := periodString(timePeriod)
	log.PrintInfo(fmt.Sprintf("Performing aggregation by %s", pString), nil)
	err := permStorage.Aggregate(timePeriod)
	if err != nil {
		log.PrintFatal(err, nil)
	}
	log.PrintInfo(fmt.Sprintf("Finished aggregation by %s", pString), nil)
}

func main() {
	log := logger.New(os.Stdout, logger.LevelInfo)
	diskStorage := storage.DiskStorage()
	mongoStorage, disconnect := storage.MongoStorage(uri, db)
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
		aggregateByTimePeriod(log, domain.Week, mongoStorage)
	}

	if *month {
		aggregateByTimePeriod(log, domain.Month, mongoStorage)
	}

	if *year {
		aggregateByTimePeriod(log, domain.Year, mongoStorage)
	}
}
