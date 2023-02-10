package cli

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/1-platform/api-catalog/internal/cli/compiler"
	"github.com/1-platform/api-catalog/internal/cli/filereader"
	"github.com/1-platform/api-catalog/internal/cli/pluginmanager"
	"github.com/1-platform/api-catalog/internal/cli/reportmanager"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type reportRuleMetrics struct {
	TotalRules  int `json:"total_rules" toml:"total_rules"`
	PassedRules int `json:"rules_passed" toml:"rules_passed"`
}

// final report export data
type reportExportData struct {
	Metrics    *reportRuleMetrics           `json:"metrics" toml:"metrics"`
	RuleReport *reportmanager.ReportManager `json:"reports" toml:"reports"`
}

// util
func unzipBuiltinPlugins(zipContent []byte, zipDir string) {
	reader, _ := zip.NewReader(bytes.NewReader(zipContent), int64(len(zipContent)))
	for _, f := range reader.File {
		fp := filepath.Join(zipDir, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(fp, os.ModePerm)
			continue
		}

		os.MkdirAll(filepath.Dir(fp), os.ModePerm)
		outFile, _ := os.OpenFile(fp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		rc, _ := f.Open()
		io.Copy(outFile, rc)

		outFile.Close()
		rc.Close()
	}
}

// some checks on running run cmd
func bootUpChecks(fr *filereader.FileReader, logger *CliLogger) {
	var versionFile map[string]string
	dir, _ := os.UserHomeDir()
	// check builtin module exit
	logger.Info("Checking builtin plugins are installed")
	err := fr.ReadFile(path.Join(dir, ".apic/builtin/version.json"), &versionFile)
	// if file exist and version is latest
	if err == nil && versionFile["version"] == version {
		logger.Info("Found latest builtin plugin")
		return
	}

	if err := os.RemoveAll(path.Join(dir, ".apic")); err != nil {
		log.Fatal(err)
	}

	// else download the file
	logger.Info("Outdated or missing builtin plugin. Installing latest one...")
	builtInZipfile, err := fr.ReadIntoRawBytes("https://github.com/1-Platform/api-catalog/raw/main/plugins/builtin.zip")
	if err != nil {
		log.Fatal(err)
	}

	unzipBuiltinPlugins(builtInZipfile, path.Join(dir, ".apic"))
	logger.Success("Builtin plugins successfully installed")
}

func runCommand(_cmd *cobra.Command, _args []string) {
	// find config file and load up the config
	var config ApiCatalogConfig

	// setup cli logger
	logger := NewCliLogger()

	configExt := filepath.Ext(configFilePath)
	if configExt == "" {
		viper.AddConfigPath(configFilePath)
		viper.SetConfigName("apic")
	} else {
		viper.AddConfigPath(strings.TrimSuffix(configFilePath, filepath.Base(configFilePath)))
		viper.SetConfigName(strings.TrimSuffix(filepath.Base(configFilePath), configExt))
	}

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			logger.Warn("No confile was found. Using default options")
		} else {
			log.Fatal("Failed to read config file\n", err)
		}
	}

	if err := viper.Unmarshal(&config); err != nil {
		log.Fatal("Error in config file\n", err)
	}

	fr, err := filereader.New()
	if err != nil {
		log.Fatal("Failed to load filereader\n", err)
	}

	if version != "development" {
		bootUpChecks(fr, logger)
	}

	// set up the js script compiler
	cmp, err := compiler.New(logger)
	if err != nil {
		log.Fatal("Error in setting up compiler\n", err)
	}

	// loading up the plugins and corresponding rules
	pManager := pluginmanager.New(fr, apiType, version == "development")
	if err := pManager.LoadBuiltinPlugin(); err != nil {
		log.Fatal(err)
	}
	logger.Completed("Loaded builtin plugins")

	if err := pManager.LoadUserPlugins(config.Plugins); err != nil {
		log.Fatal(err)
	}
	logger.Completed("Loaded user defined plugins")

	if err := pManager.OverrideRules(config.Rules); err != nil {
		log.Fatal(err)
	}

	var apiSchemaFile map[string]interface{}
	raw, err := fr.ReadFileReturnRaw(apiSchemaURL, &apiSchemaFile)
	if err != nil {
		log.Fatal("Failed to read file\n", err)
	}
	logger.Completed("Read and parsed API schema file")

	rm := reportmanager.New()

	// validation
	switch apiType {
	case "openapi":
		if err := ValidateOpenAPI(raw, apiSchemaFile, logger); err != nil {
			log.Fatal("Failed to validate openapi schema\n", err)
		}
		logger.Success("OpenAPI validation check passed")
		// iterate over rule
	default:
		logger.Error(fmt.Sprintf("Error api type not supported: %s", apiType))
		os.Exit(0)
	}

	rulesPassedCounter := 0
	totalRules := len(pManager.Rules)
	for rule, opt := range pManager.Rules {
		if opt.Disable {
			logger.Info(fmt.Sprintf("%s has been disabled", rule))
			continue
		}

		// read original code
		rawCode, err := pManager.ReadPluginCode(opt.File)
		if err != nil {
			log.Fatal("Failed to read user defined plugin\n", err)
		}

		// babel transpile
		code, err := cmp.Transform(rawCode)
		if err != nil {
			log.Fatal("Failed to : ", err)
		}

		// creating config for each rule because we also want rule name of each score and report setter
		runCfg := &compiler.RunConfig{
			Type:      apiType,
			ApiSchema: apiSchemaFile,
			SetScore: func(category string, score float32) {
				// all other ones are invalid
				if category != "performance" && category != "security" && category != "quality" {
					logger.Error(fmt.Sprintf("%s gave invalid category - %s", rule, category))
					os.Exit(0)
				}
				rm.SetScore(rule, reportmanager.Score{Category: category, Value: score})
			},
			Report: func(body *reportmanager.ReportDef) {
				if body.Message == "" {
					logger.Error(fmt.Sprintf("%s didn't give message for report", rule))
					os.Exit(0)
				}
				rm.PushReport(rule, *body)
			},
		}

		// execute the code
		err = cmp.Run(code, runCfg, opt.Options)
		if err != nil {
			if errors.Is(err, compiler.ErrExceptionInPluginCode) {
				logger.Error(fmt.Sprintf("%s threw an exception", rule))
				logger.Error(err.Error())
				continue
			} else {
				log.Fatal("Failed to: ", err)
			}
		}
		logger.Info(fmt.Sprintf("%s check completed", rule))
		rulesPassedCounter++
	}

	if exportReportPath != "" {
		logger.Info(fmt.Sprintf("Exporting reports to %s", exportReportPath))
		expData := reportExportData{
			Metrics: &reportRuleMetrics{
				TotalRules:  totalRules,
				PassedRules: rulesPassedCounter,
			},
			RuleReport: &rm,
		}
		if err := fr.SaveFile(exportReportPath, &expData); err != nil {
			log.Fatal("Failed to export report\n", err)
		}
	}

	logger.RuleMetrics(rulesPassedCounter, totalRules)

	logger.Title("Reports")
	for rule, r := range rm {
		for _, report := range r.Reports {
			logger.Report(rule, report.Method, report.Path, report.Message)
			logger.Divider()
		}
	}

	logger.Title("Score Card")
	scores := rm.GetTotalScore()
	logger.ScoreCard(scores)

}
