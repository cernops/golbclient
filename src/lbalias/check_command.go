package lbalias

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

func checkCommand(lbalias *LBalias, line string) interface{} {
	command, _ := regexp.Compile("(?i)^CHECK command[ ]*([^ ]+)[ ]*(.*)")

	found := command.FindStringSubmatch(line)

	if len(found) > 0 {
		args := []string{}
		if found[2] != "" {
			args = strings.Split(found[2], " ")
		}
		lbalias.DebugMessage("[check_command] Running '", found[1], args, "'")
		out, err := exec.Command(found[1], args...).Output()

		if err != nil {
			fmt.Println(err)
			rc := err.(*exec.ExitError)
			fmt.Println("[check_command] exception catched: ", found[1], found[2], ". Ignoring script return code", err)
			fmt.Println("[check_command] return code: ", rc)
			return true

		}
		lbalias.DebugMessage("[check_command] output", string(out))
		lbalias.DebugMessage("[check_command] return code: 0 ")
		return false
	}
	return true
}
