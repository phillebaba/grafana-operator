package grafanadashboard

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/integr8ly/grafana-operator/v3/pkg/apis/integreatly/v1alpha1"
)

const (
	DeleteDashboardByUIDUrl    = "%v/api/dashboards/uid/%v"
	CreateOrUpdateDashboardUrl = "%v/api/dashboards/db"
	CreateOrUpdateFolderUrl    = "%v/api/folders"
	DeleteFolderByUIDUrl       = "%v/api/folders/%v"
	GetFolderByIDUrl           = "%v/api/folders/id/%v"
	//TODO is this right? VV
	SearchURL         = "%v/api/search?type=dash-db"
	CreatePlaylistUrl = "%v/api/playlists/"
	DeletePlaylistUrl = "%v/api/playlists/%v"
)

const (
	NonNamespacedFolderName = "Non-Namespaced"
)

type GrafanaRequest struct {
	Dashboard  json.RawMessage `json:"dashboard"`
	FolderId   int64           `json:"folderId"`
	FolderName string          `json:"folderName"`
	Overwrite  bool            `json:"overwrite"`
}

type GrafanaResponse struct {
	ID         *uint   `json:"id"`
	OrgID      *uint   `json:"orgId"`
	Message    *string `json:"message"`
	Slug       *string `json:"slug"`
	Version    *int    `json:"version"`
	Status     *string `json:"resp"`
	UID        *string `json:"uid"`
	URL        *string `json:"url"`
	FolderId   *int64  `json:"folderId"`
	FolderName string  `json:"folderName"`
}

type GrafanaFolderRequest struct {
	Title string `json:"title"`
}

type GrafanaFolderResponse struct {
	ID    *int64 `json:"id"`
	Title string `json:"title"`
	UID   string `json:"uid"`
}

type GrafanaClient interface {
	CreateOrUpdateDashboard(dashboard []byte, folderId int64, folderName string) (GrafanaResponse, error)
	DeleteDashboardByUID(UID string) (GrafanaResponse, error)
	CreateOrUpdateFolder(folderName string) (GrafanaFolderResponse, error)
	DeleteFolder(folderID *int64) error
	SafeToDelete(dashboards []*v1alpha1.GrafanaDashboardRef, folderID *int64) bool
	CreatePlaylist(playlist []byte) error
	DeletePlaylist() error
}

type GrafanaClientImpl struct {
	url      string
	user     string
	password string
	client   *http.Client
}

func setHeaders(req *http.Request) {
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "grafana-operator")
}

func NewGrafanaClient(url, user, password string, transport *http.Transport, timeoutSeconds time.Duration) GrafanaClient {
	client := &http.Client{
		Transport: transport,
		Timeout:   time.Second * timeoutSeconds,
	}

	return &GrafanaClientImpl{
		url:      url,
		user:     user,
		password: password,
		client:   client,
	}
}

func (r *GrafanaClientImpl) getAllFolders() ([]GrafanaFolderResponse, error) {
	rawURL := fmt.Sprintf(CreateOrUpdateFolderUrl, r.url)
	parsed, err := url.Parse(rawURL)

	if err != nil {
		return nil, err
	}

	parsed.User = url.UserPassword(r.user, r.password)
	req, err := http.NewRequest("GET", parsed.String(), nil)

	if err != nil {
		return nil, err
	}

	setHeaders(req)

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf(
			"error creating folder, expected status 200 but got %v",
			resp.StatusCode)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var folders []GrafanaFolderResponse
	err = json.Unmarshal(data, &folders)

	return folders, err
}

func (r *GrafanaClientImpl) CreateOrUpdateFolder(folderInputName string) (GrafanaFolderResponse, error) {
	response := newFolderResponse()

	allfolders, err := r.getAllFolders()
	if err != nil {
		return response, err
	}

	for _, folder := range allfolders {
		if folder.Title == folderInputName {
			return folder, nil
		}
	}

	rawURL := fmt.Sprintf(CreateOrUpdateFolderUrl, r.url)
	parsed, err := url.Parse(rawURL)

	if err != nil {
		return response, err
	}

	var title = folderInputName
	if title == "" {
		title = NonNamespacedFolderName
	}

	raw, err := json.Marshal(GrafanaFolderRequest{
		Title: title,
	})
	if err != nil {
		return response, err
	}

	parsed.User = url.UserPassword(r.user, r.password)
	req, err := http.NewRequest("POST", parsed.String(), bytes.NewBuffer(raw))

	if err != nil {
		return response, err
	}

	setHeaders(req)

	resp, err := r.client.Do(req)
	if err != nil {
		return response, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return response, fmt.Errorf(
			"error creating folder, expected status 200 but got %v",
			resp.StatusCode)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return response, err
	}

	err = json.Unmarshal(data, &response)

	return response, err
}

