package api

type PopularityObject struct {
	Id      string         `json:"Id" bson:"Id"`
	Hourly  PopularityInfo `json:"Hourly" bson:"Hourly"`
	Daily   PopularityInfo `json:"Daily" bson:"Daily"`
	Weekly  PopularityInfo `json:"Weekly" bson:"Weekly"`
	Monthly PopularityInfo `json:"Monthly" bson:"Monthly"`
	Yearly  PopularityInfo `json:"Yearly" bson:"Yearly"`
}

type PopularityInfo struct {
	Usages int `json:"Usages" bson:"Usages"`
}

func (popularity PopularityObject) AddPopularityUsage() {
	popularity.Hourly.Usages++
	popularity.Daily.Usages++
	popularity.Weekly.Usages++
	popularity.Monthly.Usages++
	popularity.Yearly.Usages++
}

func CreatePopularity(id string) PopularityObject {
	popularity := PopularityObject{
		Id:      id,
		Hourly:  PopularityInfo{},
		Daily:   PopularityInfo{},
		Weekly:  PopularityInfo{},
		Monthly: PopularityInfo{},
		Yearly:  PopularityInfo{},
	}
	return popularity
}
