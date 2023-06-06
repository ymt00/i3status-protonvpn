// Integration of protonvpn python script for i3status bar for sway
// show a Bemenu for connection to servers (fastest, Japan, Netherland, US)
// if already connected add Disconnect and Refresh to Bemenu
package main

import (
	"fmt"
	"i3status/utils"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

const (
	protonvpnBin = "protonvpn"

	statusWorking      = "working"
	statusConnected    = "connected"
	statusDisconnected = "disconnected"
	statusError        = "error"
)

// MenuItem struct
type MenuItem struct {
	Name   string
	Action string
}

var (
	statusPath string

	menuItems = []MenuItem{
		{"Le plus rapide", "connect -f"},
		{"Japon", "connect --cc jp"},
		{"Pays-Bas", "connect --cc nl"},
		{"Etats-Unis", "connect --cc us"},
		{"Rafraîchir", "refresh"},
		{"Déconnecter", "disconnect"},
	}

	i3status = map[string]string{
		statusWorking:      "{\"state\": \"Warning\", \"text\": \"\\uf2f1 %s\\uf023\"}",
		statusConnected:    "{\"state\": \"Good\", \"text\": \"%s \\uf023\"}",
		statusDisconnected: "{\"state\": \"Info\", \"text\": \"%s\\uf09c\"}",
		statusError:        "{\"state\": \"Critical\", \"text\": \"\\uf071 %s \\uf09c\"}",
	}

	protonStatus = map[string]string{
		"Warning":  statusWorking,
		"Good":     statusConnected,
		"Info":     statusDisconnected,
		"Critical": statusError,
	}
)

// readStatus function read Protonvpn current status from file /home/yves/.config/i3status/protonvpn_status.conf
func readStatus() string {
	data, err := os.ReadFile(statusPath)

	// when start, the ProtonVPN status file might not exist
	// so we need to create it and set its status
	if err != nil {
		// if the file do not exist, create it and set initial status
		if os.IsNotExist(err) {
			if _, err = os.Create(statusPath); err != nil {
				panic("Could not create ProtonVPN status file.")
			}
			// write ProtonVPN status to status file
			setStatus(handleActionOutput(""))

			// read again the newly created status
			data, _ = os.ReadFile(statusPath)
		} else {
			panic("Something goes wrong reading ProtonVPN status file.")
		}
	}

	if string(data) == "" {
		setStatus(handleActionOutput(""))
	}

	return string(data)
}

// checkBinary verify protonvpn binary existnce
func checkBinary() bool {
	if _, err := exec.LookPath(protonvpnBin); err != nil {
		return false
	}
	return true
}

// getStatus get the status of protonvpn
func getStatus() string {
	matches := regexp.
		MustCompile(`"state": "([A-Za-z]*)"`).
		FindStringSubmatch(readStatus())

	if len(matches) > 1 {
		return protonStatus[matches[1]]
	}

	return statusError
}

// setStatus set the status of protonvpn
func setStatus(status string, msg string) {
	os.WriteFile(statusPath, []byte(fmt.Sprintf(i3status[status], msg)), 0644)
}

// protonVPNStatus return output of protonvpn status command
func protonVPNStatus() string {
	cmd := exec.Command(protonvpnBin, "status")
	out, err := cmd.Output()

	if err != nil {
		return statusError
	}

	return string(out)
}

// handleActionOutput return the state of protonvpn
func handleActionOutput(output string) (string, string) {
	if output == "" {
		output = protonVPNStatus()
	}

	// fmt.Printf(
	// 	"--------------------------------\nOutput command:\n%s\n--------------------------------\n",
	// 	output)

	status := statusDisconnected

	matched, _ := regexp.MatchString(`Connected`, output)

	if matched {

		re := regexp.MustCompile(`Connecting to (.*) via|Server:[[:space:]]*(.*)`)
		matches := re.FindStringSubmatch(output)

		if len(matches) > 1 {
			server := matches[1]

			if server == "" && len(matches) > 2 {
				server = matches[2]
			}

			return statusConnected, server
		}

		return statusConnected, "There has been an error."
	}

	matched, _ = regexp.MatchString(`Disconnected|No connection found`, output)

	if matched {
		return statusDisconnected, ""
	}

	matched, _ = regexp.MatchString(`error`, output)

	if matched {
		return statusError, ""
	}

	return status, ""
}

func findAction(name string) string {
	for _, item := range menuItems {
		if name == item.Name {
			return item.Action
		}
	}
	return ""
}

func action(name string) {
	setStatus(statusWorking, "")

	args := strings.Split(protonvpnBin+" "+findAction(name), " ")

	cmd := exec.Command("sudo", args...)

	output, err := cmd.Output()
	if err != nil {
		setStatus(statusError, "There has been an error.")
	} else {
		setStatus(handleActionOutput(string(output)))
	}
}

func main() {
	if len(os.Args) < 2 {
		panic("You must provide ProtonVPN status path arguemnt.")
	}

	if checkBinary() {
		// set ProtonVPN status file path
		statusPath = os.Args[1]

		if utils.IsAppRunning("protonvpn") {
			utils.ScratchpadShow("protonvpn")
		} else {
			s := getStatus()

			if s != statusWorking {
				var bemenuItems []string

				menuItemsChoices := menuItems[0:4]

				if s == statusConnected {
					menuItemsChoices = menuItems
				}

				for _, choice := range menuItemsChoices {
					bemenuItems = append(bemenuItems, choice.Name)
				}

				sel := utils.Bemenu(bemenuItems, []string{"--prompt", "\uf21b ProtonVPN"}...)
				if sel != "" {
					action(sel)
				}
			}
		}

	} else {
		setStatus(statusError, "Protonvpn not in your path or not installed.")
	}
}
