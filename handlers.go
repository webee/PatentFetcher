package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type (
	// Project is a fetch project
	Project struct {
		StartDate string `json:"startDate"`
		EndDate   string `json:"endDate"`
	}

	// Task is task info
	Task struct {
		*Project
		Pages []int `json:"pages"`
	}

	// Result is task result
	Result struct {
		*Project
		Page    int    `json:"page"`
		Content string `json:"content"`
	}
)

var (
	maxPage = 964391
	proj    = &Project{
		StartDate: "2000.01.01",
		EndDate:   "2019.12.31",
	}
)

func applyForTask(c echo.Context) error {
	task := &Task{
		Project: proj,
		Pages:   assignPages(3, maxPage),
	}
	return c.JSON(http.StatusOK, task)
}

func submitTaskResult(c echo.Context) (err error) {
	res := new(Result)
	if err = c.Bind(res); err != nil {
		return
	}

	resultsChannel <- res
	return c.NoContent(http.StatusNoContent)
}
