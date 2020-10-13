package gofmcon

import (
	"encoding/xml"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/url"
)

const (
	fmiPath = "fmi/xml/fmresultset.xml"
	// FMDBNames adds –dbnames (Database names) query command
	FMDBNames = "-dbnames"
)

// FMConnector includes all the information about FM database to be able to connect to that
type FMConnector struct {
	Host     string
	Port     string
	Username string
	Password string
	Client   *http.Client
	Debug    bool
}

// NewFMConnector creates new FMConnector object
func NewFMConnector(host string, port string, username string, password string) *FMConnector {
	newConn := &FMConnector{
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
		Client:   http.DefaultClient,
	}

	return newConn
}

// SetDebug sets debug level of logger to Debug.
// DON'T use it in production. Your record information can leak to the logs
func (fmc *FMConnector) SetDebug(v bool) {
	fmc.Debug = v
}

// Ping sends a simple request querying all available databases
// in order to check connection and credentials
func (fmc *FMConnector) Ping() error {
	var newURL = &url.URL{}
	newURL.Scheme = "http"
	newURL.Host = fmc.Host
	newURL.Path = fmiPath
	newURL.RawQuery = newURL.Query().Encode() + "&" + FMDBNames
	request, err := http.NewRequest("GET", newURL.String(), nil)
	if err != nil {
		return errors.WithMessage(err, "gofmcon.Ping: error create request")
	}
	request.SetBasicAuth(fmc.Username, fmc.Password)
	request.Header.Set("User-Agent", "Golang FileMaker Connector")
	client := &http.Client{}
	res, err := client.Do(request)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return errors.New("gofmcon.Ping: FileMaker server unreachable")
	}

	return nil
}

// Query fetches FMResultset from FileMaker server depending on FMQuery
// given to it
func (fmc *FMConnector) Query(q *FMQuery) (FMResultset, error) {
	resultSet := FMResultset{}
	queryURL := fmc.makeURL(q)

	request, err := http.NewRequest("GET", queryURL, nil)
	if err != nil {
		return resultSet, errors.WithMessage(err, "gofmcon.Query: error create request")
	}
	request.Header.Set("User-Agent", "Golang FileMaker Connector")
	request.SetBasicAuth(fmc.Username, fmc.Password)

	if fmc.Client == nil {
		fmc.Client = http.DefaultClient
	}

	res, err := fmc.Client.Do(request)
	if err != nil {
		return resultSet, errors.WithMessage(err, "gofmcon.Query: error http request")
	}
	defer res.Body.Close()

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return resultSet, errors.WithMessage(err, "gofmcon.Query: error read response body")
	}

	if res.StatusCode == 401 {
		return resultSet, errors.New("gofmcon.Query: unauthorized")
	}

	if res.StatusCode < 200 || res.StatusCode > 299 {
		if fmc.Debug {
			logrus.Infof("gofmcon.Query unknown error: %s", string(b))
		}
		return resultSet, errors.Errorf("gofmcon.Query: unknown error with status code: %d", res.StatusCode)
	}

	err = xml.Unmarshal(b, &resultSet)
	if err != nil {
		return resultSet, errors.WithMessage(err, "gofmcon.Query: error unmarshal xml")
	}

	if resultSet.HasError() {
		return resultSet, errors.New(resultSet.FMError.String())
	}

	resultSet.prepareRecords()

	return resultSet, nil
}

func (fmc *FMConnector) makeURL(q *FMQuery) string {
	var newURL = &url.URL{}
	newURL.Scheme = "http"
	newURL.Host = fmc.Host
	if fmc.Port != "" {
		newURL.Host += ":" + fmc.Port
	}
	newURL.Path = fmiPath
	return newURL.String() + "?" + q.QueryString()
}
