package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/checkmarxDev/ast-cli/internal/wrappers"

	commands "github.com/checkmarxDev/ast-cli/internal/commands"
	params "github.com/checkmarxDev/ast-cli/internal/params"
	"github.com/spf13/viper"
)

const (
	successfulExitCode = 0
	failureExitCode    = 1
)

func main() {
	err := bindKeyToEnvAndDefault(params.AstURIKey, params.AstURIEnv, "http://localhost:80")
	exitIfError(err)
	ast := viper.GetString(params.AstURIKey)

	err = bindKeyToEnvAndDefault(params.ScansPathKey, params.ScansPathEnv, "api/scans")
	exitIfError(err)
	scans := viper.GetString(params.ScansPathKey)

	err = bindKeyToEnvAndDefault(params.ProjectsPathKey, params.ProjectsPathEnv, "api/projects")
	exitIfError(err)
	projects := viper.GetString(params.ProjectsPathKey)

	err = bindKeyToEnvAndDefault(params.ResultsPathKey, params.ResultsPathEnv, "api/results")
	exitIfError(err)
	results := viper.GetString(params.ResultsPathKey)

	err = bindKeyToEnvAndDefault(params.BflPathKey, params.BflPathEnv, "api/bfl")
	exitIfError(err)
	bfl := viper.GetString(params.BflPathKey)

	err = bindKeyToEnvAndDefault(params.UploadsPathKey, params.UploadsPathEnv, "api/uploads")
	exitIfError(err)
	uploads := viper.GetString(params.UploadsPathKey)

	sastRmPathKey := strings.ToLower(params.SastRmPathEnv)
	err = bindKeyToEnvAndDefault(sastRmPathKey, params.SastRmPathEnv, "api/sast-rm")
	exitIfError(err)
	sastrm := viper.GetString(sastRmPathKey)

	err = bindKeyToEnvAndDefault(params.AstWebAppHealthCheckPathKey, params.AstWebAppHealthCheckPathEnv, "#/projects")
	exitIfError(err)
	webAppHlthChk := viper.GetString(params.AstWebAppHealthCheckPathKey)

	err = bindKeyToEnvAndDefault(params.HealthcheckPathKey, params.HealthcheckPathEnv, "api/healthcheck")
	exitIfError(err)
	healthcheck := viper.GetString(params.HealthcheckPathKey)

	err = bindKeyToEnvAndDefault(params.HealthcheckDBPathKey, params.HealthcheckDBPathEnv, "database")
	exitIfError(err)
	healthcheckDBPath := viper.GetString(params.HealthcheckDBPathKey)

	// TODO change nats to message-queue
	// TODO change minio to object-store

	err = bindKeyToEnvAndDefault(params.HealthcheckNatsPathKey, params.HealthcheckNatsPathEnv, "nats")
	exitIfError(err)
	healthcheckNatsPath := viper.GetString(params.HealthcheckNatsPathKey)

	err = bindKeyToEnvAndDefault(params.HealthcheckMinioPathKey, params.HealthcheckMinioPathEnv, "minio")
	exitIfError(err)
	healthcheckMinioPath := viper.GetString(params.HealthcheckMinioPathKey)

	// Change redis to TBD
	err = bindKeyToEnvAndDefault(params.HealthcheckRedisPathKey, params.HealthcheckRedisPathEnv, "redis")
	exitIfError(err)
	healthcheckRedisPath := viper.GetString(params.HealthcheckRedisPathKey)

	err = bindKeyToEnvAndDefault(params.AccessKeyIDConfigKey, params.AccessKeyIDEnv, "")
	exitIfError(err)

	err = bindKeyToEnvAndDefault(params.AccessKeySecretConfigKey, params.AccessKeySecretEnv, "")
	exitIfError(err)

	err = bindKeyToEnvAndDefault(params.AstAuthenticationURIConfigKey, params.AstAuthenticationURIEnv, "")
	exitIfError(err)

	err = bindKeyToEnvAndDefault(params.AstRoleKey, params.AstRoleEnv, params.ScaAgent)
	exitIfError(err)

	err = bindKeyToEnvAndDefault(params.CredentialsFilePathKey, params.CredentialsFilePathEnv, "credentials.ast")
	exitIfError(err)

	err = bindKeyToEnvAndDefault(params.TokenExpirySecondsKey, params.TokenExpirySecondsEnv, "300")
	exitIfError(err)

	scansURL := fmt.Sprintf("%s/%s", ast, scans)
	uploadsURL := fmt.Sprintf("%s/%s", ast, uploads)
	projectsURL := fmt.Sprintf("%s/%s", ast, projects)
	resultsURL := fmt.Sprintf("%s/%s", ast, results)
	sastrmURL := fmt.Sprintf("%s/%s", ast, sastrm)
	bflURL := fmt.Sprintf("%s/%s", ast, bfl)
	webAppHlthChkURL := fmt.Sprintf("%s/%s", ast, webAppHlthChk)
	hlthChekURL := fmt.Sprintf("%s/%s", ast, healthcheck)
	healthcheckDBURL := fmt.Sprintf("%s/%s", hlthChekURL, healthcheckDBPath)
	healthcheckNatsURL := fmt.Sprintf("%s/%s", hlthChekURL, healthcheckNatsPath)
	healthcheckMinioURL := fmt.Sprintf("%s/%s", hlthChekURL, healthcheckMinioPath)
	healthcheckRedisURL := fmt.Sprintf("%s/%s", hlthChekURL, healthcheckRedisPath)

	scansWrapper := wrappers.NewHTTPScansWrapper(scansURL)
	uploadsWrapper := wrappers.NewUploadsHTTPWrapper(uploadsURL)
	projectsWrapper := wrappers.NewHTTPProjectsWrapper(projectsURL)
	resultsWrapper := wrappers.NewHTTPResultsWrapper(resultsURL)
	bflWrapper := wrappers.NewHTTPBFLWrapper(bflURL)
	rmWrapper := wrappers.NewSastRmHTTPWrapper(sastrmURL)
	healthCheckWrapper := wrappers.NewHealthCheckHTTPWrapper(
		webAppHlthChkURL,
		healthcheckDBURL,
		healthcheckNatsURL,
		healthcheckMinioURL,
		healthcheckRedisURL,
	)
	defaultConfigFileLocation := "/etc/conf/cx/config.yml"

	astCli := commands.NewAstCLI(
		scansWrapper,
		uploadsWrapper,
		projectsWrapper,
		resultsWrapper,
		bflWrapper,
		rmWrapper,
		healthCheckWrapper,
		defaultConfigFileLocation,
	)

	err = astCli.Execute()
	exitIfError(err)
	os.Exit(successfulExitCode)
}

func exitIfError(err error) {
	if err != nil {
		os.Exit(failureExitCode)
	}
}

func bindKeyToEnvAndDefault(key, env, defaultVal string) error {
	err := viper.BindEnv(key, env)
	viper.SetDefault(key, defaultVal)
	return err
}
