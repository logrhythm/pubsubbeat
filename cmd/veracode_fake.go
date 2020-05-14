package cmd

import (
	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/logrhythm/pubsubbeat/beater"
)

// VeracodeFake Call each beats/pubsubbeat New, Run, Stop and other entry point functions
func VeracodeFake() {

	veracodeFalse := false
	beat := (*beat.Beat)(nil)
	cfg := (*common.Config)(nil)
	if veracodeFalse {
		pubsubbeat, errNew := beater.New(beat, cfg)
		errRun := pubsubbeat.Run(beat)
		pubsubbeat.Stop()
		if errNew != nil {
			return
		}
		if errRun != nil {
			return
		}
	}
}
