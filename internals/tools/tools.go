package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Register wires all NBA tools onto the MCP server.
func Register(s *server.MCPServer, client *APIClient) {
	s.AddTool(searchPlayersTool(), searchPlayersHandler(client))
	s.AddTool(playerSeasonsTool(), playerSeasonsHandler(client))
	s.AddTool(playerGameLogsTool(), playerGameLogsHandler(client))
	s.AddTool(comparePlayersTool(), comparePlayersHandler(client))
	s.AddTool(vsTeamTool(), vsTeamHandler(client))
	s.AddTool(leadersTool(), leadersHandler(client))
	s.AddTool(defenseStatsTool(), defenseStatsHandler(client))
}

func searchPlayersTool() mcp.Tool {
	return mcp.NewTool("search_players",
		mcp.WithDescription("Search for NBA players by name. Returns player IDs needed for other tools."),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("Player name or partial name, e.g. 'jokic', 'lebron', 'michael jordan'"),
		),
	)
}

func searchPlayersHandler(client *APIClient) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		q := stringArg(req, "query", "")
		data, err := client.SearchPlayers(ctx, q)
		if err != nil {
			return toolError(err), nil
		}
		return toolResult(data), nil
	}
}

func playerSeasonsTool() mcp.Tool {
	return mcp.NewTool("get_player_seasons",
		mcp.WithDescription("Get all season averages for a player. Includes traditional and advanced stats (PER, BPM, VORP, WS, TS%)."),
		mcp.WithString("player_id",
			mcp.Required(),
			mcp.Description("Basketball Reference player ID, e.g. 'jokicni01'. ALWAYS USE search_players first to resolve player ID."),
		),
		mcp.WithString("season_type",
			mcp.Description("'regular' or 'playoffs'. Defaults to 'regular'."),
			mcp.Enum("regular", "playoffs"),
		),
	)
}

func playerSeasonsHandler(client *APIClient) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		playerID := stringArg(req, "player_id", "")
		seasonType := stringArg(req, "season_type", "regular")

		data, err := client.PlayerSeasons(ctx, playerID, seasonType)
		if err != nil {
			return toolError(err), nil
		}
		return toolResult(data), nil
	}
}

func playerGameLogsTool() mcp.Tool {
	return mcp.NewTool("get_player_gamelogs",
		mcp.WithDescription(`Get game-by-game logs for a player. Supports filtering by season, FG%, and minimum field goal attempts.
Use this for questions like:
- "How many games did X shoot under 40%?"
- "Show me all games where X scored 30+"
- "How did X perform in January 2024?"`),
		mcp.WithString("player_id",
			mcp.Required(),
			mcp.Description("Basketball Reference player ID. ALWAYS USE search_players first to resolve player ID."),
		),
		mcp.WithString("season",
			mcp.Description("Season as a 4-digit end year, e.g. '2022-23, 22/23, 2022-2023' for '2023' season. Omit for all seasons."),
		),
		mcp.WithString("season_type",
			mcp.Description("'regular' or 'playoffs'. Defaults to 'regular'."),
			mcp.Enum("regular", "playoffs"),
		),
		mcp.WithString("fg_pct_lt",
			mcp.Description("Filter games where FG% was LESS THAN this value. Decimal between 0 and 1, e.g. '0.40'."),
		),
		mcp.WithString("min_fga",
			mcp.Description("Minimum field goal attempts to include a game. Filters out DNP/low-attempt games."),
		),
	)
}

func playerGameLogsHandler(client *APIClient) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		playerID := stringArg(req, "player_id", "")
		season := stringArg(req, "season", "")
		seasonType := stringArg(req, "season_type", "regular")
		fgPctLt := stringArg(req, "fg_pct_lt", "")
		minFGA := stringArg(req, "min_fga", "")

		data, err := client.PlayerGameLogs(ctx, playerID, season, seasonType, fgPctLt, minFGA)
		if err != nil {
			return toolError(err), nil
		}
		return toolResult(data), nil
	}
}

