package main

import (
	"log"
	"os"

	"github.com/urfave/cli"

	"github.com/KyberNetwork/reserve-data/cmd/configuration"
	"github.com/KyberNetwork/reserve-data/cmd/deployment"
	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/common/profiler"
	"github.com/KyberNetwork/reserve-data/http"
	"github.com/KyberNetwork/reserve-data/lib/app"
)

func main() {
	app := cli.NewApp()
	app.Name = "Reserve Core"
	app.Usage = "Kyber Reserve core component that helps manage reserves of tokens"
	app.Version = "0.11.0"
	app.Action = run

	app.Flags = configuration.NewCliFlags()
	app.Flags = append(app.Flags, profiler.NewCliFlags()...)

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func run(c *cli.Context) error {
	configuration.SetupLogging()

	dpl, err := deployment.NewDeploymentFromContext(c)
	if err != nil {
		return err
	}
	// TODO: port reserve core to use zap too
	sugar, flusher, err := app.NewSugaredLogger(c)
	if err != nil {
		panic(err)
	}
	defer func() {
		flusher()
	}()
	conf, err := configuration.NewConfigurationFromContext(c, sugar)
	if err != nil {
		return err
	}

	bc, err := configuration.CreateBlockchain(conf)
	if err != nil {
		log.Printf("Can not create blockchain: (%s)", err)
		return err
	}

	dryRun := configuration.NewDryRunFromContext(c)

	rData, rCore := configuration.CreateDataCore(conf, dpl, bc)
	if !dryRun {
		if dpl != deployment.Simulation {
			if err = rData.RunStorageController(); err != nil {
				log.Printf("failed to run storage controller err=%s", err.Error())
				return err
			}
		}
		if err = rData.Run(); err != nil {
			log.Printf("failed to run data service err=%s", err.Error())
			return err
		}
	}

	for _, ex := range conf.Exchanges {
		common.SupportedExchanges[ex.ID()] = ex
	}

	host := configuration.NewHTTPAddressFromContext(c)
	server := http.NewHTTPServer(
		rData, rCore,
		host,
		dpl,
		bc,
		conf.SettingStorage,
	)
	if profiler.IsEnableProfilerFromContext(c) {
		server.EnableProfiler()
	}

	if !dryRun {
		server.Run()
	} else {
		log.Printf("Dry run finished. All configs are corrected")
	}

	return err
}
