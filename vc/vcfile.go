package vc

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"time"
)

// FilePath path to the main VC file
var FilePath string

// Data Main data file
var Data *VFile

// MasterDataStr master data as read from the file as a string.
var MasterDataStr string

// LangPack language pack to use
var LangPack string

// Timestamp in the JSON file
type Timestamp struct {
	time.Time
}

//BinImage image information from a .BIN file
type BinImage struct {
	ID   int
	Name string
	Data []byte
}

// ReadMasterData Reads the master data from the file location
func ReadMasterData(file string) error {
	if Data == nil {
		Data = &(VFile{})
	}
	b, err := Read(file)
	if err != nil {
		os.Stderr.WriteString(err.Error() + "\n")
		return err
	}
	MasterDataStr = string(b)
	return nil
}

// MarshalJSON converts a JSON timestamp to a GO time
func (t *Timestamp) MarshalJSON() ([]byte, error) {
	if t.Time.IsZero() {
		return []byte("-1"), nil
	}

	ts := t.Time.Unix()
	stamp := fmt.Sprint(ts)

	return []byte(stamp), nil
}

var location = time.FixedZone("JST", 32400) //time.LoadLocation("Asia/Tokyo")

// old cards that are not available anymore
var retiredCards = []int{
	61, 62, // bandit
	55, 56, // beastmaster
	102,      // cutthroat
	323, 324, // cyborg
	121, 122, // dancer
	38, 39, // dark knight
	123, 124, // detective
	163, 164, // doll master
	41, 42, // dragon knight
	43,       // dragon slayer
	225, 226, // dragonewt
	119, 120, // druid
	174, 175, // empress
	157, 158, // farmer
	201, 202, // fox spirit
	229, 230, // gnome
	242, 243, // harpy
	86, 87, // hunter
	63,            // idol
	1, 2, 3, 4, 5, // knight
	90, 91, // kung-fu master
	195, 196, // lycaon
	88, 89, // martial artist
	321, 322, // mechanic
	348, 349, // mythic knight
	265, 266, // oni
	40,     // paladin
	94, 95, // rune knight
	73, 74, // sage
	84, 85, // strategist
	100, 101, // swordsman
	183, 184, // sylph
	304, 305, // trickster
	133, 134, // vampire hunter
}

// new cards that are named the same as an old card that is still active
var newCards = []int{
	4721, 4722, 4734, // Spinner
	4752, 4753, 4785, // Sparky
	4711, 4712, // Diana
}

// ID is the character ID, not the card ID
var characterNameOverride = map[int]string{
	62:   "Kung-Fu Master",            // Kung Fu Master
	181:  "Ariel (Light)",             // Ariel
	352:  "Ariel (Dark)",              // Ariel
	452:  "Joker",                     // Joker
	465:  "Joker (Cane)",              // Joker
	466:  "Joker (Sickle)",            // Joker
	495:  "Snowman MK II",             // Snowman MKⅡ
	1319: "Jack-O'-Sisters",           // Jack-o'-Sisters (older card with newer name format)
	1173: "Al-mi'raj",                 // Al-Mi'Raj
	1536: "Valiant Bellona (Bronze)",  // Valiant Bellona
	1537: "Valiant Bellona (Silver)",  // Valiant Bellona
	1538: "Valiant Bellona (Gold)",    // Valiant Bellona
	1816: "Gold Girl (SR)",            // Gold Girl
	1846: "Medal Girl (SR)",           // Medal Girl
	1869: "Super Chimry (Passion)",    // Super Chimry
	1870: "Super Chimry (Cool)",       // Super Chimry
	1871: "Super Chimry (Light)",      // Super Chimry
	1872: "Super Chimry (Dark)",       // Super Chimry
	1874: "Hyper Chimry (Passion)",    // Hyper Chimry
	1875: "Hyper Chimry (Cool)",       // Hyper Chimry
	1876: "Hyper Chimry (Light)",      // Hyper Chimry
	1877: "Hyper Chimry (Dark)",       // Hyper Chimry
	2024: "Playful Hades (Red)",       // Playful Hades
	2025: "Playful Hades (Green)",     // Playful Hades
	2026: "Playful Hades (Blue)",      // Playful Hades
	2397: "PM Demise",                 // Pm Demise
	2479: "DIY Ninja",                 // Diy Ninja
	2549: "Thunder Stone Shard (L)",   // Thunderstone Shard (L)
	2550: "Thunder Stone Shard (D)",   // Thunderstone Shard (D)
	2554: "Lightning Stone Shard (L)", // Lightning Shard (L)
	2555: "Lightning Stone Shard (D)", // Lightning Shard (D)
}

