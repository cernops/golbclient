package lbalias

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

const LEMON_CLI = "/usr/sbin/lemon-cli"


type MetricEntry struct {
	PrefixOp, Metric, Op string
	Index                int
	Value                float64
}

func checkLemon(lbalias *LBalias, line string) bool {

	lbalias.DebugMessage("[add_lemon_check] Adding Lemon metric ", line)

	actions, _ := regexp.Compile("(?i)^CHECK LEMON ([-])?(_)?([0-9]+)(:[0-9]+)?([^0-9]+)([0-9]*.?[0-9]*)")

	found := actions.FindStringSubmatch(line)
	if len(found) > 0 {
		prefix_op, underscore, metric, index, op, value := found[1], found[2], found[3], found[4], found[5], found[6]
		if underscore != "_" {
			fmt.Printf("[add_lemon_check] Invalid metric.  Must start with _ (", found[0], ")")
			return true
		}
		if index == "" {
			index = "1"
		}
		lbalias.DebugMessage("[add_lemon_check] prefix=", prefix_op, ", metric=", metric, ", index=", index, ", op=", op, ", value=", value)
		indexInt, err := strconv.Atoi(index)
		if err != nil {
			fmt.Printf("Error converting the index ", index, " to a string")
			return true
		}
		valueFloat, err := strconv.ParseFloat(value, 64)
		lbalias.CheckMetricList = append(lbalias.CheckMetricList, MetricEntry{prefix_op, metric, op, indexInt, valueFloat})
		//lbalias.MetricList= append(lbalias.MetricList, "")
	} else {
		fmt.Printf("[add_lemon_check] Invalid expresion: ", line)
		return true
	}
	return false
}

func (lbalias *LBalias) checkLemonMetric() bool {

	var commandArgs = []string{"--script", "-m", ""}
	for _, m := range lbalias.CheckMetricList {
		commandArgs[2] += m.Metric + " "
	}
	lbalias.DebugMessage("Running ", commandArgs)
	out, err := exec.Command(LEMON_CLI, commandArgs...).Output()

	if err != nil {
		fmt.Println("Error executing the lemon cli!", err)
		//return true

	}

	valuelist := map[string]string{}
	//Let's parse the output, and get the status of each metric
	for _, line := range strings.Split(string(out), "\n") {
		if len(line) == 0 {
			continue
		}
		result := strings.SplitN(line, " ", 4)
		valuelist[result[1]] = strings.Join(result[3:], " ")
	}
	lbalias.DebugMessage("[lemon metric] ", valuelist)

	//And now, going through the metrics, let's see if the formula holds

	for _, m := range lbalias.CheckMetricList {
		if val, ok := valuelist[m.Metric]; ok {
			values := strings.Split(val, " ")
			position := m.Index - 1
			if len(values) >= position {
				value, err := strconv.ParseFloat(values[position], 64)
				if err != nil {
					fmt.Println("Error converting the string to float")
				}
				if m.PrefixOp == "-" {
					value = -value
				}
				lbalias.DebugMessage("Lemon Metric ", m.Metric, " value ", value)
				lbalias.DebugMessage("Compare ", m.Op, " ", value, " with limit ", m.Value)
				var result bool
				switch m.Op {
				case "=":
					result = value == m.Value
				case "==":
					result = value == m.Value
				case "!=":
					result = value != m.Value
				case "<>":
					result = value != m.Value
				case "<=":
					result = value <= m.Value
				case "<":
					result = value < m.Value
				case ">=":
					result = value >= m.Value
				case ">":
					result = value > m.Value
				default:
					fmt.Println("Don't now!")
				}
				if !result {
					lbalias.DebugMessage("The comparison failed!")
					return true
				}

			} else {
				fmt.Println("The metric ", m.Metric, " is supposed to create at least ", position, " values")
				return true
			}

		} else {
			fmt.Println("The metric ", m.Metric, " does not have a value!")
			return true
		}

	}

	return false
}
