package app

import (
	"github.com/robjporter/go-functions/as"
	"github.com/robjporter/go-functions/kingpin"
	"strings"
)

var (
	add       = kingpin.Command("add", "Register a new UCS domain.")
	update    = kingpin.Command("update", "Update a UCS domain.")
	delete    = kingpin.Command("delete", "Remove a UCS domain.")
	show      = kingpin.Command("show", "Show a UCS domain.")
	run       = kingpin.Command("run", "Run the main application.")
	addUCS    = add.Command("ucs", "Add a UCS Domain")
	updateUCS = update.Command("ucs", "Update a UCS Domain")
	deleteUCS = delete.Command("ucs", "Delete a UCS Domain")
	showUCS   = show.Command("ucs", "Show a UCS Domain")

	showAll = show.Command("all", "Show all")

	addUCSIP       = addUCS.Flag("ip", "IP Address or DNS name for UCS Manager, without http(s).").Required().IP()
	addUCSUsername = addUCS.Flag("username", "Name of user.").Required().String()
	addUCSPassword = addUCS.Flag("password", "Password for user in plain text.").Required().String()

	updateUCSIP       = updateUCS.Flag("ip", "IP Address or DNS name for UCS Manager, without http(s).").Required().IP()
	updateUCSUsername = updateUCS.Flag("username", "Name of user.").Required().String()
	updateUCSPassword = updateUCS.Flag("password", "Password for user in plain text.").Required().String()

	deleteUCSIP = deleteUCS.Flag("ip", "IP Address or DNS name for UCS Manager, without http(s).").Required().IP()

	showUCSIP = showUCS.Flag("ip", "IP Address or DNS name for UCS Manager, without http(s).").Required().IP()
)

func ProcessCommandLineArguments() string {
	switch kingpin.Parse() {
	case "run":
		return "RUN"
	case "add ucs":
		return "ADDUCS|" + as.ToString(*addUCSIP) + "|" + *addUCSUsername + "|" + *addUCSPassword
	case "update ucs":
		return "UPDATEUCS|" + as.ToString(*updateUCSIP) + "|" + *updateUCSUsername + "|" + *updateUCSPassword
	case "delete ucs":
		return "DELETEUCS|" + as.ToString(*deleteUCSIP)
	case "show ucs":
		return "SHOWUCS|" + as.ToString(*showUCSIP)
	case "show all":
		return "SHOWALL"
	}
	return ""
}

func (a *Application) processResponse(response string) {
	a.Log("Processing command line options.", map[string]interface{}{"args": response}, true)
	splits := strings.Split(response, "|")
	switch splits[0] {
	case "RUN":
		a.runAll()
	case "ADDUCS":
		a.addUCSSystem(splits[1], splits[2], splits[3])
	case "UPDATEUCS":
		a.updateUCSSystem(splits[1], splits[2], splits[3])
	case "DELETEUCS":
		a.deleteUCSSystem(splits[1])
	case "SHOWUCS":
		a.showUCSSystem(splits[1])
	case "SHOWALL":
		a.showUCSSystems()

	}
}

func (a *Application) addUCSSystem(ip, username, password string) {
	if !a.checkUCSExists(ip) {
		if a.addUCS(ip, username, password) {
			a.saveConfig()
			a.LogInfo("New UCS system has been added successfully.", map[string]interface{}{"IP": ip, "Username": username}, false)
		} else {
			a.LogInfo("UCS System could not be added.", map[string]interface{}{"IP": ip, "Username": username}, false)
		}
	} else {
		a.LogInfo("A UCS System already exsists in the config file.", map[string]interface{}{"IP": ip, "Username": username}, false)
	}
}

func (a *Application) addUCS(ip, username, password string) bool {
	if ip != "" {
		if username != "" {
			if password != "" {
				tmp := UCSSystemInfo{}
				tmp.ip = ip
				tmp.username = username
				tmp.password = a.EncryptPassword(password)
				a.UCS = append(a.UCS, tmp)
				return true
			} else {
				a.Log("The password for the UCS System cannot be blank.", nil, false)
			}
		} else {
			a.Log("The username for the UCS System cannot be blank.", nil, false)
		}
	} else {
		a.Log("The URL for the UCS System cannot be blank.", nil, false)
	}
	return false
}

