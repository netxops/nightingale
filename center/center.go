package center

import (
	"context"
	"fmt"
	"github.com/ccfos/nightingale/v6/pushgw/labeler"
	"github.com/toolkits/pkg/logger"
	"time"

	"github.com/ccfos/nightingale/v6/alert"
	"github.com/ccfos/nightingale/v6/alert/astats"
	"github.com/ccfos/nightingale/v6/alert/process"
	"github.com/ccfos/nightingale/v6/center/cconf"
	"github.com/ccfos/nightingale/v6/center/metas"
	"github.com/ccfos/nightingale/v6/center/sso"
	"github.com/ccfos/nightingale/v6/conf"
	"github.com/ccfos/nightingale/v6/memsto"
	"github.com/ccfos/nightingale/v6/models"
	"github.com/ccfos/nightingale/v6/pkg/ctx"
	"github.com/ccfos/nightingale/v6/pkg/httpx"
	"github.com/ccfos/nightingale/v6/pkg/i18nx"
	"github.com/ccfos/nightingale/v6/pkg/logx"
	"github.com/ccfos/nightingale/v6/prom"
	"github.com/ccfos/nightingale/v6/pushgw/idents"
	"github.com/ccfos/nightingale/v6/pushgw/writer"
	"github.com/ccfos/nightingale/v6/storage"

	alertrt "github.com/ccfos/nightingale/v6/alert/router"
	centerrt "github.com/ccfos/nightingale/v6/center/router"
	pushgwrt "github.com/ccfos/nightingale/v6/pushgw/router"
)

func Initialize(configDir string, cryptoKey string) (func(), error) {
	config, err := conf.InitConfig(configDir, cryptoKey)
	if err != nil {
		return nil, fmt.Errorf("failed to init config: %v", err)
	}

	cconf.LoadMetricsYaml(configDir, config.Center.MetricsYamlFile)
	cconf.LoadOpsYaml(configDir, config.Center.OpsYamlFile)

	logxClean, err := logx.Init(config.Log)
	if err != nil {
		return nil, err
	}

	i18nx.Init()

	db, err := storage.New(config.DB)
	if err != nil {
		return nil, err
	}
	ctx := ctx.NewContext(context.Background(), db, true)
	models.InitRoot(ctx)

	redis, err := storage.NewRedis(config.Redis)
	if err != nil {
		return nil, err
	}

	metas := metas.New(redis)
	idents := idents.New(ctx)

	syncStats := memsto.NewSyncStats()
	alertStats := astats.NewSyncStats()

	sso := sso.Init(config.Center, ctx)

	busiGroupCache := memsto.NewBusiGroupCache(ctx, syncStats)
	targetCache := memsto.NewTargetCache(ctx, syncStats, redis)
	dsCache := memsto.NewDatasourceCache(ctx, syncStats)
	alertMuteCache := memsto.NewAlertMuteCache(ctx, syncStats)
	alertRuleCache := memsto.NewAlertRuleCache(ctx, syncStats)
	notifyConfigCache := memsto.NewNotifyConfigCache(ctx)

	promClients := prom.NewPromClient(ctx, config.Alert.Heartbeat)

	externalProcessors := process.NewExternalProcessors()
	alert.Start(config.Alert, config.Pushgw, syncStats, alertStats, externalProcessors, targetCache, busiGroupCache, alertMuteCache, alertRuleCache, notifyConfigCache, dsCache, ctx, promClients)

	writers := writer.NewWriters(config.Pushgw)

	alertrtRouter := alertrt.New(config.HTTP, config.Alert, alertMuteCache, targetCache, busiGroupCache, alertStats, ctx, externalProcessors)
	centerRouter := centerrt.New(config.HTTP, config.Center, cconf.Operations, dsCache, notifyConfigCache, promClients, redis, sso, ctx, metas, targetCache)
	pushgwRouter := pushgwrt.New(config.HTTP, config.Pushgw, targetCache, busiGroupCache, idents, writers, ctx)
	loopTagTask(*pushgwRouter)

	r := httpx.GinEngine(config.Global.RunMode, config.HTTP)

	centerRouter.Config(r)
	alertrtRouter.Config(r)
	pushgwRouter.Config(r)

	httpClean := httpx.Init(config.HTTP, r)

	return func() {
		logxClean()
		httpClean()
	}, nil
}

func loopTagTask(pushgwRouter pushgwrt.Router) {
	task := func() {
		for {
			duration, _ := time.ParseDuration("30s")
			time.Sleep(duration)
			richLabels := pushgwRouter.EnrichLabelsFromRedis()
			labeler.REDIS_TAGS = richLabels
			//logger.Infof("源标签数据：%#v", labeler.REDIS_TAGS)
			logger.Infof("==================源标签数据=================")
			for _, pair := range richLabels {
				fmt.Println("-----------------------------------------------")
				fmt.Println(fmt.Sprintf("name:%s, resource_key:%s, ip:%s, id:%s, tags:[", pair.DeviceName, pair.ResourceKey, pair.IP, pair.ResourceId))
				for _, tag := range pair.Tags {
					fmt.Print(fmt.Sprintf("%s:%s", tag.TagName, tag.TagValue), "  ||  ")
				}
				fmt.Println("]")
			}
			//logger.Infof("源标签数据：%#v", labeler.REDIS_TAGS)
		}
	}

	go task()
}