func init() {
	sort.Ints(retiredCards)
	sort.Ints(newCards)
}

// UnmarshalJSON converts a GO time to a JSON timestamp
func (t *Timestamp) UnmarshalJSON(b []byte) error {
	ts, err := strconv.Atoi(string(b))
	if err != nil {
		return err
	}

	if ts == -1 {
		t.Time = time.Time{}
	} else {
		if location != nil {
			t.Time = time.Unix(int64(ts), 0).In(location)
		} else {
			t.Time = time.Unix(int64(ts), 0)
		}
	}

	return nil
}

// VFile Main Structure for the VC data file located in responce/maindata
type VFile struct {
	Code   int `json:"code"`
	Common struct {
		UnixTime Timestamp `json:"unixtime"`
	} `json:"common"`
	Defs []struct {
		ID    int `json:"_id"`
		Value int `json:"value"`
	} `json:"defs"`
	DefsTune []struct {
		ID            int       `json:"_id"`
		MstDefsID     int       `json:"mst_defs_id"`
		Value         int       `json:"value"`
		PublicFlg     int       `json:"public_flg"`
		StartDateTime Timestamp `json:"start_datetime"`
		EndDateTime   Timestamp `json:"end_datetime"`
	} `json:"defs_tune"`
	ShortcutURL                 string                      `json:"shortcut_url"`
	Version                     int                         `json:"version"`
	Cards                       CardList                    `json:"cards"`
	Skills                      []Skill                     `json:"skills"`
	SkillLevels                 []SkillLevel                `json:"skill_level"`
	CustomSkillLevels           []CustomSkillLevel          `json:"custom_skill_level"`
	SkillCostIncrementPatterns  []SkillCostIncrementPattern `json:"skill_cost_increment_pattern"`
	Amalgamations               []Amalgamation              `json:"fusion_list"`
	Awakenings                  []CardAwaken                `json:"card_awaken"`
	Rebirths                    []CardAwaken                `json:"card_super_awaken"`
	CardCharacters              []CardCharacter             `json:"card_character"`
	FollowerKinds               []FollowerKind              `json:"follower_kinds"`
	CardRarities                []CardRarity                `json:"card_rares"`
	CardSpecialComposes         []CardSpecialCompose        `json:"card_special_compose"`
	Levels                      []Level                     `json:"levels"`
	LevelupBonuses              []LevelupBonus              `json:"levelup_bonus"`
	CardLevels                  []CardLevel                 `json:"cardlevel"`
	CardLevelsLR                []CardLevel                 `json:"cardlevel_lr"`
	CardLevelsX                 []CardLevel                 `json:"cardlevel_x"`
	CardLevelsXLR               []CardLevel                 `json:"cardlevel_xlr"`
	LevelLRResources            []LevelResource             `json:"card_compose_resource"`
	LevelXResources             []LevelResource             `json:"card_compose_resource_x"`
	LevelXLRResources           []LevelResource             `json:"card_compose_resource_xlr"`
	DeckBonuses                 []DeckBonus                 `json:"deck_bonus"`
	DeckBonusConditions         []DeckBonusCond             `json:"deck_bonus_cond"`
	Archwitches                 ArchwitchList               `json:"kings"`
	ArchwitchSeries             []ArchwitchSeries           `json:"king_series"`
	ArchwitchFriendships        []ArchwitchFriendship       `json:"king_friendship"`
	Events                      []Event                     `json:"mst_event"`
	EventBooks                  []EventBook                 `json:"mst_event_book"`
	EventCards                  []EventCard                 `json:"mst_event_card"`
	RankRewards                 []RankReward                `json:"ranking_bonus"`
	RankRewardSheets            []RankRewardSheet           `json:"ranking_bonussheet"`
	Maps                        []Map                       `json:"map"`
	Areas                       []Area                      `json:"area"`
	Items                       []Item                      `json:"items"`
	Structures                  []Structure                 `json:"structures"`
	StructureLevels             []StructureLevel            `json:"structure_level"`
	StructureNumCosts           []StructureCost             `json:"structure_num_cost"`
	ResourceLevels              []ResourceLevel             `json:"resource"`
	BankLevels                  []BankLevel                 `json:"bank_level"`
	CastleLevels                []CastleLevel               `json:"castle_level"`
	ThorEvents                  []ThorEvent                 `json:"mst_thorhammer"`
	ThorKings                   []ThorKing                  `json:"mst_thorhammer_king"`
	ThorKingCosts               []ThorKingCost              `json:"mst_thorhammer_king_cost"`
	ThorRankRewards             []ThorReward                `json:"mst_thorhammer_ranking_reward"`
	ThorPointRewards            []ThorReward                `json:"mst_thorhammer_point_reward"`
	GuildBattles                []GuildBattle               `json:"mst_guildbattle_schedule"`
	GuildBingoBattles           []GuildBingoBattle          `json:"mst_guildbingo"`
	GuildBingoExchangeRewards   []GuildBingoExchangeReward  `json:"mst_guildbingo_exchange_reward"`
	GuildBingoPointCampaigns    []GuildBingoPointCampaign   `json:"mst_guildbingo_point_campaign"`
	GuildBattleRewardRefs       []GuildBattleRewardRef      `json:"mst_guildbattle_point_reward"`
	GuildBattleIndividualPoints []RankRewardSheet           `json:"mst_guildbattle_point_rewardsheet"`
	GuildBattleRankingRewards   []RankRewardSheet           `json:"mst_guildbattle_individual_ranking_reward"`
	GuildAUBWinRewards          []GuildAUBWinReward         `json:"mst_guildbattle_win_reward"`
	Towers                      []Tower                     `json:"mst_tower"`
	TowerRewards                []RankRewardSheet           `json:"mst_tower_ranking_reward"`
	TowerArrivalRewards         []RankRewardSheet           `json:"mst_tower_arrival_point_reward"`
	Dungeons                    []Dungeon                   `json:"mst_dungeon"`
	DungeonAreaTypes            []DungeonAreaType           `json:"mst_dungeon_area_type"`
	DungeonRewards              []RankRewardSheet           `json:"mst_dungeon_ranking_reward"`
	DungeonArrivalRewards       []RankRewardSheet           `json:"mst_dungeon_arrival_point_reward"`
}

