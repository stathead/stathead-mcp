package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type APIClient struct {
	base   string
	client *http.Client
}

func NewAPIClient(base string) *APIClient {
	return &APIClient{
		base:   strings.TrimRight(base, "/"),
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *APIClient) get(ctx context.Context, path string, params url.Values) (json.RawMessage, error) {
	u := c.base + path
	if len(params) > 0 {
		u += "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var e map[string]string
		json.NewDecoder(resp.Body).Decode(&e)
		return nil, fmt.Errorf("api error %d: %s", resp.StatusCode, e["error"])
	}

	var raw json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}
	return raw, nil
}

func (c *APIClient) SearchPlayers(ctx context.Context, q string) (json.RawMessage, error) {
	return c.get(ctx, "/players/search", url.Values{"q": {q}})
}

func (c *APIClient) PlayerSeasons(ctx context.Context, playerID, seasonType string) (json.RawMessage, error) {
	return c.get(ctx, "/players/"+playerID+"/seasons", url.Values{
		"season_type": {seasonType},
	})
}

func (c *APIClient) PlayerGameLogs(ctx context.Context, playerID, season, seasonType, fgPctLt, minFGA string) (json.RawMessage, error) {
	p := url.Values{
		"season_type": {seasonType},
	}
	if season != "" {
		p.Set("season", season)
	}
	if fgPctLt != "" {
		p.Set("fg_pct_lt", fgPctLt)
	}
	if minFGA != "" {
		p.Set("min_fga", minFGA)
	}
	return c.get(ctx, "/players/"+playerID+"/gamelogs", p)
}

func (c *APIClient) VsTeam(ctx context.Context, playerID, team, season, seasonType string) (json.RawMessage, error) {
	p := url.Values{
		"season_type": {seasonType},
		"team":        {team},
	}
	if season != "" {
		p.Set("season", season)
	}

	return c.get(ctx, "/players/"+playerID+"/vs-team", p)
}

func (c *APIClient) VsTeamAllTime(ctx context.Context, playerID, team, seasonType string) (json.RawMessage, error) {
	p := url.Values{
		"season_type": {seasonType},
		"team":        {team},
	}

	return c.get(ctx, "/players/"+playerID+"/vs-team/all-time", p)
}

func (c *APIClient) ComparePlayers(ctx context.Context, playerIDs, seasons []string, seasonType string) (json.RawMessage, error) {
	p := url.Values{"season_type": {seasonType}}
	for _, id := range playerIDs {
		p.Add("player_id", id)
	}
	for _, s := range seasons {
		p.Add("season", s)
	}
	return c.get(ctx, "/players/compare", p)
}

func (c *APIClient) Leaders(ctx context.Context, stat, season, seasonType, limit string) (json.RawMessage, error) {
	p := url.Values{
		"stat":        {stat},
		"season_type": {seasonType},
	}
	if season != "" {
		p.Set("season", season)
	}
	if limit != "" {
		p.Set("limit", limit)
	}
	return c.get(ctx, "/players/leaders", p)
}
