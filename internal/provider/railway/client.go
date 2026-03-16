package railway

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const graphqlURL = "https://backboard.railway.app/graphql/v2"

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
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

type Service struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Project struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	Services []Service // flattened from edges
}

func (c *Client) GetUser() (*User, error) {
	var resp struct {
		Data struct {
			Me User `json:"me"`
		} `json:"data"`
	}
	err := c.query(`{ me { id email name } }`, &resp)
	if err != nil {
		return nil, err
	}
	return &resp.Data.Me, nil
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
								Node Service `json:"node"`
							} `json:"edges"`
						} `json:"services"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"projects"`
		} `json:"data"`
	}

	q := `{
		projects {
			edges {
				node {
					id
					name
					services {
						edges {
							node {
								id
								name
							}
						}
					}
				}
			}
		}
	}`

	if err := c.query(q, &resp); err != nil {
		return nil, err
	}

	var projects []Project
	for _, e := range resp.Data.Projects.Edges {
		p := Project{ID: e.Node.ID, Name: e.Node.Name}
		for _, se := range e.Node.Services.Edges {
			p.Services = append(p.Services, se.Node)
		}
		projects = append(projects, p)
	}
	return projects, nil
}

type Deployment struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type LogLine struct {
	Timestamp string `json:"timestamp"`
	Message   string `json:"message"`
	Severity  string `json:"severity"`
}

func (c *Client) GetActiveDeployment(serviceID string) (*Deployment, error) {
	var resp struct {
		Data struct {
			Deployments struct {
				Edges []struct {
					Node Deployment `json:"node"`
				} `json:"edges"`
			} `json:"deployments"`
		} `json:"data"`
	}

	q := `query($serviceId: String!) {
		deployments(
			input: { serviceId: $serviceId }
			first: 1
		) {
			edges {
				node {
					id
					status
				}
			}
		}
	}`

	vars := map[string]any{"serviceId": serviceID}
	if err := c.queryWithVars(q, vars, &resp); err != nil {
		return nil, err
	}

	if len(resp.Data.Deployments.Edges) == 0 {
		return nil, fmt.Errorf("no deployments found for service %s", serviceID)
	}
	d := resp.Data.Deployments.Edges[0].Node
	return &d, nil
}

func (c *Client) QueryLogs(deploymentID string, limit int) ([]LogLine, error) {
	var resp struct {
		Data struct {
			DeploymentLogs []LogLine `json:"deploymentLogs"`
		} `json:"data"`
	}

	q := `query($deploymentId: String!, $limit: Int) {
		deploymentLogs(deploymentId: $deploymentId, limit: $limit) {
			timestamp
			message
			severity
		}
	}`

	vars := map[string]any{
		"deploymentId": deploymentID,
		"limit":        limit,
	}
	if err := c.queryWithVars(q, vars, &resp); err != nil {
		return nil, err
	}
	return resp.Data.DeploymentLogs, nil
}

func (c *Client) queryWithVars(q string, vars map[string]any, out any) error {
	payload := map[string]any{"query": q, "variables": vars}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", graphqlURL, bytes.NewReader(body))
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

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned %d: %s", resp.StatusCode, string(b))
	}

	return json.NewDecoder(resp.Body).Decode(out)
}

func (c *Client) query(q string, out any) error {
	body, _ := json.Marshal(map[string]string{"query": q})

	req, err := http.NewRequest("POST", graphqlURL, bytes.NewReader(body))
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

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned %d: %s", resp.StatusCode, string(b))
	}

	return json.NewDecoder(resp.Body).Decode(out)
}