// Submit dashboard json to grafana
func (r *GrafanaClientImpl) CreateOrUpdateDashboard(dashboard []byte, folderID int64, folderName string) (GrafanaResponse, error) {
	rawURL := fmt.Sprintf(CreateOrUpdateDashboardUrl, r.url)
	response := newResponse()

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return response, err
	}

	// Grafana expects some additional data along with the dashboard
	raw, err := json.Marshal(GrafanaRequest{
		Dashboard: dashboard,

		FolderId: folderID,

		FolderName: folderName,

		// We always want to set `overwrite` because the uids in the CRs map
		// directly to dashboards in grafana
		Overwrite: true,
	})
	if err != nil {
		return response, err
	}

	parsed.User = url.UserPassword(r.user, r.password)
	req, err := http.NewRequest("POST", parsed.String(), bytes.NewBuffer(raw))

	if err != nil {
		return response, err
	}

	setHeaders(req)

	resp, err := r.client.Do(req)
	if err != nil {
		return response, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return response, fmt.Errorf(
			"error creating dashboard, expected status 200 but got %v",
			resp.StatusCode)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return response, err
	}

	err = json.Unmarshal(data, &response)

	return response, err
}

// Delete a dashboard given by a UID
func (r *GrafanaClientImpl) DeleteDashboardByUID(UID string) (GrafanaResponse, error) {
	rawURL := fmt.Sprintf(DeleteDashboardByUIDUrl, r.url, UID)
	response := newResponse()

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return response, err
	}

	parsed.User = url.UserPassword(r.user, r.password)
	req, err := http.NewRequest("DELETE", parsed.String(), nil)

	if err != nil {
		return response, err
	}

	setHeaders(req)

	resp, err := r.client.Do(req)
	if err != nil {
		return response, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return response, fmt.Errorf(
			"error deleting dashboard, expected status 200 but got %v",
			resp.StatusCode)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return response, err
	}

	err = json.Unmarshal(data, &response)

	return response, err
}

func newFolderResponse() GrafanaFolderResponse {
	var id int64 = 0

	return GrafanaFolderResponse{
		ID: &id,
	}
}

func newResponse() GrafanaResponse {
	var id uint = 0
	var orgID uint = 0
	var version int = 0
	var status = "(empty)"
	var message = "(empty)"
	var slug string
	var uid string
	var url string

	return GrafanaResponse{
		ID:      &id,
		OrgID:   &orgID,
		Message: &message,
		Slug:    &slug,
		Version: &version,
		Status:  &status,
		UID:     &uid,
		URL:     &url,
	}
}

func (r *GrafanaClientImpl) getFolderUID(IDInput *int64) (*string, error) {
	rawURL := fmt.Sprintf(GetFolderByIDUrl, r.url, *IDInput)
	response := newFolderResponse()

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	parsed.User = url.UserPassword(r.user, r.password)
	req, err := http.NewRequest("GET", parsed.String(), nil)

	if err != nil {
		return nil, err
	}

	setHeaders(req)

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf(
			"error finding folder %v, expected status 200 but got %v", IDInput,
			resp.StatusCode)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, &response)
	if err != nil {
		return nil, err
	}

	return &response.UID, nil
}

func (r *GrafanaClientImpl) DeleteFolder(deleteID *int64) error {

	deleteUID, err := r.getFolderUID(deleteID)
	if err != nil {
		return err
	}

	rawURL := fmt.Sprintf(DeleteFolderByUIDUrl, r.url, *deleteUID)
	response := newResponse()

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return err
	}

	parsed.User = url.UserPassword(r.user, r.password)
	req, err := http.NewRequest("DELETE", parsed.String(), nil)

	if err != nil {
		return err
	}

	setHeaders(req)

	resp, err := r.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf(
			"error deleting folder, expected status 200 but got %v",
			resp.StatusCode)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, &response)
	log.V(1).Info(fmt.Sprintf("delete result was %v", *response.Message))

	return err
}

func (r *GrafanaClientImpl) SafeToDelete(dashlist []*v1alpha1.GrafanaDashboardRef, id *int64) bool {
	for _, dashboard := range dashlist {
		if *dashboard.FolderId == *id {
			return false
		}
	}
	return true
}

type GrafanaPlaylistRequest struct {
	Name     string                       `json:"name"`
	Interval time.Duration                `json:"interval"`
	Items    []GrafanaPlaylistRequestItem `json:"items"`
}

type GrafanaPlaylistRequestItem struct {
	Type  string `json:"type"`
	Value string `json:"value"`
	Title string `json:"title"`
	Order string `json:"order"`
}

type GrafanaPlaylistResponse struct {
	ID       int           `json:"id"`
	Name     string        `json:"name"`
	Interval time.Duration `json:"interval"`
}

func (r *GrafanaClientImpl) CreatPlaylist(grafanaRequest GrafanaPlaylistRequest) (*GrafanaPlaylistResponse, error) {
	rawURL := fmt.Sprintf(CreatePlaylistUrl, r.url)
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	parsed.User = url.UserPassword(r.user, r.password)

	raw, err := json.Marshal(grafanaRequest)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", parsed.String(), bytes.NewBuffer(raw))
	if err != nil {
		return nil, err
	}
	setHeaders(req)

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("error creating playlist, expected status 200 but got %v", resp.StatusCode)
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	response := &GrafanaPlaylistResponse{}
	err = json.Unmarshal(data, response)
	return response, nil
}

func (r *GrafanaClientImpl) DeletePlaylist(playlistID string) error {
	rawURL := fmt.Sprintf(DeletePlaylistUrl, r.url, playlistID)
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return err
	}
	parsed.User = url.UserPassword(r.user, r.password)

	req, err := http.NewRequest("DELETE", parsed.String(), nil)
	if err != nil {
		return err
	}
	setHeaders(req)

	resp, err := r.client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("error deleting playlist, expected status 200 but got %v", resp.StatusCode)
	}
	return nil
}
