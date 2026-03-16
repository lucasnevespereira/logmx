package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const railwayGraphQL = "https://backboard.railway.com/graphql/v2"

type RailwayClient struct {
	token      string
	httpClient *http.Client
}

func NewRailwayClient(token string) *RailwayClient {
	return &RailwayClient{token: token, httpClient: &http.Client{}}
}

type RailwayUser struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (c *RailwayClient) ValidateToken() (*RailwayUser, error) {
	var resp struct {
		Data struct {
			Me RailwayUser `json:"me"`
		} `json:"data"`
	}
	if err := c.query(`{ me { name email } }`, &resp); err != nil {
		return nil, err
	}
	if resp.Data.Me.Email == "" && resp.Data.Me.Name == "" {
		return nil, fmt.Errorf("invalid or expired token")
	}
	return &resp.Data.Me, nil
}

type RailwayProject struct {
	ID       string
	Name     string
	Services []RailwayService
}

type RailwayService struct {
	ID   string
	Name string
}

func (c *RailwayClient) ListProjects() ([]RailwayProject, error) {
	var resp struct {
		Data struct {
			Me struct {
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
			} `json:"me"`
		} `json:"data"`
	}

	q := `{ me { projects { edges { node { id name services { edges { node { id name } } } } } } } }`
	if err := c.query(q, &resp); err != nil {
		return nil, err
	}

	var projects []RailwayProject
	for _, e := range resp.Data.Me.Projects.Edges {
		p := RailwayProject{ID: e.Node.ID, Name: e.Node.Name}
		for _, se := range e.Node.Services.Edges {
			p.Services = append(p.Services, RailwayService{
				ID:   se.Node.ID,
				Name: se.Node.Name,
			})
		}
		projects = append(projects, p)
	}
	return projects, nil
}

func (c *RailwayClient) query(q string, out any) error {
	body, _ := json.Marshal(map[string]string{"query": q})

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
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("invalid or expired token")
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned %d: %s", resp.StatusCode, string(b))
	}

	return json.NewDecoder(resp.Body).Decode(out)
}
