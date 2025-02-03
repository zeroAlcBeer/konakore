package syncer

import (
	"encoding/json"
	"konakore/pkg/models"
	"testing"
)

func TestUpdateTags(t *testing.T) {
	_, err := models.OpenDb("root:please_change@tcp(192.168.0.110:3307)/konakore?charset=utf8mb4&parseTime=True&loc=Local", "dev")
	if err != nil {
		t.Fatal(err)
	}
	InitDB()

	UpdateTags()

	tws := models.NewTagWeightSystem()

	// 模拟训练数据
	likedPosts := models.GetLikes()
	tws.Learn(likedPosts)

	tests := []struct {
		tags string
	}{
		{
			tags: "blush boots breasts choker couch cum dress gray_hair hat herta_(honkai:_star_rail) honkai_(series) honkai:_star_rail long_hair machi_(7769) necklace nipples no_bra nopan penis purple_eyes pussy sex shirt skirt_lift spread_legs uncensored witch_hat",
		},
		{
			tags: "ass cameltoe close cropped garter mochiko-san_(voicevox) panties tail torimaru underwear voicevox",
		},
		{
			tags: "aegir_(azur_lane) anthropomorphism anus ass azur_lane breasts couch dress horns long_hair nipples nopan pussy skirt_lift spread_legs thighhighs turewindwalker uncensored white_hair yellow_eyes",
		},
		{
			tags: "black_hair dark dress long_hair original shiina_1001",
		},
		{
			tags: "black_hair blush breasts garter halo long_hair nipples no_bra panties pantyhose red_eyes shirt_lift skirt underwear",
		},
		{
			tags: "animal animal_ears asuma_toki bicycle bird black_hair blonde_hair blue_archive blue_eyes bow braids breast_hold breasts brown_eyes brown_hair choker cleavage dark_skin fang flat_chest food foxgirl glasses group halo haowei_wu ichinose_asuna kakudate_karin kneehighs long_hair mikamo_neru murokasa_akane open_shirt phone ponytail red_eyes red_hair scarf school_uniform shirt short_hair skirt sora_(blue_archive) sunaookami_shiroko water wolfgirl yellow_eyes yurizono_seia",
		},
	}

	for _, tt := range tests {
		p := &models.Post{Tags: tt.tags}
		tws.ScorePostV2(p)

		b, _ := json.MarshalIndent(p.Alg, "", " ")
		t.Log(string(b))
		t.Log(p.MyScore)
	}
}
