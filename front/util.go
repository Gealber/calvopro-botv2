package front

import (
	"github.com/goombaio/namegenerator"
	"time"
)

func genName() string {
	seed := time.Now().UTC().UnixNano()
	nameGenerator := namegenerator.NewNameGenerator(seed)
	return nameGenerator.Generate()
}