// Read This reads the main data file and all associated files for strings
// the data is inserted directly into the struct.
func Read(root string) ([]byte, error) {
	filename := root + "/response/master_all"

	var data []byte
	var err error
	var jsonFileInfo os.FileInfo
	if jsonFileInfo, err = os.Stat(filename + ".json"); os.IsNotExist(err) {
		_, data, err = DecodeAndSave(filename)
		if err != nil {
			return nil, errors.New("no such file or directory: " + filename)
		}
	} else {
		md, err := os.Stat(filename)
		if err != nil {
			return nil, err
		}
		// check the timestamp on the saved file and verify the master data has not been updated
		if jsonFileInfo.ModTime().Unix() >= md.ModTime().Unix() {
			data, err = ioutil.ReadFile(filename + ".json")
			if err != nil {
				return nil, err
			}
		} else {
			_, data, err = DecodeAndSave(filename)
			if err != nil {
				return nil, errors.New("no such file or directory: " + filename)
			}
		}
	}

	// decode the main file
	err = json.Unmarshal(data[:], Data)
	if err != nil {
		debug.PrintStack()
		return nil, err
	}

	// card names
	names, err := ReadStringFile(root + "/string/MsgCardName_en.strb")
	if err != nil {
		debug.PrintStack()
		return nil, err
	}
	if len(Data.Cards) > len(names) {
		fmt.Fprintf(os.Stdout, "names: %v\n", names)
		debug.PrintStack()
		return nil, fmt.Errorf("%s did not match data file. master: %d, strings: %d",
			"Character Names", len(Data.Cards), len(names))
	}
	for key := range Data.Cards {
		Data.Cards[key].Name = cleanCardName(names[key], Data.Cards[key])
	}
	// initialize the evolutions
	for key := range Data.Cards {
		card := Data.Cards[key]
		// the name 'Goddess Crystal Shard' is reused, so we use a naming convention for it.
		if card.Name == "Goddess Crystal Shard" {
			for _, a := range card.Amalgamations() {
				if a.FusionCardID != card.ID { // this is the material card
					rCard := CardScan(a.FusionCardID)
					card.Name += " (" + rCard.Name + ")"
				}
			}
		}
		card.GetEvolutions()
	}

	description, err := ReadStringFile(root + "/string/MsgCharaDesc_en.strb")
	if err != nil {
		debug.PrintStack()
		return nil, err
	}
	if len(Data.CardCharacters) > len(description) {
		debug.PrintStack()
		return nil, fmt.Errorf("%s did not match data file. master: %d, strings: %d",
			"Character descriptions", len(Data.CardCharacters), len(description))
	}

	friendship, err := ReadStringFile(root + "/string/MsgCharaFriendship_en.strb")
	if err != nil {
		debug.PrintStack()
		return nil, err
	}
	if len(Data.CardCharacters) > len(friendship) {
		debug.PrintStack()
		return nil, fmt.Errorf("%s did not match data file. master: %d, strings: %d",
			"Character friendship", len(Data.CardCharacters), len(friendship))
	}

	login, err := ReadStringFile(root + "/string/MsgCharaWelcome_en.strb")
	if err != nil {
		debug.PrintStack()
		return nil, err
	}

	meet, err := ReadStringFile(root + "/string/MsgCharaMeet_en.strb")
	if err != nil {
		debug.PrintStack()
		return nil, err
	}
	if len(Data.CardCharacters) > len(meet) {
		debug.PrintStack()
		return nil, fmt.Errorf("%s did not match data file. master: %d, strings: %d",
			"Character meet", len(Data.CardCharacters), len(meet))
	}

	battleStart, err := ReadStringFile(root + "/string/MsgCharaBtlStart_en.strb")
	if err != nil {
		debug.PrintStack()
		return nil, err
	}
	if len(Data.CardCharacters) > len(battleStart) {
		debug.PrintStack()
		return nil, fmt.Errorf("%s did not match data file. master: %d, strings: %d",
			"Character battle_start", len(Data.CardCharacters), len(battleStart))
	}

	battleEnd, err := ReadStringFile(root + "/string/MsgCharaBtlEnd_en.strb")
	if err != nil {
		debug.PrintStack()
		return nil, err
	}
	if len(Data.CardCharacters) > len(battleEnd) {
		debug.PrintStack()
		return nil, fmt.Errorf("%s did not match data file. master: %d, strings: %d",
			"Character battle_end", len(Data.CardCharacters), len(battleEnd))
	}

	friendshipMax, err := ReadStringFile(root + "/string/MsgCharaFriendshipMax_en.strb")
	if err != nil {
		debug.PrintStack()
		return nil, err
	}
	if len(Data.CardCharacters) > len(friendshipMax) {
		debug.PrintStack()
		return nil, fmt.Errorf("%s did not match data file. master: %d, strings: %d",
			"Character friendship_max", len(Data.CardCharacters), len(friendshipMax))
	}

	friendshipEvent, err := ReadStringFile(root + "/string/MsgCharaBonds_en.strb")
	if err != nil {
		debug.PrintStack()
		return nil, err
	}
	if len(Data.CardCharacters) > len(friendshipEvent) {
		debug.PrintStack()
		return nil, fmt.Errorf("%s did not match data file. master: %d, strings: %d",
			"Character friendship_event", len(Data.CardCharacters), len(friendshipEvent))
	}

	rebirthEvent, err := ReadStringFile(root + "/string/MsgCharaSuperAwaken_en.strb")
	if err != nil {
		debug.PrintStack()
		return nil, err
	}
	if len(Data.CardCharacters) > len(rebirthEvent) {
		debug.PrintStack()
		return nil, fmt.Errorf("%s did not match data file. master: %d, strings: %d",
			"Character friendship_event", len(Data.CardCharacters), len(rebirthEvent))
	}

	for key := range Data.CardCharacters {
		Data.CardCharacters[key].Description = strings.Replace(description[key], "\n", " ", -1)
		Data.CardCharacters[key].Friendship = friendship[key]
		if key < len(login) {
			Data.CardCharacters[key].Login = login[key]
		}
		Data.CardCharacters[key].Meet = meet[key]
		Data.CardCharacters[key].BattleStart = battleStart[key]
		Data.CardCharacters[key].BattleEnd = battleEnd[key]
		Data.CardCharacters[key].FriendshipMax = friendshipMax[key]
		Data.CardCharacters[key].FriendshipEvent = friendshipEvent[key]
		Data.CardCharacters[key].Rebirth = rebirthEvent[key]
	}
	description = nil
	friendship = nil
	login = nil
	meet = nil
	battleStart = nil
	battleEnd = nil
	friendshipMax = nil
	friendshipEvent = nil

	//Read Skill strings
	names, err = ReadStringFile(root + "/string/MsgSkillName_en.strb")
	if err != nil {
		debug.PrintStack()
		return nil, err
	}

	description, err = ReadStringFile(root + "/string/MsgSkillDesc_en.strb")
	if err != nil {
		debug.PrintStack()
		return nil, err
	}

	fire, err := ReadStringFile(root + "/string/MsgSkillFire_en.strb")
	if err != nil {
		debug.PrintStack()
		return nil, err
	}

	for key := range Data.Skills {
		if key < len(names) {
			Data.Skills[key].Name = filterSkill(names[key])
		}
		if key < len(description) {
			Data.Skills[key].Description = filterSkill(description[key])
		}
		if key < len(fire) {
			Data.Skills[key].Fire = filterSkill(fire[key])
		}
	}

	// event strings
	evntNames, err := ReadStringFile(root + "/string/MsgEventName_en.strb")
	if err != nil {
		debug.PrintStack()
		return nil, err
	}
	evntDescrs, err := ReadStringFile(root + "/string/MsgEventDesc_en.strb")
	if err != nil {
		debug.PrintStack()
		return nil, err
	}

	for key := range Data.Events {
		evntID := Data.Events[key].ID - 1
		if evntID < len(evntNames) {
			Data.Events[key].Name = filter(evntNames[evntID])
		}
		if evntID < len(evntDescrs) {
			Data.Events[key].Description = filterElementImages(filter(filterColors(evntDescrs[evntID])))
		}
	}

	// map strings
	mapNames, err := ReadStringFile(root + "/string/MsgNPCMapName_en.strb")
	if err != nil {
		debug.PrintStack()
		return nil, err
	}

	mapStart, err := ReadStringFile(root + "/string/MsgNPCMapStart_en.strb")
	if err != nil {
		debug.PrintStack()
		return nil, err
	}

	for key := range Data.Maps {
		if key < len(mapNames) {
			Data.Maps[key].Name = mapNames[key]
		}
		if key < len(mapStart) {
			Data.Maps[key].StartMsg = filter(filterColors(mapStart[key]))
		}
	}

	areaName, err := ReadStringFile(root + "/string/MsgNPCAreaName_en.strb")
	if err != nil {
		debug.PrintStack()
		return nil, err
	}
	areaLongName, err := ReadStringFile(root + "/string/MsgNPCAreaLongName_en.strb")
	if err != nil {
		debug.PrintStack()
		return nil, err
	}
	areaStart, err := ReadStringFile(root + "/string/MsgNPCAreaStart_en.strb")
	if err != nil {
		debug.PrintStack()
		return nil, err
	}
	areaEnd, err := ReadStringFile(root + "/string/MsgNPCAreaEnd_en.strb")
	if err != nil {
		debug.PrintStack()
		return nil, err
	}
	areaStory, err := ReadStringFile(root + "/string/MsgNPCAreaStory_en.strb")
	if err != nil {
		debug.PrintStack()
		return nil, err
	}
	bossStart, err := ReadStringFile(root + "/string/MsgNPCBossEnd_en.strb")
	if err != nil {
		debug.PrintStack()
		return nil, err
	}
	bossEnd, err := ReadStringFile(root + "/string/MsgNPCBossStart_en.strb")
	if err != nil {
		debug.PrintStack()
		return nil, err
	}

	for key := range Data.Areas {
		if key < len(bossStart) {
			Data.Areas[key].BossStart = filterColors(bossStart[key])
		}
		if key < len(bossEnd) {
			Data.Areas[key].BossEnd = filterColors(bossEnd[key])
		}
		if key < len(areaStart) {
			Data.Areas[key].Start = filterColors(areaStart[key])
		}
		if key < len(areaEnd) {
			Data.Areas[key].End = filterColors(areaEnd[key])
		}
		if key < len(areaName) {
			Data.Areas[key].Name = filterColors(areaName[key])
		}
		if key < len(areaLongName) {
			Data.Areas[key].LongName = filterColors(areaLongName[key])
		}
		if key < len(areaStory) {
			Data.Areas[key].Story = filterColors(areaStory[key])
		}
	}

	awlikeability, err := ReadStringFile(root + "/string/MsgKingFriendshipDesc_en.strb")
	if err != nil {
		debug.PrintStack()
		return nil, err
	}

	// Archwitch Likeability
	for key := range Data.ArchwitchFriendships {
		if key < len(awlikeability) {
			Data.ArchwitchFriendships[key].Likability = filter(awlikeability[key])
		}
	}

	kingDescription, err := ReadStringFile(root + "/string/MsgKingTitle_en.strb")
	// king series descriptions
	for key := range Data.ArchwitchSeries {
		if key < len(kingDescription) {
			Data.ArchwitchSeries[key].Description = filter(kingDescription[key])
		}
	}

	dbonusName, err := ReadStringFile(root + "/string/MsgDeckBonusName_en.strb")
	if err != nil {
		debug.PrintStack()
		return nil, err
	}
	dbonusDesc, err := ReadStringFile(root + "/string/MsgDeckBonusDesc_en.strb")
	if err != nil {
		debug.PrintStack()
		return nil, err
	}

	// Deck Bonuses
	for key := range Data.DeckBonuses {
		if key < len(dbonusName) {
			Data.DeckBonuses[key].Name = filter(dbonusName[key])
		}
		if key < len(dbonusDesc) {
			Data.DeckBonuses[key].Description = filter(dbonusDesc[key])
		}
	}

	//Items
	itemdsc, err := ReadStringFile(root + "/string/MsgShopItemDesc_en.strb")
	if err != nil {
		debug.PrintStack()
		return nil, err
	}
	itemdscshp, err := ReadStringFile(root + "/string/MsgShopItemDescInShop_en.strb")
	if err != nil {
		debug.PrintStack()
		return nil, err
	}
	itemdscsub, err := ReadStringFile(root + "/string/MsgShopItemDescSub_en.strb")
	if err != nil {
		debug.PrintStack()
		return nil, err
	}
	itemname, err := ReadStringFile(root + "/string/MsgShopItemName_en.strb")
	if err != nil {
		debug.PrintStack()
		return nil, err
	}
	itemuse, err := ReadStringFile(root + "/string/MsgShopItemUseResult_en.strb")
	if err != nil {
		debug.PrintStack()
		return nil, err
	}

	for key := range Data.Items {
		if key < len(itemdsc) {
			Data.Items[key].Description = filter(itemdsc[key])
		}
		if key < len(itemdscshp) {
			Data.Items[key].DescriptionInShop = filter(itemdscshp[key])
		}
		if key < len(itemdscsub) {
			Data.Items[key].DescriptionSub = filter(itemdscsub[key])
		}
		if key < len(itemname) {
			Data.Items[key].NameEng = filter(itemname[key])
		}
		if key < len(itemuse) {
			Data.Items[key].MsgUse = filter(itemuse[key])
		}
	}

	buildname, err := ReadStringFile(root + "/string/MsgBuildingName_en.strb")
	if err != nil {
		debug.PrintStack()
		return nil, err
	}
	builddesc, err := ReadStringFile(root + "/string/MsgBuildingDesc_en.strb")
	if err != nil {
		debug.PrintStack()
		return nil, err
	}

	for key := range Data.Structures {
		if key < len(buildname) {
			Data.Structures[key].Name = filter(buildname[key])
		}
		if key < len(builddesc) {
			Data.Structures[key].Description = filter(builddesc[key])
		}
	}

	if Data.ThorEvents != nil {
		thorTitle, err := ReadStringFile(root + "/string/MsgThorhammerTitle_en.strb")
		if err != nil {
			debug.PrintStack()
			return data, err
		}
		for key := range Data.ThorEvents {
			if key < len(thorTitle) {
				Data.ThorEvents[key].Title = filter(thorTitle[key])
			}
		}
	}
	return data, nil
}

