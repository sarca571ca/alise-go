package models

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type VanaTimeResponse struct {
	FormattedTime       string `json:"formatted_time"`
	Hour                int    `json:"hours"`
	Minute              int    `json:"minutes"`
	Second              int    `json:"seconds"`
	FormattedDate       string `json:"formatted_date"`
	Weekday             string `json:"weekday"`
	Day                 int    `json:"day"`
	Month               int    `json:"month"`
	Year                int    `json:"year"`
	MoonPhaseName       string `json:"moon_phase"`
	MoonPhase           int    `json:"moon_percent"`
	IsDaytime           bool   `json:"is_daytime"`
	SecondsUntilNextDay int    `json:"seconds_until_next_day"`
	FormattedDateTime   string `json:"formatted_datetime"`
	RealTime            string `json:"real_time"`
}

type vanaTimeAPIResponse struct {
	OK   bool             `json:"ok"`
	Data VanaTimeResponse `json:"data"`
}

type VaneTime struct {
	Epoch int
	Year  int
	Month int
	Week  int
	Day   int
	Hour  int
}

func FetchCurrentVanaTime() (*VanaTimeResponse, error) {
	req, err := http.NewRequest(http.MethodGet, "https://vana-time.com/api/v1/time/vanadiel", nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("vana-time api returned %s", resp.Status)
	}

	var apiResp vanaTimeAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, err
	}

	if !apiResp.OK {
		return nil, fmt.Errorf("vana-time api returned ok=false")
	}

	return &apiResp.Data, nil
}

func AbsoluteVanaDay(year, month, day int) int64 {
	const baseYear = 886
	return int64((year-baseYear)*360 + (month-1)*30 + (day - 1))
}

func CurrentVanaDayFromUnix(unix int64) int64 {
	return currentVanaDayFromUnix(unix)
}

func VanaDayStartUnix(vanaDay int64) int64 {
	return VanaBaseEpoch + (vanaDay * VanaDaySeconds)
}

func VanaDayStartTime(vanaDay int64) time.Time {
	return time.Unix(VanaDayStartUnix(vanaDay), 0).UTC()
}

func NewVanaTime() VaneTime {
	return VaneTime{
		Epoch: 1009810800,
		Year:  518400,
		Month: 43200,
		Week:  11520,
		Day:   1440,
		Hour:  60,
	}
}

func VanaYMDFromDay(ut int64) (year, month, day int64) {
	const baseYear int64 = 886
	const daysPerYear int64 = 360
	const daysPerMonth int64 = 30

	year = ut/daysPerYear + baseYear
	month = (ut%daysPerYear)/daysPerMonth + 1
	day = (ut % daysPerMonth) + 1
	return
}