func vsTeamTool() mcp.Tool {
	return mcp.NewTool("getplayerstatsvsteam",
		mcp.WithDescription(`Get a player's game logs or season averages against a specific TEAM (not a player).
Use this ONLY when the opponent is a franchise/team abbreviation like GSW, LAL, BOS, MIA.
Do NOT use this for player vs player comparisons — use compare_players instead.
If season is provided, returns individual game logs for that season.
If season is omitted, returns per-season averages for all time against that team.
ALWAYS USE search_players first to resolve player ID.
Examples:
- "How does Jokić perform against the Lakers?" → use this tool with team=LAL
- "Jokić vs LeBron head to head" → use compare_players instead`),
		mcp.WithString("playerid",
			mcp.Required(),
			mcp.Description("Player ID, e.g. jokicni01. ALWAYS USE search_players first to resolve player IDs"),
		),
		mcp.WithString("team",
			mcp.Required(),
			mcp.Description("Opponent TEAM abbreviation only, e.g. GSW, LAL, BOS. Not a player name."),
		),
		mcp.WithString("season",
			mcp.Description("Season as a 4-digit end year. e.g. '2022-23, 22/23, 2022-2023' for '2023'  Omit for all-time averages."),
		),
		mcp.WithString("seasontype",
			mcp.Description("regular or playoffs, default regular"),
			mcp.Enum("regular", "playoffs"),
		),
	)
}

func vsTeamHandler(client *APIClient) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		playerID := stringArg(req, "playerid", "")
		team := stringArg(req, "team", "")
		season := stringArg(req, "season", "")
		seasonType := stringArg(req, "seasontype", "regular")

		if playerID == "" {
			return toolError(fmt.Errorf("playerid is required")), nil
		}
		if team == "" {
			return toolError(fmt.Errorf("team is required")), nil
		}

		var (
			data json.RawMessage
			err  error
		)
		if season != "" {
			data, err = client.VsTeam(ctx, playerID, team, season, seasonType)
		} else {
			data, err = client.VsTeamAllTime(ctx, playerID, team, seasonType)
		}
		if err != nil {
			return toolError(err), nil
		}
		return toolResult(data), nil
	}
}

func comparePlayersTool() mcp.Tool {
	return mcp.NewTool("compare_players",
		mcp.WithDescription(`Compare two or more NBA players head-to-head. Use this for ANY query involving one player versus another player.
Returns season stats side-by-side plus a head-to-head win/loss record when exactly 2 players are compared.
ALWAYS USE search_players first to resolve player IDs.
Use this for questions like:
- "Jokić vs Embiid head to head"
- "Anthony Edwards vs Jokić in the 2024-25 season"
- "How many games has X won against Y?"
- "Compare LeBron and Curry stats"
- "Who has the better record between X and Y?"
Do NOT use getplayerstatsvsteam for player vs player — that tool is for player vs team only.`),
		mcp.WithString("player_ids",
			mcp.Required(),
			mcp.Description("Comma-separated Basketball Reference player IDs, e.g. 'jokicni01,embiijo01'. ALWAYS USE search_players first to resolve player IDs."),
		),
		mcp.WithString("seasons",
			mcp.Description("Comma-separated seasons in YYYY-YY format e.g. '2023-24' or '2022-23,2023-24'. It can also be a range of seasons like '2022-2026 which will translate into 2021-22, 2022-23, 2023-24, 2024-25 and 2025-26' Omit for all-time."),
		),
		mcp.WithString("season_type",
			mcp.Description("'regular' or 'playoffs'. Defaults to 'regular'."),
			mcp.Enum("regular", "playoffs"),
		),
	)
}

func comparePlayersHandler(client *APIClient) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		idsRaw := stringArg(req, "player_ids", "")
		seasonsRaw := stringArg(req, "seasons", "")
		seasonType := stringArg(req, "season_type", "regular")

		if idsRaw == "" {
			return toolError(fmt.Errorf("player_ids is required")), nil
		}

		ids := splitTrim(idsRaw)
		if len(ids) < 2 {
			return toolError(fmt.Errorf("provide at least 2 comma-separated player IDs")), nil
		}
		if len(ids) > 5 {
			return toolError(fmt.Errorf("maximum 5 players can be compared at once")), nil
		}

		var seasons []string
		if seasonsRaw != "" {
			seasons = splitTrim(seasonsRaw)
		}

		data, err := client.ComparePlayers(ctx, ids, seasons, seasonType)
		if err != nil {
			return toolError(err), nil
		}
		return toolResult(data), nil
	}
}