//ReadStringFile Reads a binary string file
func ReadStringFile(fname string) ([]string, error) {
	filename := strings.Replace(fname, "_en.strb", "_"+LangPack+".strb", 1)
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		debug.PrintStack()
		return nil, errors.New("no such file or directory: " + filename)
	}
	f, err := os.Open(filename)
	if err != nil {
		debug.PrintStack()
		return nil, errors.New("Error opening: " + filename)
	}
	defer f.Close()

	r := bufio.NewReader(f)

	//skip the 8 byte header
	_, err = r.Discard(8)
	if err != nil {
		debug.PrintStack()
		return nil, errors.New("Error skipping the file header for file " + filename)
	}

	// find the "null" seperator between the binary info and the strings
	null := []byte("null\000")
	var line []byte
	for {
		if line, err = r.ReadBytes('\000'); err != nil {
			debug.PrintStack()
			return nil, errors.New("Error reading the file " + filename)
		}
		if bytes.Equal(line, null) {
			break
		}
	}

	//read the strings
	ret := make([]string, 0)
	for {
		if line, err = r.ReadBytes('\000'); err == io.EOF {
			break
		}
		if err != nil {
			debug.PrintStack()
			return nil, errors.New("Error reading the file " + filename)
		}
		// remove the null terminator
		ret = append(ret, filter(string(line[:len(line)-1])))
	}
	return ret, nil
}

