package vercel

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

const baseURL = "https://api.vercel.com"

type Client struct {
	token      string
	httpClient *http.Client
}

func NewClient(token string) *Client {
	return &Client{
		token:      token,
		httpClient: &http.Client{},
	}
}

type User struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Username string `json:"username"`
}

type Project struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type listProjectsResponse struct {
	Projects []Project `json:"projects"`
}

func (c *Client) GetUser() (*User, error) {
	var resp struct {
		User User `json:"user"`
	}
	if err := c.get("/v2/user", &resp); err != nil {
		return nil, err
	}
	return &resp.User, nil
}

func (c *Client) ListProjects() ([]Project, error) {
	var resp listProjectsResponse
	if err := c.get("/v9/projects", &resp); err != nil {
		return nil, err
	}
	return resp.Projects, nil
}

type Deployment struct {
	UID     string `json:"uid"`
	Name    string `json:"name"`
	State   string `json:"state"`
	Created int64  `json:"created"` // unix ms
}

type LogEvent struct {
	ID      string `json:"id"`
	Date    int64  `json:"date"` // unix ms
	Text    string `json:"text"`
	Type    string `json:"type"` // stdout, stderr, command, exit
	Serial  string `json:"serial"`
}

func (c *Client) ListDeployments(projectID string, limit int) ([]Deployment, error) {
	var resp struct {
		Deployments []Deployment `json:"deployments"`
	}
	path := fmt.Sprintf("/v6/deployments?projectId=%s&limit=%d", projectID, limit)
	if err := c.get(path, &resp); err != nil {
		return nil, err
	}
	return resp.Deployments, nil
}

func (c *Client) GetDeploymentEvents(ctx context.Context, deploymentID string, since int64, limit int) ([]LogEvent, error) {
	path := fmt.Sprintf("/v2/deployments/%s/events?limit=%d&direction=forward", deploymentID, limit)
	if since > 0 {
		path += "&since=" + strconv.FormatInt(since, 10)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", baseURL+path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned %d: %s", resp.StatusCode, string(body))
	}

	var events []LogEvent
	if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
		return nil, fmt.Errorf("decoding events: %w", err)
	}
	return events, nil
}

func (c *Client) get(path string, out any) error {
	req, err := http.NewRequest("GET", baseURL+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned %d: %s", resp.StatusCode, string(body))
	}

	return json.NewDecoder(resp.Body).Decode(out)
}
