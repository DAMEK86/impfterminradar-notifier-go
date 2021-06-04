package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Vaccines struct {
	Id   string `json:"ID"`
	Slug string `json:"Slug"`
}

type VaccinationCenter struct {
	Name     string     `json:"Zentrumsname"`
	Zip      string     `json:"PLZ"`
	City     string     `json:"Ort"`
	State    string     `json:"BundeslandRealName"`
	BaseUrl  string     `json:"BookingURL"`
	Address  string     `json:"Adress"`
	Slug     string     `json:"Slug"`
	Vaccines []Vaccines `json:"Vaccines"`
}

func (v *VaccinationCenter) GetVaccineNameBySlug(slug string) string {
	for _, vaccine := range v.Vaccines {
		if vaccine.Slug == slug {
			return vaccine.Id
		}
	}
	return ""
}

type AvailableVaccines struct {
	Available    bool   `json:"Available"`
	NoBooking    bool   `json:"NoBooking"`
	Slug         string `json:"Slug"`
	Time         int64  `json:"Time"`
	Unknown      bool   `json:"Unknown"`
	WaitingRoom  bool   `json:"WaitingRoom"`
	FriendlyName string
	Center       VaccinationCenter
}

const (
	baseUrl              = "https://impfterminradar.de/api/"
	availabilityEndpoint = "vaccinations/availability"
	centersEndpoint      = "centers?PLZ=%s&Radius=%d"
)

type Client interface {
	GetVacationCenters(center string, radius int) (centers []VaccinationCenter, err error)
	GetVaccinesIn(centers []VaccinationCenter) (availableVaccines []AvailableVaccines, err error)
}

type client struct {
	httpClient http.Client
}

func NewClient(httpClient http.Client) Client {
	return &client{
		httpClient: httpClient,
	}
}

func (c *client) GetVacationCenters(center string, radius int) (centers []VaccinationCenter, err error) {
	reqUrl := baseUrl + fmt.Sprintf(centersEndpoint, center, radius)
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

func (c *client) GetVaccinesIn(centers []VaccinationCenter) (availableVaccines []AvailableVaccines, err error) {
	reqUrl := baseUrl + availabilityEndpoint
	requestVaccines := make([]string, 0)
	for _, center := range centers {
		for _, vaccine := range center.Vaccines {
			requestVaccines = append(requestVaccines, vaccine.Slug)
		}
	}

	requestBody, err := json.Marshal(requestVaccines)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPatch, reqUrl, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("content-type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&availableVaccines)
	if err != nil {
		return nil, err
	}

	for i := range availableVaccines {
		setName(centers, &availableVaccines[i])
	}
	return availableVaccines, nil
}

func setName(centers []VaccinationCenter, availableVaccines *AvailableVaccines) {
	for _, center := range centers {
		name := center.GetVaccineNameBySlug(availableVaccines.Slug)
		if name != "" {
			availableVaccines.FriendlyName = name
			availableVaccines.Center = center
			return
		}
	}
}