//ReadBinFileImages reads a binary file and returns the image data (PNG only)
func ReadBinFileImages(filename string) ([]BinImage, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	l := len(data)

	nameStart := []byte("\x00\x00\x00")
	lnameStart := len(nameStart)
	nameEnd := byte('\000')

	pngStart := []byte("\x89PNG")
	lpngStart := len(pngStart)
	pngEnd := []byte("IEND\xAEB`\x82")
	lpngEnd := len(pngEnd)

	findNameStart := func(data []byte, startIdx int) int {
		for i := startIdx; i < (l - lnameStart); i++ {
			if bytes.Equal(data[i:i+lnameStart], nameStart) {
				if i+lnameStart+1 < l {
					c := data[i+lnameStart]
					if c >= 'a' && c <= 'z' {
						return i + lnameStart // exclude the 3 null bytes
					}
				}
			}
		}
		return -1
	}
	findNameEnd := func(data []byte, startIdx int) int {
		for i := startIdx; i < (l - 1); i++ {
			if data[i] == nameEnd {
				return i
			}
		}
		return -1
	}
	findPngStart := func(data []byte, startIdx int) int {
		for i := startIdx; i < (l - lpngStart); i++ {
			if bytes.Equal(data[i:i+lpngStart], pngStart) {
				return i
			}
		}
		return -1
	}
	findPngEnd := func(data []byte, startIdx int) int {
		for i := startIdx; i < (l - lpngEnd); i++ {
			if bytes.Equal(data[i:i+lpngEnd], pngEnd) {
				return i + lpngEnd
			}
		}
		return -1
	}
	//start of the PNG image
	firstPng := findPngStart(data, 0)
	if firstPng < 0 {
		return nil, errors.New("unable to locate any images")
	}
	//parse names
	start := 0
	names := make([]string, 0)
	for start < firstPng {
		start = findNameStart(data, start)
		if start < 0 || start > firstPng {
			break
		}
		end := findNameEnd(data, start)
		if end < 0 || end > firstPng {
			break
		}
		if end-start < 5 { // none of the names are shorter than 5 characters
			start = end
			continue
		}
		name := string(data[start:end])
		names = append(names, name)
		//log.Printf("found image name '%s', idx: %d-%d\n", name, start, end)
		start = end
	}

	lnames := len(names)
	getImageName := func(idx int) string {
		if idx < lnames && names[idx] != "" {
			return names[idx] + ".png"
		}
		return fmt.Sprintf("structure_%05d.png", idx+1)
	}

	start = firstPng
	// look for PNG images
	ret := make([]BinImage, 0)
	i := 0 // skip the "dummy" name
	for start < (l - (lpngStart + lpngEnd)) {
		i++
		start = findPngStart(data, start)
		if start < 0 {
			break
		}
		end := findPngEnd(data, start)
		if end < 0 {
			return nil, errors.New("unable to locate the end of an image")
		}

		ret = append(ret, BinImage{ID: i, Name: getImageName(i), Data: data[start:end]})
		//log.Printf("found image, idx: %d\n", start)
		start = end
	}
	log.Printf("found %d image names and %d images\n",
		lnames,
		len(ret),
	)

	return ret, nil
}

