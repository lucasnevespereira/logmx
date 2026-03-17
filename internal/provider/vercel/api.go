package vercel

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const vercelBase = "https://api.vercel.com"

type Client struct {
	token      string
	httpClient *http.Client
}

func NewClient(token string) *Client {
	return &Client{token: token, httpClient: &http.Client{}}
}

type User struct {
	Username string `json:"username"`
	Email    string `json:"email"`
}

func (c *Client) ValidateToken() (*User, error) {
	var resp struct {
		User User `json:"user"`
	}
	if err := c.get("/v2/user", &resp); err != nil {
		return nil, err
	}
	return &resp.User, nil
}

type Project struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (c *Client) ListProjects() ([]Project, error) {
	var resp struct {
		Projects []Project `json:"projects"`
	}
	if err := c.get("/v9/projects?limit=100", &resp); err != nil {
		return nil, err
	}
	return resp.Projects, nil
}

func (c *Client) get(path string, out any) error {
	req, err := http.NewRequest("GET", vercelBase+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("invalid or expired token")
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned %d: %s", resp.StatusCode, string(body))
	}

	return json.NewDecoder(resp.Body).Decode(out)
}
