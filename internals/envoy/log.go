package envoy

import (
	"github.com/sepaper/envoy-hot-restarter/internals/util"
	log "github.com/sirupsen/logrus"
)

var (
	logger log.FieldLogger
)

func init() {
	logger = util.GetLogger()
}