func cleanCardName(name string, card *Card) string {
	ret := ""
	if newName, ok := characterNameOverride[card.CardCharaID]; ok {
		// use an overridden hard-coded name
		ret = newName
	} else {
		ret = strings.Replace(strings.Title(strings.ToLower(name)), "'S", "'s", -1)
		ret = strings.Replace(ret, "(Sr)", "(SR)", -1)
		ret = strings.Replace(ret, "(Ur)", "(UR)", -1)
		ret = strings.Replace(ret, "(Lr)", "(LR)", -1)
		if card.CardCharaID < 1450 {
			// use lowecase prepositions and articles as these are cards in the wiki before this program.
			ret = strings.Replace(ret, " Of ", " of ", -1)
			ret = strings.Replace(ret, "-Of-", "-of-", -1)
			ret = strings.Replace(ret, " The ", " the ", -1)
			ret = strings.Replace(ret, "-The-", "-the-", -1)
			ret = strings.Replace(ret, " In ", " in ", -1)
			ret = strings.Replace(ret, "-In-", "-in-", -1)
			ret = strings.Replace(ret, " O'", " o'", -1)
			ret = strings.Replace(ret, "-O'", "-o'", -1)
			ret = strings.Replace(ret, " Du ", " du ", -1) // french "of"
		}
	}
	// old cards
	if card.IsRetired() {
		ret += " (Old)"
	} else {
		// new cards that are named the same as an old card that is still active
		newIDx := sort.SearchInts(newCards, card.ID)
		if newIDx >= 0 && newIDx < len(newCards) && newCards[newIDx] == card.ID {
			ret += " (New)"
		}
	}
	return ret
}

