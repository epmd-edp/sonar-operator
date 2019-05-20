package client

import (
	"gotest.tools/assert"
	"log"
	"testing"
)

const (
	url       = "https://sonar-mr-1944-1-edp-cicd.delivery.aws.main.edp.projects.epam.com/api"
	username  = "admin"
	groupName = "non-interactive-users"
	userName  = "jenkins"
	token     = ""
)

func TestExampleConfiguration_checkProfileExist(t *testing.T) {
	cs := SonarClient{}
	err := cs.InitNewRestClient(url, username, token)
	if err != nil {
		log.Print(err)
	}

	exist, result, err := cs.checkProfileExist()
	if err != nil {
		log.Print(err)
	}

	log.Println(result, exist)
}

func TestExampleConfiguration_UploadProfile(t *testing.T) {
	cs := SonarClient{}
	err := cs.InitNewRestClient(url, username, token)
	if err != nil {
		log.Print(err)
	}

	id, err := cs.UploadProfile()
	if err != nil {
		log.Print(err)
	}

	log.Println(*id)
}

func TestExampleConfiguration_checkInstallPlugins(t *testing.T) {
	sc := SonarClient{}
	err := sc.InitNewRestClient(url, username, token)
	if err != nil {
		log.Print(err)
	}

	plugins := []string{"pmd"}
	err = sc.InstallPlugins(plugins)
	if err != nil {
		log.Print(err)
	}
}

func TestExampleConfiguration_checkGroupExist(t *testing.T) {
	cs := SonarClient{}
	err := cs.InitNewRestClient(url, username, token)
	if err != nil {
		log.Print(err)
	}

	exist, err := cs.checkGroupExist(groupName)
	if err != nil {
		log.Print(err)
	}
	log.Println(exist)

	assert.Equal(t, exist, true)
}

func TestExampleConfiguration_CreateGroup(t *testing.T) {
	cs := SonarClient{}
	err := cs.InitNewRestClient(url, username, token)
	if err != nil {
		log.Print(err)
	}

	err = cs.CreateGroup(groupName)
	if err != nil {
		log.Print(err)
	}
}

func TestExampleConfiguration_AddUserToGroup(t *testing.T) {
	cs := SonarClient{}
	err := cs.InitNewRestClient(url, username, token)
	if err != nil {
		log.Print(err)
	}

	err = cs.AddUserToGroup(groupName, "jenkins")
	if err != nil {
		log.Print(err)
	}
}

func TestExampleConfiguration_AddPermissionsToUser(t *testing.T) {
	cs := SonarClient{}
	err := cs.InitNewRestClient(url, username, token)
	if err != nil {
		log.Print(err)
	}

	err = cs.AddPermissionsToUser(userName, "admin")
	if err != nil {
		log.Print(err)
	}
}

func TestExampleConfiguration_AddPermissionsToGroup(t *testing.T) {
	cs := SonarClient{}
	err := cs.InitNewRestClient(url, username, token)
	if err != nil {
		log.Print(err)
	}

	err = cs.AddPermissionsToGroup(groupName, "scan")
	if err != nil {
		log.Print(err)
	}
}
