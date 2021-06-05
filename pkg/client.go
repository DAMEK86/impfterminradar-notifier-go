package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Vaccine struct {
	Id          string `json:"ID"`
	Slug        string `json:"Slug"`
	Available   bool   `json:"Available"`
	NoBooking   bool   `json:"NoBooking"`
	Time        int64  `json:"Time"`
	Unknown     bool   `json:"Unknown"`
	WaitingRoom bool   `json:"WaitingRoom"`
}

type VaccinationCenter struct {
	Name     string    `json:"Zentrumsname"`
	Zip      string    `json:"PLZ"`
	City     string    `json:"Ort"`
	State    string    `json:"BundeslandRealName"`
	BaseUrl  string    `json:"BookingURL"`
	Address  string    `json:"Adress"`
	Slug     string    `json:"Slug"`
	Vaccines []Vaccine `json:"Vaccines"`
}

func (c *VaccinationCenter) UpdateVaccineOnMatch(v Vaccine) {
	for i := range c.Vaccines {
		if c.Vaccines[i].Slug == v.Slug {
			c.Vaccines[i].Available = v.Available
			c.Vaccines[i].NoBooking = v.NoBooking
			c.Vaccines[i].Time = v.Time
			c.Vaccines[i].Unknown = v.Unknown
			c.Vaccines[i].WaitingRoom = v.WaitingRoom
			break
		}
	}
}

const (
	baseUrl              = "https://www.impfterminradar.de/api/"
	availabilityEndpoint = "vaccinations/availability"
	centersEndpoint      = "centers?PLZ=%s&Radius=%d"
)

type Client interface {
	GetVacationCenters(zip string, radius int) (centers []VaccinationCenter, err error)
	UpdateVaccinesIn(centers []VaccinationCenter) error
}

type client struct {
	httpClient http.Client
}

func NewClient(httpClient http.Client) Client {
	return &client{
		httpClient: httpClient,
	}
}

func (c *client) GetVacationCenters(zip string, radius int) (centers []VaccinationCenter, err error) {
	reqUrl := baseUrl + fmt.Sprintf(centersEndpoint, zip, radius)
	req, err := http.NewRequest(http.MethodGet, reqUrl, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&centers)

	return centers, err
}

func (c *client) UpdateVaccinesIn(centers []VaccinationCenter) error {
	reqUrl := baseUrl + availabilityEndpoint
	requestVaccines := make([]string, 0)
	for _, center := range centers {
		for _, vaccine := range center.Vaccines {
			requestVaccines = append(requestVaccines, vaccine.Slug)
		}
	}

	requestBody, err := json.Marshal(requestVaccines)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPatch, reqUrl, bytes.NewBuffer(requestBody))
	if err != nil {
		return err
	}

	req.Header.Set("content-type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	vaccines := make([]Vaccine, 0)
	err = json.NewDecoder(resp.Body).Decode(&vaccines)
	if err != nil {
		return err
	}

	for i := range vaccines {
		for _, center := range centers {
			center.UpdateVaccineOnMatch(vaccines[i])
		}
	}
	return nil
}