// GetBinFileImages gets a subset of images from the bin index. 1-based index.
func GetBinFileImages(filename string, idxs ...int) ([]BinImage, error) {
	if len(idxs) == 0 {
		return nil, errors.New("Index out of bounds")
	}
	images, err := ReadBinFileImages(filename)
	if err != nil {
		return nil, err
	}
	ret := make([]BinImage, 0, len(idxs))
	for _, idx := range idxs {
		if idx < 1 || idx > len(images) {
			return nil, errors.New("Index out of bounds")
		}
		ret = append(ret, images[idx-1])
	}
	return ret, nil
}

//Use this to do common string replacements in the VC data files
func filter(s string) string {
	if s == "null" {
		return ""
	}
	ret := strings.TrimSpace(s)
	// standardize utf enocoded symbols
	ret = strings.Replace(ret, "％", "%", -1)
	ret = strings.Replace(ret, "　", " ", -1)
	ret = strings.Replace(ret, "／", "/", -1)
	ret = strings.Replace(ret, "＞", ">", -1)
	ret = strings.Replace(ret, "・", " • ", -1)
	// game controls that aren't needed for fandom
	ret = strings.Replace(ret, "<i><break>", "\n", -1)
	// remove duplicate newlines
	for strings.Contains(ret, "\n\n") {
		ret = strings.Replace(ret, "\n\n", "\n", -1)
	}
	//remove duplicate spaces
	for strings.Contains(ret, "  ") {
		ret = strings.Replace(ret, "  ", " ", -1)
	}
	//ret = strings.Replace(ret, "\n", "<br />", -1)

	ret = strings.Replace(ret, "<img=1>Gold", "{{Icon|gold}}", -1)
	ret = strings.Replace(ret, "<img=4>Iron", "{{Icon|iron}}", -1)
	ret = strings.Replace(ret, "<img=3>Ether", "{{Icon|ether}}", -1)
	ret = strings.Replace(ret, "<img=56>Gem", "{{Icon|gem}}", -1)
	ret = strings.Replace(ret, "<img=1>", "{{Icon|gold}}", -1)
	ret = strings.Replace(ret, "<img=4>", "{{Icon|iron}}", -1)
	ret = strings.Replace(ret, "<img=3>", "{{Icon|ether}}", -1)
	ret = strings.Replace(ret, "<img=56>", "{{Icon|gem}}", -1)
	ret = strings.Replace(ret, "<img=5>", "{{Icon|jewel}}", -1)

	return ret
}

