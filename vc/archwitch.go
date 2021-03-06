package vc

import (
	"fmt"
	"sort"
)

// Archwitch represents the "kings" istructure in the data file
type Archwitch struct {
	ID            int    `json:"_id"`
	KingSeriesID  int    `json:"king_series_id"`
	CardMasterID  int    `json:"card_master_id"`
	StatusGroupID int    `json:"status_group_id"`
	PublicFlg     int    `json:"public_flg"`
	RareFlg       int    `json:"rare_flg"`
	RareIntensity int    `json:"rare_intensity"`
	BattleTime    int    `json:"battle_time"`
	Exp           int    `json:"exp"`
	MaxFriendship int    `json:"max_friendship"`
	SkillID1      int    `json:"skill_id_1"`
	SkillID2      int    `json:"skill_id_2"`
	WeatherID     int    `json:"weather_id"`
	ModelName     string `json:"model_name"`
	ChainRatio2   int    `json:"chain_ratio_2"`
	ServantID1    int    `json:"servant_id_1"`
	ServantID2    int    `json:"servant_id_2"`
	likeability   []ArchwitchFriendship
}

// ArchwitchList helper interface for looking at lists of Archwitches
type ArchwitchList []*Archwitch

// ArchwitchSeries is "king_series" in the data file and is the
// list of AW events and the RR prizes Indicates what "event"
// the archwitch was a part of.
type ArchwitchSeries struct {
	ID                   int       `json:"_id"`
	RewardCardID         int       `json:"reward_card_id"`
	PublicStartDatetime  Timestamp `json:"public_start_datetime"`
	PublicEndDatetime    Timestamp `json:"public_end_datetime"`
	ReceiveLimitDatetime Timestamp `json:"receive_limit_datetime"`
	IsBeginnerKing       int       `json:"is_beginner_king"`
	Description          string    `json:"-"`
	archwitches          ArchwitchList
}

//ArchwitchFriendship is "king_friendship" in the data file.
// It is the chance of friendship increasing on an AW.
type ArchwitchFriendship struct {
	ID         int    `json:"_id"`
	KingID     int    `json:"king_id"`
	Friendship int    `json:"friendship"`
	UpRate     int    `json:"up_rate"`
	Likability string `json:"-"` // MsgKingFriendshipDesc_en.strb
}

/*
Limited items and enemies `lmtd`
*/
type Limited struct {
	ID               int       `json:"_id"`
	LmtdTypeID       int       `json:"lmtd_type_id"`
	DayTypeID        int       `json:"day_type_id"`
	StartDatetime    Timestamp `json:"start_datetime"`
	EndDatetime      Timestamp `json:"end_datetime"`
	GroupID          int       `json:"group_id"`
	Enable           int       `json:"enable"`
	JumpBanner       int       `json:"jump_banner"`
	MapID            int       `json:"map_id"`
	QuestID          int       `json:"quest_id"`
	JumpButton       int       `json:"jump_button"`
	TexIDCell        int       `json:"tex_id_cell"`
	TexIDImage       int       `json:"tex_id_image"`
	TexIDImage2      int       `json:"tex_id_image2"`
	EnableDoButton   int       `json:"enable_do_button"`
	EnableRewardDisp int       `json:"enable_reward_disp"`
	RewardCardID     int       `json:"reward_card_id"`
	RewardItemID     int       `json:"reward_item_id"`
}

// Likeability information for the AW
func (a *Archwitch) Likeability() []ArchwitchFriendship {
	if a == nil {
		return []ArchwitchFriendship{}
	}
	if a.likeability == nil {
		a.likeability = make([]ArchwitchFriendship, 0)
		for _, af := range Data.ArchwitchFriendships {
			if a.ID == af.KingID {
				a.likeability = append(a.likeability, af)
			}
		}
		sort.Slice(a.likeability, func(l1, l2 int) bool {
			return a.likeability[l1].Friendship < a.likeability[l2].Friendship
		})
	}
	return a.likeability
}

// IsFAW returns true if this AW is a FAW
func (a *Archwitch) IsFAW() bool {
	return a.ServantID1 > 0 && a.StatusGroupID != 25
}

// IsLAW returns true if this AW is a LAW
func (a *Archwitch) IsLAW() bool {
	return a.ServantID1 > 0 && a.StatusGroupID == 25
}

// IsAW returns true if this AW is a normal AW (not a FAW or a LAW)
func (a *Archwitch) IsAW() bool {
	return !(a.IsFAW() || a.IsLAW())
}

// Archwitches in this AW Series
func (as *ArchwitchSeries) Archwitches() ArchwitchList {
	if as.archwitches == nil {
		as.archwitches = make(ArchwitchList, 0)
		for _, a := range Data.Archwitches {
			if as.ID == a.KingSeriesID {
				as.archwitches = append(as.archwitches, a)
			}
		}
		sort.Slice(as.archwitches, func(a, b int) bool {
			awa := as.archwitches[a]
			awb := as.archwitches[b]
			return awa.ID < awb.ID
		})
	}
	return as.archwitches
}

// Earliest gets the earliest AW from the list
func (awl ArchwitchList) Earliest() (first *Archwitch) {
	l := len(awl)
	if l == 0 {
		return
	}

	for _, aw := range awl {
		if first == nil || first.ID > aw.ID {
			first = aw
		}
	}
	return
}

// Latest gets the latest AW from the list
func (awl ArchwitchList) Latest() (last *Archwitch) {
	l := len(awl)
	if l == 0 {
		return
	}

	for _, aw := range awl {
		if last == nil || last.ID < aw.ID {
			last = aw
		}
	}
	return
}

//Contains checks if the test item exists in the list
func (awl ArchwitchList) Contains(test *Archwitch) bool {
	for _, aw := range awl {
		if aw == test {
			return true
		}
		if aw != nil && test != nil && aw.ID == test.ID {
			return true
		}
	}
	return false
}

func (awl ArchwitchList) String() (ret string) {
	for _, aw := range awl {
		ret += fmt.Sprintf("{id: %d}", aw.ID)
	}
	return
}
