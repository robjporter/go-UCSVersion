package app

import (
	"fmt"
	functions2 "github.com/robjporter/go-functions"
	"github.com/robjporter/go-functions/as"
	"github.com/robjporter/go-functions/banner"
	"github.com/robjporter/go-functions/cisco/ucs"
	"github.com/robjporter/go-functions/colors"
	"github.com/robjporter/go-functions/lfshook"
	"github.com/robjporter/go-functions/logrus"
	"github.com/robjporter/go-functions/terminal"
	"github.com/robjporter/go-functions/viper"
	"github.com/robjporter/go-functions/yaml"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

var (
	Core Application
)

func (a *Application) runAll() {
	a.getAllSystems()
	if len(a.UCS) > 0 {
		for i := 0; i < len(a.UCS); i++ {
			a.LogInfo("Attempting to connect to UCS System", map[string]interface{}{"System": a.UCS[i].ip}, false)
			myucs := ucs.New()
			myucs.Login(a.UCS[i].ip, a.UCS[i].username, a.DecryptPassword(a.UCS[i].password))
			if myucs.LastResponse.Errors == nil {
				a.UCS[i].version = myucs.GetVersion()
				a.UCS[i].status = true
				a.LogInfo("Successfully connected to UCS System", map[string]interface{}{"System": a.UCS[i].ip}, false)
			} else {
				a.UCS[i].status = false
				a.LogInfo("Failed to connect to UCS System", map[string]interface{}{"System": a.UCS[i].ip}, false)
			}
			myucs.Logout()
		}
	} else {
		fmt.Println("No UCS Systems detected in the config file.  Please trying adding one and try again.")
	}
	a.processVersions()
	a.outputVersionSuggestions()
}

func (a *Application) outputVersionSuggestions() {
	a.LogInfo("Displaying UCS Version information", nil, false)
	table := terminal.New([]string{"UCS Domain", "Current Version", "Is Deferred", "Suggested Version"})
	for i := 0; i < len(a.UCS); i++ {
		if a.UCS[i].status != false {
			table.AddRow(map[string]interface{}{"UCS Domain": a.UCS[i].ip, "Current Version": a.UCS[i].version, "Is Deferred": a.UCS[i].deferredVersion, "Suggested Version": a.UCS[i].suggestedVersion})
		}
	}
	table.Print()
}

func (a *Application) processVersions() {
	a.LogInfo("Getting UCS Version information from Cisco", nil, false)
	ucs.GetWebData()
	a.LogInfo("Processing discovered UCS Manager versions", nil, false)
	for i := 0; i < len(a.UCS); i++ {
		if a.UCS[i].status != false {
			a.LogInfo("Validating UCS Version Information", map[string]interface{}{"Version": a.UCS[i].version}, false)
			a.UCS[i].suggestedVersion = ucs.GetSuggestedReleaseTrain(strings.TrimSpace(a.UCS[i].version))
			a.LogInfo("Validating if deferred Version", map[string]interface{}{"Version": a.UCS[i].version}, false)
			a.UCS[i].deferredVersion = ucs.GetIsDeferredRelease(strings.TrimSpace(a.UCS[i].version))
		}
	}
}

func (a *Application) displayBanner() {
	terminal.ClearScreen()
	banner.PrintNewFigure("UCS Version", "rounded", true)
	fmt.Println(colors.Color("Cisco Unified Computing System Version checking tool v"+a.Version, colors.BRIGHTYELLOW))
	banner.BannerPrintLineS("=", 80)
}

func (a *Application) Run() {
	a.LogInfo("Application", map[string]interface{}{"Version": a.Version}, false)
	a.LogInfo("Starting main application Run stage 1", nil, false)
	runtime.GOMAXPROCS(runtime.NumCPU())
	a.processResponse(ProcessCommandLineArguments())
}

func (a *Application) saveConfig() {
	a.LogInfo("Saving configuration file.", nil, false)
	if len(a.UCS) > 0 {
		items := a.processSystems()
		a.Config.Set("ucs.systems", items)
	}
	out, err := yaml.Marshal(a.Config.AllSettings())
	if err == nil {
		fp, err := os.Create(a.ConfigFile)
		if err == nil {
			defer fp.Close()
			_, err = fp.Write(out)
		}
	}
	a.Log("Saving configuration file complete.", nil, true)
}

func (a *Application) processSystems() []interface{} {
	var items []interface{}
	var item map[string]interface{}
	for i := 0; i < len(a.UCS); i++ {

		item = make(map[string]interface{})
		item["url"] = a.UCS[i].ip
		item["username"] = a.UCS[i].username
		item["password"] = a.UCS[i].password
		items = append(items, item)
	}
	return items
}

func (a *Application) init() {
	a.Config = viper.New()
	a.Logger = logrus.New()
	a.Logger.Level = logrus.DebugLevel
	customFormatter := new(logrus.TextFormatter)
	customFormatter.TimestampFormat = "02-01-2006 15:04:05.000"
	customFormatter.FullTimestamp = true
	a.Logger.Formatter = customFormatter
	a.Logger.Out = os.Stdout
	timestamp := as.ToString(time.Now().Unix())
	path := "./logs/" + timestamp + "/"

	os.Mkdir("logs", 0700)
	os.Mkdir(path, 0700)

	a.Logger.Hooks.Add(lfshook.NewHook(lfshook.PathMap{
		logrus.InfoLevel:  path + "info-" + timestamp + ".log",
		logrus.ErrorLevel: path + "error-" + timestamp + ".log",
		logrus.WarnLevel:  path + "warn-" + timestamp + ".log",
		logrus.DebugLevel: path + "debug-" + timestamp + ".log",
		logrus.FatalLevel: path + "fatal-" + timestamp + ".log",
	}))
	a.Key = []byte("random123456")
	a.displayBanner()
}

func (a *Application) LoadConfig() {
	a.init()
	a.Log("Loading Configuration File.", nil, true)
	configName := ""
	configExtension := ""
	configPath := ""

	splits := strings.Split(filepath.Base(a.ConfigFile), ".")
	if len(splits) == 2 {
		configName = splits[0]
		configExtension = splits[1]
	}
	configPath = filepath.Dir(a.ConfigFile)

	a.Config.SetConfigName(configName)
	a.Config.SetConfigType(configExtension)
	a.Config.AddConfigPath(configPath)

	a.Log("Configuration File defined", map[string]interface{}{"Path": configPath, "Name": configName, "Extension": configExtension}, true)

	if functions2.Exists(a.ConfigFile) {
		err := a.Config.ReadInConfig()
		if err != nil {
			panic(fmt.Errorf("Fatal error config file: %s \n", err))
			os.Exit(0)
		}
		a.Log("Configuration File read successfully.", nil, true)
	} else {
		a.LogInfo("Configuration File not found.", nil, true)
	}
}

func (a *Application) LogInfo(message string, fields map[string]interface{}, infoMessage bool) {
	if infoMessage && a.Debug || !infoMessage {
		if fields != nil {
			a.Logger.WithFields(fields).Info(message)
		} else {
			a.Logger.Info(message)
		}
	}
}

func (a *Application) Log(message string, fields map[string]interface{}, debugMessage bool) {
	if debugMessage && a.Debug || !debugMessage {
		if fields != nil {
			a.Logger.WithFields(fields).Info(message)
		} else {
			a.Logger.Info(message)
		}
	}
}

func (a *Application) EncryptPassword(password string) string {
	return functions2.Encrypt(a.Key, []byte(password))
}

func (a *Application) DecryptPassword(password string) string {
	return functions2.Decrypt(a.Key, password)
}