func filterElementImages(s string) string {
	ret := strings.TrimSpace(s)
	//element icons
	ret = strings.Replace(ret, "<img=24>", "{{Passion}}", -1)
	ret = strings.Replace(ret, "<img=25>", "{{Cool}}", -1)
	ret = strings.Replace(ret, "<img=26>", "{{Dark}}", -1)
	ret = strings.Replace(ret, "<img=27>", "{{Light}}", -1)
	return ret
}

var regexpSlash = regexp.MustCompile("\\s*[/]\\s*")

func filterSkill(s string) string {
	ret := filterElementImages(s)

	//atk def icons
	ret = strings.Replace(ret, "<img=48>", "{{Atk}}", -1)
	ret = strings.Replace(ret, "<img=51>", "{{Atkdef}}", -1)

	// clean up '/' spacing
	ret = regexpSlash.ReplaceAllString(ret, " / ")
	// make counter attack consistent
	ret = strings.Replace(ret, "% Counter", "%\nCounter", -1)
	ret = strings.Replace(ret, "%, Counter", "%\nCounter", -1)
	return ret
}

func filterColors(s string) string {
	ret := strings.TrimSpace(s)
	rc, _ := regexp.Compile("<col=(.+?)>\\n*")
	ret = rc.ReplaceAllString(ret, "<span class=\"vc_color$1\">")

	rc, _ = regexp.Compile("<colrgb=(.+?)>\\n*")
	ret = rc.ReplaceAllString(ret, "<span style=\"color:rgb($1);\">")

	ret = strings.Replace(ret, "</col>", "</span>", -1)

	// strip all size commands out
	rs, _ := regexp.Compile("<(/?)size(=.+?)?>")
	ret = rs.ReplaceAllLiteralString(ret, "")
	return ret
}
