package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"strings"

	"gopkg.in/yaml.v2"
)

const (
	TestAuth      = "[Authentication]"
	TestCreation  = "[Creation]"
	TestStability = "[Stability]"
	TestForbidden = "[Forbidden]"
)

type Config struct {
	GlobalVariables map[string]string `yaml:"globalVariables"`
	Tests           []Test            `yaml:"testsForDockerContainerAuthPlugin"`
}

type Test struct {
	Number int    `yaml:"number"`
	Name   string `yaml:"name"`
	Value  string `yaml:"value"`
	Result string `yaml:"result"`
}

func PrintLines() {
	fmt.Println("-----------------------------------")
}

func Performer(command string) string {

	args := strings.Fields(command)
	cmd := exec.Command(args[0], args[1:]...)

	// Получение вывода команды и ошибки
	output, _ := cmd.CombinedOutput()
	return string(output)
}

func Testing(caseOfTest Test, containerid string) []string {
	lines := strings.Split(caseOfTest.Value, "\n")
	var results []string
	for _, i := range lines {
		if i == "" {
			continue
		}
		var d string
		if strings.Contains(i, "%s") {
			d = fmt.Sprintf(i, containerid)
		} else {
			d = i
		}
		result := Performer(d)
		result = strings.TrimRight(result, "\n")
		res := strings.Split(result, "\n")
		if len(res) > 1 {
			for index, info := range res {
				if strings.Contains(info, "--help") {
					res = append(res[:index], res[index+1:]...)
				}
			}
			results = append(results, res...)
		} else {
			results = append(results, result)
		}
	}
	return results
}

func arraysEqual(arr1, arr2 []string) bool {
	if len(arr1) != len(arr2) {
		return false
	}

	for i := 0; i < len(arr1); i++ {
		if arr1[i] != arr2[i] {
			return false
		}
	}
	return true
}

func findNotEqualArrays(arrays [][]string) (int, int) {
	for i := 0; i < len(arrays); i++ {
		for j := i + 1; j < len(arrays); j++ {
			if !arraysEqual(arrays[i], arrays[j]) {
				return i, j
			}
		}
	}
	return -1, -1
}

func main() {

	yfile, err := ioutil.ReadFile("tests/testCases.yaml")
	if err != nil {
		log.Fatalf("cannot open: %v", err)
	}

	var config Config
	err = yaml.Unmarshal(yfile, &config)
	if err != nil {
		log.Fatalf("cannot unmarshal data: %v", err)
	}

	containerid := config.GlobalVariables["ContainerID"]
	nameOfContainer := config.GlobalVariables["NameOfContainer"]
	fullContainerID := config.GlobalVariables["FullContainerID"]

	for _, test := range config.Tests {

		var wait string
		var results []string

		if strings.HasPrefix(test.Name, TestAuth) {
			wait = fmt.Sprintf(test.Result, config.GlobalVariables["ErrorAuthMessage"])
			wait = strings.TrimRight(wait, "\n")

			results = Testing(test, containerid)
			resultsOfCallnameOfContainer := Testing(test, nameOfContainer)
			resultsOfCallFullContainerID := Testing(test, fullContainerID)
			resultsOfCallLessThan12 := Testing(test, fullContainerID[:8])
			resultsOfCallLessThan64 := Testing(test, fullContainerID[:32])

			arrays := [][]string{results, resultsOfCallnameOfContainer,
				resultsOfCallFullContainerID, resultsOfCallLessThan12, resultsOfCallLessThan64}

			i, j := findNotEqualArrays(arrays)

			if i != -1 && j != -1 {
				fmt.Printf("Массивы %s \n\tи %s не равны.\n", arrays[i+1], arrays[j+1])
			}
		}

		if strings.HasPrefix(test.Name, TestCreation) {
			template := strings.TrimRight(config.GlobalVariables["ErrorCreationMessage"], "\n")
			wait = fmt.Sprintf(test.Result, template)
			wait = strings.TrimRight(wait, "\n")
			results = Testing(test, containerid)
		}

		if strings.HasPrefix(test.Name, TestStability) {
			template := strings.TrimRight(config.GlobalVariables["WhatIdontExpectFromStabilitisTests"], "\n")
			wait = fmt.Sprintf(test.Result, template)
			wait = strings.TrimRight(wait, "\n")
			results = Testing(test, containerid)
		}

		if strings.HasPrefix(test.Name, TestForbidden) {
			template := strings.TrimRight(config.GlobalVariables["ForbiddenMessage"], "\n")
			wait = fmt.Sprintf(test.Result, template)
			wait = strings.TrimRight(wait, "\n")
			results = Testing(test, containerid)
		}

		testIsOK := false
		isItTestInStability := strings.Contains(test.Name, TestStability)
		for i, _ := range results {
			// forbbiden make this if with contains
			if results[i] == wait {
				if isItTestInStability {
					PrintLines()
					msg := fmt.Sprintf("Test number #%d FAILED! I've got: %s, \n\tbut expect: %s", test.Number, results, wait)
					fmt.Println(msg)
					testIsOK = false
					break
				}
				PrintLines()
				msg := fmt.Sprintf("Test number #%d OK!", test.Number)
				fmt.Println(msg)
				testIsOK = true
				break
			}
		}
		if !testIsOK {
			if isItTestInStability {
				PrintLines()
				msg := fmt.Sprintf("Test number #%d OK!", test.Number)
				fmt.Println(msg)
				continue
			}
			PrintLines()
			msg := fmt.Sprintf("Test number #%d FAILED! I've got: %s, \n\tbut expect: %s", test.Number, results, wait)
			fmt.Println(msg)
		}
	}
}
