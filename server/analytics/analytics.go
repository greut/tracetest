package analytics

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"

	"github.com/google/uuid"
)

const (
	gaURL           = "https://www.google-analytics.com/mp/collect?measurement_id=%s&api_secret=%s"
	gaValidationURL = "https://www.google-analytics.com/debug/mp/collect?measurement_id=%s&api_secret=%s"
)

var defaultClient ga

var (
	MeasurementID = ""
	SecretKey     = ""
)

func Init(enabled bool, serverID, appName, appVersion string) error {
	// ga not enabled, use dumb settings
	if !enabled {
		defaultClient = ga{enabled: false}
		return nil
	}

	// setup an actual client
	hostname, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("could not get hostname: %w", err)
	}

	defaultClient = ga{
		enabled:       true,
		measurementID: MeasurementID,
		secretKey:     SecretKey,
		appVersion:    appVersion,
		appName:       appName,
		hostname:      hostname,
		serverID:      serverID,
	}

	return nil
}

func CreateAndSendEvent(name, category string) error {
	fmt.Printf(`sending event "%s" (%s)%s`, name, category, "\n")
	if !defaultClient.ready() {
		err := fmt.Errorf("uninitalized cliGent. Call analytics.Init")
		fmt.Printf(`could not send event "%s" (%s): %s%s`, name, category, err.Error(), "\n")
		return err
	}
	err := defaultClient.CreateAndSendEvent(name, category)
	if err != nil {
		fmt.Printf(`could not send event "%s" (%s): %s%s`, name, category, err.Error(), "\n")
	} else {
		fmt.Printf(`event sent "%s" (%s)%s`, name, category, "\n")
	}

	return err
}

func Ready() bool {
	return defaultClient.ready()
}

type ga struct {
	enabled       bool
	appVersion    string
	appName       string
	measurementID string
	secretKey     string
	hostname      string
	serverID      string
}

func (ga ga) ready() bool {
	return !ga.enabled || (ga.appVersion != "" &&
		ga.appName != "" &&
		ga.hostname != "" &&
		ga.serverID != "")

}

func (ga ga) CreateAndSendEvent(name, category string) error {
	if !ga.enabled {
		fmt.Printf(`Ignore event "%s" (%s): ga not enabled%s`, name, category, "\n")
		return nil
	}
	event, err := ga.newEvent(name, category)
	if err != nil {
		return fmt.Errorf("could not create event: %w", err)
	}

	return ga.sendEvent(event)
}

func (ga ga) newEvent(name, category string) (event, error) {
	return event{
		Name: name,
		Params: params{
			ID:              uuid.NewString(), // prevent event caching
			EventCount:      1,
			EventCategory:   category,
			AppName:         ga.appName,
			Host:            ga.hostname,
			MachineID:       ga.serverID,
			AppVersion:      ga.appVersion,
			Architecture:    runtime.GOARCH,
			OperatingSystem: runtime.GOOS,
		},
	}, nil
}

func (ga ga) sendEvent(e event) error {
	payload := payload{
		ClientID: ga.serverID,
		Events: []event{
			e,
		},
	}

	err := ga.sendValidationRequest(payload)
	if err != nil {
		return err
	}

	err = ga.sendDataToGA(payload)
	if err != nil {
		return fmt.Errorf("could not send request to google analytics: %w", err)
	}

	return nil
}

func (ga ga) sendValidationRequest(p payload) error {
	response, body, err := ga.sendPayloadToURL(p, gaValidationURL)

	if err != nil {
		return err
	}

	if response.StatusCode >= 300 {
		return fmt.Errorf("validation response got unexpected status. Got: %d", response.StatusCode)
	}

	validationResponse := validationResponse{}
	err = json.Unmarshal(body, &validationResponse)
	if err != nil {
		return fmt.Errorf("could not unmarshal response body: %w", err)
	}

	if len(validationResponse.ValidationMessages) > 0 {
		return fmt.Errorf("google analytics request validation failed")
	}

	return nil
}

func (ga ga) sendDataToGA(p payload) error {
	response, _, err := ga.sendPayloadToURL(p, gaURL)
	if err != nil {
		return fmt.Errorf("could not send event to google analytics: %w", err)
	}

	if response.StatusCode >= 300 {
		return fmt.Errorf("google analytics returned unexpected status. Got: %d", response.StatusCode)
	}

	return nil
}

func (ga ga) sendPayloadToURL(p payload, url string) (*http.Response, []byte, error) {
	jsonData, err := json.Marshal(p)
	if err != nil {
		return nil, []byte{}, fmt.Errorf("could not marshal json payload: %w", err)
	}

	request, err := http.NewRequest("POST", fmt.Sprintf(url, ga.measurementID, ga.secretKey), bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, []byte{}, fmt.Errorf("could not create request: %w", err)
	}

	request.Header.Set("Content-Type", "application/json")

	client := http.DefaultClient

	resp, err := client.Do(request)
	if err != nil {
		return nil, []byte{}, fmt.Errorf("could not execute request: %w", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, []byte{}, fmt.Errorf("could not read response body: %w", err)
	}

	return resp, body, err
}

type params struct {
	ID               string `json:"tt_event_id,omitempty"`
	EventCount       int64  `json:"event_count,omitempty"`
	EventCategory    string `json:"event_category,omitempty"`
	AppVersion       string `json:"app_version,omitempty"`
	AppName          string `json:"app_name,omitempty"`
	CustomDimensions string `json:"custom_dimensions,omitempty"`
	DataSource       string `json:"data_source,omitempty"`
	Host             string `json:"host,omitempty"`
	MachineID        string `json:"machine_id,omitempty"`
	OperatingSystem  string `json:"operating_system,omitempty"`
	Architecture     string `json:"architecture,omitempty"`
}

type event struct {
	Name   string `json:"name"`
	Params params `json:"params,omitempty"`
}

type payload struct {
	ClientID string  `json:"client_id,omitempty"`
	Events   []event `json:"events,omitempty"`
}

type validationResponse struct {
	ValidationMessages []validationMessage `json:"validationMessages"`
}

type validationMessage struct {
	FieldPath      string `json:"fieldPath"`
	Description    string `json:"description"`
	ValidationCode string `json:"validationCode"`
}
