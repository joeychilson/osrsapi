package osrsapi

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient() *Client {
	return &Client{
		baseURL:    "https://secure.runescape.com",
		httpClient: &http.Client{},
	}
}

type GameMode string

const (
	ModeRegular         GameMode = ""
	ModeIronman         GameMode = "_ironman"
	ModeHardcoreIronman GameMode = "_hardcore_ironman"
	ModeUltimateIronman GameMode = "_ultimate"
	ModeDeadman         GameMode = "_deadman"
	ModeSeasonal        GameMode = "_seasonal"
	ModeTournament      GameMode = "_tournament"
	ModeFreshStart      GameMode = "_fresh_start"
)

type ResponseFormat string

const (
	FormatCSV  ResponseFormat = "csv"
	FormatJSON ResponseFormat = "json"
)

type Skill struct {
	Name       string `json:"name"`
	Rank       int    `json:"rank"`
	Level      int    `json:"level"`
	Experience int    `json:"experience"`
}

type Activity struct {
	Name  string `json:"name"`
	Rank  int    `json:"rank"`
	Score int    `json:"score"`
}

type PlayerStats struct {
	Skills     []Skill    `json:"skills"`
	Activities []Activity `json:"activities"`
}

func (c *Client) Hiscores(ctx context.Context, username string, mode GameMode, format ResponseFormat) (*PlayerStats, error) {
	if username == "" {
		return nil, fmt.Errorf("username is required")
	}
	if mode == "" {
		mode = ModeRegular
	}
	if format == "" {
		format = FormatJSON
	}

	var path string
	if format == FormatJSON {
		path = fmt.Sprintf("m=hiscore_oldschool%s/index_lite.json?player=%s", mode, username)
	} else {
		path = fmt.Sprintf("m=hiscore_oldschool%s/index_lite.ws?player=%s", mode, username)
	}

	data, err := c.sendRequest(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}

	switch format {
	case FormatJSON:
		var stats PlayerStats
		if err := json.Unmarshal(data, &stats); err != nil {
			return nil, fmt.Errorf("decoding JSON response: %w", err)
		}
		return &stats, nil
	case FormatCSV:
		reader := csv.NewReader(strings.NewReader(string(data)))
		reader.FieldsPerRecord = 3

		var (
			stats     PlayerStats
			lineCount int
		)

		for {
			record, err := reader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, fmt.Errorf("reading CSV: %w", err)
			}

			rank, err := strconv.Atoi(record[0])
			if err != nil {
				return nil, fmt.Errorf("converting rank to int: %w", err)
			}
			levelOrScore, err := strconv.Atoi(record[1])
			if err != nil {
				return nil, fmt.Errorf("converting level or score to int: %w", err)
			}
			expOrEmpty, err := strconv.Atoi(record[2])
			if err != nil {
				return nil, fmt.Errorf("converting experience to int: %w", err)
			}

			if lineCount < 24 {
				stats.Skills = append(stats.Skills, Skill{
					Name:       getSkillName(lineCount),
					Rank:       rank,
					Level:      levelOrScore,
					Experience: expOrEmpty,
				})
			} else {
				stats.Activities = append(stats.Activities, Activity{
					Name:  getActivityName(lineCount - 24),
					Rank:  rank,
					Score: levelOrScore,
				})
			}
			lineCount++
		}
		return &stats, nil
	default:
		return nil, fmt.Errorf("invalid format: %s", format)
	}
}

func getSkillName(index int) string {
	skills := []string{
		"Overall", "Attack", "Defence", "Strength", "Hitpoints", "Ranged", "Prayer", "Magic",
		"Cooking", "Woodcutting", "Fletching", "Fishing", "Firemaking", "Crafting", "Smithing",
		"Mining", "Herblore", "Agility", "Thieving", "Slayer", "Farming", "Runecrafting",
		"Hunter", "Construction",
	}
	return skills[index]
}

func getActivityName(index int) string {
	activities := []string{
		"League Points", "Deadman Points", "Bounty Hunter - Hunter", "Bounty Hunter - Rogue",
		"Bounty Hunter (Legacy) - Hunter", "Bounty Hunter (Legacy) - Rogue", "Clue Scrolls (all)",
		"Clue Scrolls (beginner)", "Clue Scrolls (easy)", "Clue Scrolls (medium)", "Clue Scrolls (hard)",
		"Clue Scrolls (elite)", "Clue Scrolls (master)", "LMS - Rank", "PvP Arena - Rank", "Soul Wars Zeal",
		"Rifts closed", "Colosseum Glory", "Abyssal Sire", "Alchemical Hydra", "Artio", "Barrows Chests",
		"Bryophyta", "Callisto", "Cal'varion", "Cerberus", "Chambers of Xeric",
		"Chambers of Xeric: Challenge Mode", "Chaos Elemental", "Chaos Fanatic", "Commander Zilyana",
		"Corporeal Beast", "Crazy Archaeologist", "Dagannoth Prime", "Dagannoth Rex", "Dagannoth Supreme",
		"Deranged Archaeologist", "Duke Sucellus", "General Graardor", "Giant Mole", "Grotesque Guardians",
		"Hespori", "Kalphite Queen", "King Black Dragon", "Kraken", "Kree'Arra", "K'ril Tsutsaroth",
		"Lunar Chests", "Mimic", "Nex", "Nightmare", "Phosani's Nightmare", "Obor", "Phantom Muspah",
		"Sarachnis", "Scorpia", "Scurrius", "Skotizo", "Sol Heredit", "Spindel", "Tempoross",
		"The Gauntlet", "The Corrupted Gauntlet", "The Leviathan", "The Whisperer", "Theatre of Blood",
		"Theatre of Blood: Hard Mode", "Thermonuclear Smoke Devil", "Tombs of Amascut",
		"Tombs of Amascut: Expert Mode", "TzKal-Zuk", "TzTok-Jad", "Vardorvis", "Venenatis", "Vet'ion",
		"Vorkath", "Wintertodt", "Zalcano", "Zulrah",
	}
	return activities[index]
}

