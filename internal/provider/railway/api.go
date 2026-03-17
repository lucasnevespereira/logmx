package railway

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const railwayGraphQL = "https://backboard.railway.com/graphql/v2"

type Client struct {
	token      string
	httpClient *http.Client
}

func NewClient(token string) *Client {
	return &Client{token: token, httpClient: &http.Client{}}
}

type Project struct {
	ID       string
	Name     string
	Services []Service
}

type Service struct {
	ID   string
	Name string
}

func (c *Client) ValidateToken() (string, error) {
	var resp struct {
		Data struct {
			Projects struct {
				Edges []struct {
					Node struct {
						Name string `json:"name"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"projects"`
		} `json:"data"`
	}

	if err := c.query(`{ projects { edges { node { name } } } }`, &resp); err != nil {
		return "", err
	}

	n := len(resp.Data.Projects.Edges)
	if n == 0 {
		return "0 projects", nil
	}
	return fmt.Sprintf("%d project(s)", n), nil
}

func (c *Client) ListProjects() ([]Project, error) {
	var resp struct {
		Data struct {
			Projects struct {
				Edges []struct {
					Node struct {
						ID       string `json:"id"`
						Name     string `json:"name"`
						Services struct {
							Edges []struct {
								Node struct {
									ID   string `json:"id"`
									Name string `json:"name"`
								} `json:"node"`
							} `json:"edges"`
						} `json:"services"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"projects"`
		} `json:"data"`
	}

	q := `{ projects { edges { node { id name services { edges { node { id name } } } } } } }`
	if err := c.query(q, &resp); err != nil {
		return nil, err
	}

	var projects []Project
	for _, e := range resp.Data.Projects.Edges {
		p := Project{ID: e.Node.ID, Name: e.Node.Name}
		for _, se := range e.Node.Services.Edges {
			p.Services = append(p.Services, Service{
				ID:   se.Node.ID,
				Name: se.Node.Name,
			})
		}
		projects = append(projects, p)
	}
	return projects, nil
}

type gqlResponse struct {
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

func (c *Client) query(q string, out any) error {
	body, err := json.Marshal(map[string]string{"query": q})
	if err != nil {
		return fmt.Errorf("marshaling query: %w", err)
	}

	req, err := http.NewRequest("POST", railwayGraphQL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("invalid or expired token")
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned %d: %s", resp.StatusCode, string(b))
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response: %w", err)
	}

	var gql gqlResponse
	if err := json.Unmarshal(b, &gql); err == nil && len(gql.Errors) > 0 {
		return fmt.Errorf("railway API: %s", gql.Errors[0].Message)
	}

	return json.Unmarshal(b, out)
}