func (a *Application) deleteUCSSystem(ip string) {
	if a.checkUCSExists(ip) {
		if a.deleteUCS(ip) {
			a.saveConfig()
			a.LogInfo("UCS system has been deleted successfully.", map[string]interface{}{"IP": ip}, true)
		} else {
			a.Log("UCS System could not be deleted.", map[string]interface{}{"IP": ip}, false)
		}
	} else {
		a.LogInfo("UCS System does not exsists and so cannot be deleted.", map[string]interface{}{"IP": ip}, false)
	}
}

func (a *Application) deleteUCS(ip string) bool {
	for i := 0; i < len(a.UCS); i++ {
		if a.UCS[i].ip == as.ToString(ip) {
			a.UCS = append(a.UCS[:i], a.UCS[i+1:]...)
		}
	}
	return true
}

func (a *Application) showUCS(ip string) {
	for i := 0; i < len(a.UCS); i++ {
		if a.UCS[i].ip == as.ToString(ip) {
			a.LogInfo("UCS Domain", map[string]interface{}{"URL": a.UCS[i].ip}, false)
			a.LogInfo("UCS Domain", map[string]interface{}{"Username": a.UCS[i].username}, false)
			a.LogInfo("UCS Domain", map[string]interface{}{"Password": a.UCS[i].password}, false)
		}
	}
}

func (a *Application) showUCSSystem(ip string) {
	if a.checkUCSExists(ip) {
		a.showUCS(ip)
	} else {
		a.Log("The UCS Domain does not exist and so cannot be displayed.", map[string]interface{}{"URL": ip}, false)
	}
}

func (a *Application) showUCSSystems() {
	a.getAllSystems()
	for i := 0; i < len(a.UCS); i++ {
		a.LogInfo("UCS Domain", map[string]interface{}{"URL": a.UCS[i].ip}, false)
		a.LogInfo("UCS Domain", map[string]interface{}{"Username": a.UCS[i].username}, false)
		a.LogInfo("UCS Domain", map[string]interface{}{"Password": a.UCS[i].password}, false)
	}
}

func (a *Application) updateUCS(ip, username, password string) bool {
	for i := 0; i < len(a.UCS); i++ {
		if a.UCS[i].ip == as.ToString(ip) {
			a.UCS[i].username = username
			a.UCS[i].password = a.EncryptPassword(password)
		}
	}
	return true
}

func (a *Application) updateUCSSystem(ip, username, password string) {
	if a.checkUCSExists(ip) {
		if a.updateUCS(ip, username, password) {
			a.saveConfig()
			a.LogInfo("Update to UCS system has been completed successfully.", map[string]interface{}{"IP": ip, "Username": username}, false)
		} else {
			a.LogInfo("UCS System could not be updated.", map[string]interface{}{"IP": ip, "Username": username}, false)
		}
	} else {
		a.LogInfo("UCS System does not exsist and can therefore not be updated.", map[string]interface{}{"IP": ip, "Username": username}, false)
	}
}

func (a *Application) checkUCSExists(ip string) bool {
	a.Log("Search for UCS System in config file", map[string]interface{}{"IP": ip}, true)
	if a.Config.IsSet("ucs.systems") {
		a.getAllSystems()
		for i := 0; i < len(a.UCS); i++ {
			if strings.TrimSpace(a.UCS[i].ip) == strings.TrimSpace(ip) {
				return true
			}
		}
		return false
	}
	return false
}

func (a *Application) getAllSystems() {
	tmp := as.ToSlice(a.Config.Get("ucs.systems"))
	a.Log("Located UCS Systems in the config file", map[string]interface{}{"Systems": len(tmp)}, true)
	a.readSystems(tmp)
}

func (a *Application) readSystems(ucss []interface{}) bool {
	a.UCS = nil
	for i := 0; i < len(ucss); i++ {
		var newlist map[string]string
		newlist = as.ToStringMapString(ucss[i])
		tmp := UCSSystemInfo{}
		tmp.ip = newlist["url"]
		tmp.username = newlist["username"]
		tmp.password = newlist["password"]
		a.UCS = append(a.UCS, tmp)
	}
	return true
}