type PriceTrend struct {
	Price Price  `json:"price"`
	Trend string `json:"trend"`
}

type Price int

func (p *Price) UnmarshalJSON(data []byte) error {
	var intPrice int
	if err := json.Unmarshal(data, &intPrice); err == nil {
		*p = Price(intPrice)
		return nil
	}

	var strPrice string
	if err := json.Unmarshal(data, &strPrice); err != nil {
		return err
	}

	strPrice = strings.ReplaceAll(strPrice, " ", "")
	strPrice = strings.ReplaceAll(strPrice, ",", "")
	strPrice = strings.ToLower(strPrice)

	var multiplier float64 = 1
	switch {
	case strings.HasSuffix(strPrice, "k"):
		multiplier = 1000
		strPrice = strings.TrimSuffix(strPrice, "k")
	case strings.HasSuffix(strPrice, "m"):
		multiplier = 1000000
		strPrice = strings.TrimSuffix(strPrice, "m")
	case strings.HasSuffix(strPrice, "b"):
		multiplier = 1000000000
		strPrice = strings.TrimSuffix(strPrice, "b")
	}

	price, err := strconv.ParseFloat(strPrice, 64)
	if err != nil {
		return fmt.Errorf("failed to parse price: %w", err)
	}

	*p = Price(int(price * multiplier))
	return nil
}

type Items struct {
	Total int `json:"total"`
	Items []struct {
		ID          int        `json:"id"`
		Icon        string     `json:"icon"`
		IconLarge   string     `json:"icon_large"`
		Type        string     `json:"type"`
		TypeIcon    string     `json:"typeIcon"`
		Name        string     `json:"name"`
		Description string     `json:"description"`
		Members     string     `json:"members"`
		Current     PriceTrend `json:"current"`
		Today       PriceTrend `json:"today"`
	} `json:"items"`
}

func (c *Client) Items(ctx context.Context, alpha string, page int) (*Items, error) {
	path := fmt.Sprintf("m=itemdb_oldschool/api/catalogue/items.json?category=1&alpha=%s&page=%d", alpha, page)

	data, err := c.sendRequest(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}

	var resp Items
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("decoding JSON response: %w", err)
	}
	return &resp, nil
}

type Item struct {
	ID          int         `json:"id"`
	Icon        string      `json:"icon"`
	IconLarge   string      `json:"icon_large"`
	Type        string      `json:"type"`
	TypeIcon    string      `json:"typeIcon"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Members     string      `json:"members"`
	Current     PriceTrend  `json:"current"`
	Today       PriceTrend  `json:"today"`
	Day30       ChangeTrend `json:"day30"`
	Day90       ChangeTrend `json:"day90"`
	Day180      ChangeTrend `json:"day180"`
}

type ChangeTrend struct {
	Change PercentChange `json:"change"`
	Trend  string        `json:"trend"`
}

type PercentChange float64

func (pc *PercentChange) UnmarshalJSON(data []byte) error {
	var strChange string
	if err := json.Unmarshal(data, &strChange); err != nil {
		return err
	}
	strChange = strings.TrimSpace(strings.TrimSuffix(strChange, "%"))

	change, err := strconv.ParseFloat(strChange, 64)
	if err != nil {
		return fmt.Errorf("failed to parse percentage change: %w", err)
	}

	*pc = PercentChange(change)
	return nil
}

func (pc PercentChange) String() string {
	return fmt.Sprintf("%.2f%%", pc)
}

func (c *Client) Item(ctx context.Context, itemID int) (*Item, error) {
	path := fmt.Sprintf("m=itemdb_oldschool/api/catalogue/detail.json?item=%d", itemID)

	data, err := c.sendRequest(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}

	var resp struct {
		Item Item `json:"item"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("decoding JSON response: %w", err)
	}
	return &resp.Item, nil
}

type GraphData struct {
	Daily   map[string]int `json:"daily"`
	Average map[string]int `json:"average"`
}

func (c *Client) ItemGraph(ctx context.Context, itemID int) (*GraphData, error) {
	path := fmt.Sprintf("m=itemdb_oldschool/api/graph/%d.json", itemID)

	data, err := c.sendRequest(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}

	var graph GraphData
	if err := json.Unmarshal(data, &graph); err != nil {
		return nil, fmt.Errorf("decoding JSON response: %w", err)
	}
	return &graph, nil
}

func (c *Client) sendRequest(ctx context.Context, path string) ([]byte, error) {
	url := fmt.Sprintf("%s/%s", c.baseURL, path)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	return body, nil
}
