package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/dchest/uniuri"
	"gopkg.in/resty.v1"
	"log"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sonar-operator/pkg/apis/edp/v1alpha1"
	sonarClient "sonar-operator/pkg/client"
	"time"
)

const (
	StatusInstall   = "installing"
	StatusFailed    = "failed"
	StatusCreated   = "created"
	JenkinsUsername = "jenkins"
	GroupName       = "non-interactive-users"
	WebhookUrl      = "http://jenkins:8080/sonarqube-webhook/"
)

type Client struct {
	client resty.Client
}

type SonarService interface {
	// This is an entry point for service package. Invoked in err = r.service.Install(*instance) sonar_controller.go, Reconcile method.
	Install(instance v1alpha1.Sonar) error
	Configure(instance v1alpha1.Sonar) error
}

func NewSonarService(platformService PlatformService, k8sClient client.Client) SonarService {
	return SonarServiceImpl{platformService: platformService, k8sClient: k8sClient}
}

type SonarServiceImpl struct {
	// Providing sonar service implementation through the interface (platform abstract)
	platformService PlatformService
	k8sClient       client.Client
}

func (s SonarServiceImpl) Configure(instance v1alpha1.Sonar) error {
	log.Println("Sonar component configuration has been started")
	sonarApiUrl := fmt.Sprintf("http://%v.%v:9000/api", instance.Name, instance.Namespace)

	sc := sonarClient.SonarClient{}
	err := sc.InitNewRestClient(sonarApiUrl, "admin", "admin")
	if err != nil {
		return logErrorAndReturn(err)
	}

	sc.WaitForStatusIsUp(60, 10)

	credentials := s.platformService.GetSecret(instance.Namespace, instance.Name+"-admin-password")
	if credentials == nil {
		logErrorAndReturn(errors.New("Sonar secret not found. Configuration failed"))
	}
	password := string(credentials["password"])

	err = sc.ChangePassword("admin", "admin", password)
	if err != nil {
		return logErrorAndReturn(err)
	}

	err = sc.InitNewRestClient(sonarApiUrl, "admin", password)
	if err != nil {
		return logErrorAndReturn(err)
	}

	plugins := []string{"authoidc", "checkstyle", "findbugs", "pmd"}
	sc.InstallPlugins(plugins)

	_, err = sc.UploadProfile()
	if err != nil {
		return err
	}

	_, err = sc.GenerateUserToken(JenkinsUsername)
	if err != nil {
		return err
	}

	err = sc.CreateGroup(GroupName)
	if err != nil {
		return err
	}

	err = sc.AddUserToGroup(GroupName, JenkinsUsername)
	if err != nil {
		return err
	}

	err = sc.AddPermissionsToUser(JenkinsUsername, "admin")
	if err != nil {
		return err
	}

	err = sc.AddPermissionsToGroup(GroupName, "scan")
	if err != nil {
		return err
	}

	err = sc.AddWebhook(JenkinsUsername, WebhookUrl)
	if err != nil {
		return err
	}

	log.Println("Sonar component configuration has been finished")
	return nil
}

// Invoking install method against SonarServiceImpl object should trigger list of methods, stored in client edp.PlatformService
func (s SonarServiceImpl) Install(instance v1alpha1.Sonar) error {

	if instance.Status.Status != StatusCreated {
		log.Printf("Installing Sonar component has been started")
		updateStatus(&instance, StatusInstall, time.Now())
	}

	dbSecret := map[string][]byte{
		"database-user":     []byte("admin"),
		"database-password": []byte(uniuri.New()),
	}

	err := s.platformService.CreateSecret(instance, instance.Name+"-db", dbSecret)
	if err != nil {
		return err
	}

	adminSecret := map[string][]byte{
		"user":     []byte("admin"),
		"password": []byte(uniuri.New()),
	}

	err = s.platformService.CreateSecret(instance, instance.Name+"-admin-password", adminSecret)
	if err != nil {
		return resourceActionFailed(&instance, err)
	}

	sa, err := s.platformService.CreateServiceAccount(instance)
	if err != nil {
		return resourceActionFailed(&instance, err)
	}

	err = s.platformService.CreateSecurityContext(instance, sa)
	if err != nil {
		return resourceActionFailed(&instance, err)
	}

	err = s.platformService.CreateDeployConf(instance)
	if err != nil {
		return resourceActionFailed(&instance, err)
	}

	err = s.platformService.CreateExternalEndpoint(instance)
	if err != nil {
		return resourceActionFailed(&instance, err)
	}

	err = s.platformService.CreateVolume(instance)
	if err != nil {
		return resourceActionFailed(&instance, err)
	}

	err = s.platformService.CreateService(instance)
	if err != nil {
		return resourceActionFailed(&instance, err)
	}

	err = s.platformService.CreateDbDeployConf(instance)
	if err != nil {
		return resourceActionFailed(&instance, err)
	}

	if instance.Status.Status != StatusCreated {
		log.Printf("Installing Sonar component has been finished")
		updateStatus(&instance, StatusCreated, time.Now())
	}

	err = s.updateAvailableStatus(instance, true)
	if err != nil {
		return resourceActionFailed(&instance, err)
	}

	_ = s.k8sClient.Update(context.TODO(), &instance)

	return nil
}

func (s SonarServiceImpl) updateAvailableStatus(instance v1alpha1.Sonar, value bool) error {
	if instance.Status.Available != value {
		instance.Status.Available = value
		instance.Status.LastTimeUpdated = time.Now()
		err := s.k8sClient.Update(context.TODO(), &instance)
		if err != nil {
			return err
		}
	}
	return nil
}

func updateStatus(s *v1alpha1.Sonar, status string, time time.Time) {
	s.Status.Status = status
	s.Status.LastTimeUpdated = time
	log.Printf("Status for Sonar %v has been updated to '%v' at %v.", s.Name, status, time)
}

func resourceActionFailed(instance *v1alpha1.Sonar, err error) error {
	updateStatus(instance, StatusFailed, time.Now())
	log.Printf("[ERROR] %v", err)
	return err
}
