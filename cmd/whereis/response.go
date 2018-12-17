package main

import (
	"fmt"
	"time"
)

type whereis struct {
	Content []content `json:"content"`
	Size    int       `json:"size"`
	Number  int       `json:"number"`
	Sort    []struct {
		Direction  string `json:"direction"`
		Property   string `json:"property"`
		IgnoreCase bool   `json:"ignoreCase"`
		Ascending  bool   `json:"ascending"`
	} `json:"sort"`
	NumberOfElements int  `json:"numberOfElements"`
	TotalPages       int  `json:"totalPages"`
	TotalElements    int  `json:"totalElements"`
	FirstPage        bool `json:"firstPage"`
	LastPage         bool `json:"lastPage"`
}

type content struct {
	CreateTime  string `json:"createTime"`
	EmpNo       string `json:"empNo"`
	EmpName     string `json:"empName"`
	WorkPlace   string `json:"workPlace"`
	AbsenceDate string `json:"absenceDate"`
	AbsenceTime string `json:"absenceTime"`
	AbsenceType string `json:"absenceType"`
	AbsenceDesc string `json:"absenceDesc"`
}

func (w *whereis) summary() string {
	return fmt.Sprintf("Showing %v to %v of %v items, %v of %v pages",
		w.Size*w.Number+1,
		w.Size*w.Number+w.NumberOfElements,
		w.TotalElements,
		w.Number+1,
		w.TotalPages)
}

func (c *content) place() string {
	return c.WorkPlace
}

func (c *content) name() string {
	return c.EmpName
}

func (c *content) date() string {
	t, err := time.Parse("20060102", c.AbsenceDate)
	if err != nil {
		fmt.Printf(err.Error())
		return fmt.Sprintf("%s %s", c.AbsenceDate, c.AbsenceTime)
	}
	return fmt.Sprintf("%s %s", t.Format(layout+` Mon`), c.AbsenceTime)
}

func (c *content) whereTo() string {
	return fmt.Sprintf("%s %s", c.AbsenceType, c.AbsenceDesc)
}
