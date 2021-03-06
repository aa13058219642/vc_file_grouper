package vc

import (
	"sort"
)

// Tower mst_tower
type Tower struct {
	ID                    int       `json:"_id"`
	PublicStartDatetime   Timestamp `json:"public_start_datetime"`
	PublicEndDatetime     Timestamp `json:"public_end_datetime"`
	RankingStartDatetime  Timestamp `json:"ranking_start_datetime"`
	RankingEndDatetime    Timestamp `json:"ranking_end_datetime"`
	RankingRewardGroupID  int       `json:"ranking_reward_group_id"`
	RankingArrivalGroupID int       `json:"arrival_point_reward_group_id"`
	ElementID             int       `json:"element_id"`
	LevelSkillGroupID     int       `json:"level_skill_group_id"`
	URLSchemeID           int       `json:"url_scheme_id"`
	Title                 string    `json:"-"` // MsgTowerTitle_en.strb
}

// RankRewards rank rewards for the tower
func (t *Tower) RankRewards() []RankRewardSheet {
	set := make([]RankRewardSheet, 0)
	if t.RankingRewardGroupID > 0 {
		for _, val := range Data.TowerRewards {
			if val.SheetID == t.RankingRewardGroupID {
				set = append(set, val)
			}
		}
	}
	return set
}

// ArrivalRewards rewards for arriving at a certain Tower floor
func (t *Tower) ArrivalRewards() []RankRewardSheet {
	set := make([]RankRewardSheet, 0)
	if t.RankingArrivalGroupID > 0 {
		for _, val := range Data.TowerArrivalRewards {
			if val.SheetID == t.RankingArrivalGroupID {
				set = append(set, val)
			}
		}
	}
	return set
}

// TowerScan search for a tower by ID
func TowerScan(id int) *Tower {
	if id > 0 {
		l := len(Data.Towers)
		i := sort.Search(l, func(i int) bool { return Data.Towers[i].ID >= id })
		if i >= 0 && i < l && Data.Towers[i].ID == id {
			return &(Data.Towers[i])
		}
	}
	return nil

}
