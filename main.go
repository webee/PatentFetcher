package main

import (
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"

	"github.com/webee/x/xconfig"
)

type (
	// Status is the status of application
	Status struct {
		Startup time.Time `json:"startup"`
	}
)

var (
	log            = logrus.New()
	config         = new(Config)
	status         = new(Status)
	finishedPages  *BitSet
	resultsChannel = make(chan *Result, 100)

	// assignPages
	lock              sync.Mutex
	assignedPages     = make(map[int]*time.Time, 1000)
	minUnfinishedPage int
)

func init() {
	// logger
	log.SetLevel(logrus.DebugLevel)
	log.SetFormatter(&logrus.TextFormatter{
		ForceColors:   true,
		FullTimestamp: true,
	})

	// config
	xconfig.Load(config)
	// load finished pages bit set
	finishedPages = loadFinishedPages()
	// minUnfinishedPage
	minUnfinishedPage = finishedPages.MinNotExistsFrom(0)
}

func loadFinishedPages() *BitSet {
	var (
		err error
		f   *os.File
	)
	f, err = os.Open(config.PagesFilePath)
	if os.IsNotExist(err) {
		if f, err = os.Create(config.PagesFilePath); err != nil {
			panic(err)
		}
		f.Close()
		return NewBitSetFromBytes(nil)
	} else if err != nil {
		panic(err)
	}
	bs, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}
	f.Close()
	return NewBitSetFromBytes(bs)
}

func main() {
	StartPProfListen(":6060")

	// workers
	go updatePagesFile()
	go writeResults()
	go checkAssignedPages()

	e := echo.New()

	// Echo debug setting
	if config.Debug {
		e.Debug = true
	}

	// common middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// routes
	e.GET("/status", getStatus)

	e.GET("/task", applyForTask)
	e.POST("/result", submitTaskResult)

	status.Startup = time.Now()
	e.Logger.Fatal(e.Start(config.Address))
}

// API状态 成功204 失败500
func getStatus(c echo.Context) error {
	return c.JSON(http.StatusOK, status)
}

func updatePagesFile() {
	ticker := time.NewTicker(15 * time.Second)
	for range ticker.C {
		if err := ioutil.WriteFile(config.PagesFilePath, finishedPages.Bytes(), 0644); err != nil {
			log.WithError(err).WithField("file", config.PagesFilePath).Error("updatePagesFile")
		} else {
			log.WithField("file", config.PagesFilePath).Info("updatePagesFile")
		}
	}
}

func checkAssignedPages() {
	ticker := time.NewTicker(15 * time.Second)
	for range ticker.C {
		lock.Lock()
		expired := 0
		minExpiredPage := maxPage + 1
		nt := time.Now()
		for p, t := range assignedPages {
			if nt.After(t.Add(15 * time.Second)) {
				expired++
				// expired
				delete(assignedPages, p)
				if p < minExpiredPage {
					minExpiredPage = p
				}
			}
		}
		if expired > 0 {
			// from start
			minUnfinishedPage = minExpiredPage
		}
		lock.Unlock()
		log.WithFields(logrus.Fields{"expired": expired, "minUnfinishedPage": minUnfinishedPage}).Info("checkAssignedPages")
	}
}

func writeResults() {
	for res := range resultsChannel {
		if finishedPages.Has(res.Page) {
			log.WithFields(logrus.Fields{"page": res.Page, "warn": "already finished"}).Warn("writeResults")
			continue
		}

		idx := res.Page / 10000
		path := config.ResultDirPath + strconv.Itoa(idx)
		f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			log.WithError(err).WithFields(logrus.Fields{"file": path, "page": res.Page}).Error("writeResults")
			f.Close()
			continue
		}
		if _, err := io.WriteString(f, res.Content); err != nil {
			log.WithError(err).WithFields(logrus.Fields{"file": path, "page": res.Page}).Error("writeResults")
			f.Close()
			continue
		}
		f.Close()

		markAssignedFinished(res.Page)
	}
}

func markAssignedFinished(p int) {
	// add page
	finishedPages.Add(p)

	lock.Lock()
	delete(assignedPages, p)
	lock.Unlock()
}

func assignPages(n int, maxPage int) []int {
	lock.Lock()
	minPage := minUnfinishedPage
	lock.Unlock()

	pages := make([]int, 0, n)
ASSIGNING:
	for i := 0; i < n; i++ {
		for finishedPages.Has(minPage) {
			minPage = finishedPages.MinNotExistsFrom(minPage + 1)
		}
		if minPage > maxPage {
			break ASSIGNING
		}
		pages = append(pages, minPage)
		minPage++
	}
	t := time.Now()
	lock.Lock()
	for _, p := range pages {
		assignedPages[p] = &t
	}
	minUnfinishedPage = minPage
	lock.Unlock()
	return pages
}