func leadersTool() mcp.Tool {
	return mcp.NewTool("get_leaders",
		mcp.WithDescription("Get the league leaders for any stat category in a given season."),
		mcp.WithString("stat",
			mcp.Required(),
			mcp.Description("Stat column to rank by."),
			mcp.Enum("pts", "reb", "ast", "stl", "blk", "fg_pct", "fg3_pct", "ts_pct", "per", "bpm", "vorp", "ws", "usg_pct"),
		),
		mcp.WithString("season",
			mcp.Description("Season as a 4-digit end year, e.g. '2022-23, 22/23, 2022-2023' for '2023' season. Omit for all-time."),
		),
		mcp.WithString("season_type",
			mcp.Description("'regular' or 'playoffs'. Defaults to 'regular'."),
			mcp.Enum("regular", "playoffs"),
		),
		mcp.WithString("limit",
			mcp.Description("Number of results to return. Max 50. Defaults to 10."),
		),
	)
}

func leadersHandler(client *APIClient) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		stat := stringArg(req, "stat", "")
		season := stringArg(req, "season", "")
		seasonType := stringArg(req, "season_type", "regular")
		limit := stringArg(req, "limit", "10")

		data, err := client.Leaders(ctx, stat, season, seasonType, limit)
		if err != nil {
			return toolError(err), nil
		}
		return toolResult(data), nil
	}
}

func defenseStatsTool() mcp.Tool {
	return mcp.NewTool("get_player_defense_stats",
		mcp.WithDescription(
			"Get per-game defensive stats for a player — opponent FGA and FGM per game when this player "+
				"is the closest defender, plus opponent FG% and DFG% diff vs league average. "+
				"def_fga and def_fgm are season averages (e.g. 4.2 FGA allowed per game), NOT single-game counts. "+
				"def_fgpct, dfg_diff, and normal_fgpct are expressed as percentages (e.g. 40.7, -8.6, 49.3). "+
				"dfg_diff is opponent FG% minus league average in percentage points — negative = better defender. "+
				"Also use this for questions like: how are players shooting when X is the primary defender?"+
				"Omit season to get the full career defensive history. "+
				"ALWAYS USE search_players first to resolve the player_id.",
		),
		mcp.WithString("player_id",
			mcp.Required(),
			mcp.Description("Basketball Reference player ID, e.g. 'jokicni01'. ALWAYS USE search_players first."),
		),
		mcp.WithString("season",
			mcp.Description("Season as a 4-digit end year, e.g. '2022-23, 22/23, 2022-2023' for '2023' season. Omit for all-time."),
		),
		mcp.WithString("season_type",
			mcp.Description("'regular' or 'playoffs'. Defaults to 'regular'."),
			mcp.Enum("regular", "playoffs"),
		),
	)
}

func defenseStatsHandler(client *APIClient) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		playerID := stringArg(req, "player_id", "")
		season := stringArg(req, "season", "")
		seasonType := stringArg(req, "season_type", "regular")

		if playerID == "" {
			return toolError(fmt.Errorf("player_id is required")), nil
		}

		data, err := client.DefenseStats(ctx, playerID, season, seasonType)
		if err != nil {
			return toolError(err), nil
		}
		return toolResult(data), nil
	}
}

func args(req mcp.CallToolRequest) map[string]any {
	if req.Params.Arguments == nil {
		return map[string]any{}
	}
	m, ok := req.Params.Arguments.(map[string]any)
	if !ok {
		return map[string]any{}
	}
	return m
}

func stringArg(req mcp.CallToolRequest, key, fallback string) string {
	m := args(req)
	v, ok := m[key]
	if !ok || v == nil {
		return fallback
	}
	s, ok := v.(string)
	if !ok || s == "" {
		return fallback
	}
	return s
}

func toolResult(data json.RawMessage) *mcp.CallToolResult {
	return mcp.NewToolResultText(string(data))
}

func toolError(err error) *mcp.CallToolResult {
	return mcp.NewToolResultError(err.Error())
}

func splitTrim(s string) []string {
	parts := strings.Split(s, ",")
	for i, p := range parts {
		parts[i] = strings.TrimSpace(p)
	}
	return parts
}
