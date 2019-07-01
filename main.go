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
		Startup time.Time    `json:"startup"`
		Fetcher *FetcherInfo `json:"fetcher"`
	}

	// FetcherInfo is fetcher's info
	FetcherInfo struct {
		finishedPages  *BitSet
		resultsChannel chan *Result
		// assignPages
		lock              sync.Mutex
		assignedPages     map[int]*time.Time
		MinUnfinishedPage int     `json:"minUnifinishedPage"`
		Rate              float64 `json:"rate"`
	}
)

var (
	log         = logrus.New()
	config      = new(Config)
	status      = new(Status)
	fetcherInfo = &FetcherInfo{
		resultsChannel: make(chan *Result, 100),
		assignedPages:  make(map[int]*time.Time, 1000),
	}
)

func init() {
	// logger
	log.SetLevel(logrus.InfoLevel)
	log.SetFormatter(&logrus.TextFormatter{
		ForceColors:   true,
		FullTimestamp: true,
	})

	// config
	xconfig.Load(config)

	status.Fetcher = fetcherInfo
	// load finished pages bit set
	fetcherInfo.finishedPages = loadFinishedPages()
	// minUnfinishedPage
	fetcherInfo.MinUnfinishedPage = fetcherInfo.finishedPages.MinNotExistsFrom(0)
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
	go fetcherInfo.updatePagesFile()
	go fetcherInfo.writeResults()
	go fetcherInfo.checkAssignedPages()

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

func (f *FetcherInfo) updatePagesFile() {
	finishedPages := f.finishedPages
	ticker := time.NewTicker(15 * time.Second)
	for range ticker.C {
		if err := ioutil.WriteFile(config.PagesFilePath, finishedPages.Bytes(), 0644); err != nil {
			log.WithError(err).WithField("file", config.PagesFilePath).Error("updatePagesFile")
		} else {
			log.WithField("file", config.PagesFilePath).Info("updatePagesFile")
		}
	}
}

func (f *FetcherInfo) checkAssignedPages() {
	assignedPages := f.assignedPages
	ticker := time.NewTicker(20 * time.Second)
	for range ticker.C {
		f.lock.Lock()
		expired := 0
		minExpiredPage := maxPage + 1
		nt := time.Now()
		for p, t := range assignedPages {
			if nt.After(t.Add(20 * time.Second)) {
				expired++
				// expired
				delete(assignedPages, p)
				if p < minExpiredPage {
					minExpiredPage = p
				}
			}
		}
		if expired > 0 {
			if minExpiredPage < f.MinUnfinishedPage {
				f.MinUnfinishedPage = minExpiredPage
			}
		}
		f.lock.Unlock()
		log.WithFields(logrus.Fields{"expired": expired, "minUnfinishedPage": f.MinUnfinishedPage}).Info("checkAssignedPages")
	}
}

func (f *FetcherInfo) writeResults() {
	startAt := time.Now()
	count := 0
	for res := range f.resultsChannel {
		if f.finishedPages.Has(res.Page) {
			log.WithFields(logrus.Fields{"page": res.Page, "warn": "already finished"}).Warn("writeResults")
			continue
		}
		if len(res.Content) < 10 {
			// 10 是随便选择的，内容长度不会小于10
			log.WithFields(logrus.Fields{"page": res.Page, "warn": "request failed"}).Warn("writeResults")
			continue
		}

		idx := res.Page / 10000
		path := config.ResultDirPath + strconv.Itoa(idx)
		resultFile, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			log.WithError(err).WithFields(logrus.Fields{"file": path, "page": res.Page}).Error("writeResults")
			resultFile.Close()
			continue
		}
		if _, err := io.WriteString(resultFile, res.Content); err != nil {
			log.WithError(err).WithFields(logrus.Fields{"file": path, "page": res.Page}).Error("writeResults")
			resultFile.Close()
			continue
		}
		resultFile.Close()

		f.markAssignedFinished(res.Page)

		// stat
		count++
		tn := time.Now()
		d := tn.Sub(startAt)
		if d > 15*time.Second {
			f.Rate = float64(count) / (float64(d) / float64(time.Second))
			log.Infof("Rate: %f/s", f.Rate)
			startAt = tn
			count = 0
		}
	}
}

func (f *FetcherInfo) markAssignedFinished(p int) {
	// add page
	f.finishedPages.Add(p)

	f.lock.Lock()
	delete(f.assignedPages, p)
	f.lock.Unlock()
}

func (f *FetcherInfo) assignPages(n int, maxPage int) []int {
	f.lock.Lock()
	minPage := f.MinUnfinishedPage

	pages := make([]int, 0, n)
ASSIGNING:
	for i := 0; i < n; i++ {
		for f.finishedPages.Has(minPage) || f.assignedPages[minPage] != nil {
			// 已完成或已分配
			minPage = f.finishedPages.MinNotExistsFrom(minPage + 1)
		}
		if minPage > maxPage {
			break ASSIGNING
		}
		pages = append(pages, minPage)
		minPage++
	}
	t := time.Now()
	for _, p := range pages {
		f.assignedPages[p] = &t
	}
	f.MinUnfinishedPage = minPage
	f.lock.Unlock()
	return pages
}
