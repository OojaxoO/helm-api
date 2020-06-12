package charts

import (
	"os"
	"io"
	"strconv"
	"fmt"


	"github.com/golang/glog"
	"github.com/pkg/errors"
	"github.com/unknwon/com"
	"github.com/gin-gonic/gin"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/strvals"
	"helm.sh/helm/v3/pkg/storage/driver"
	"sigs.k8s.io/yaml"

	"helm-api/util"
	"helm-api/models"
)

type releaseElement struct {
	Name         string `json:"name"`
	Namespace    string `json:"namespace"`
	Revision     string `json:"revision"`
	Updated      string `json:"updated"`
	Status       string `json:"status"`
	Chart        string `json:"chart"`
	ChartVersion string `json:"chart_version"`
	AppVersion   string `json:"app_version"`

	Notes string `json:"notes,omitempty"`

	// TODO: Test Suite?
}

type releaseOptions struct {
	Values          string   `json:"values"`
	SetValues       []string `json:"set"`
	SetStringValues []string `json:"set_string"`
}


var settings = cli.New()

func actionConfigInit(cluster int, namespace string) (action.Configuration, error) {
	settings.Debug, _ = strconv.ParseBool(os.Getenv("HELM_DEBUG"))
	actionConfig := new(action.Configuration)
	kubeCluster, err := models.GetCluster(cluster)
	if err != nil {
		glog.Errorf("%+v", err)
		return *actionConfig, err
	}
	kubeConfig := "/tmp/config"
	kubeContext := kubeCluster.Config 
	util.WriteFile(kubeConfig, kubeContext)
	settings.KubeConfig = kubeConfig 
	err = actionConfig.Init(settings.RESTClientGetter(), namespace, os.Getenv("HELM_DRIVER"), glog.Infof)
	if err != nil {
		glog.Errorf("%+v", err)
		return *actionConfig, err
	}

	return *actionConfig, nil
}

func List(c *gin.Context) {
	namespace := c.Param("namespace")
	cluster := com.StrTo(c.Param("cluster")).MustInt()
	actionConfig, err := actionConfigInit(cluster, namespace)
	if err != nil {
		util.RespErr(c, err)
		return
	}
	client := action.NewList(&actionConfig)
	client.Deployed = true
	results, err := client.Run()
	if err != nil {
		util.RespErr(c, err)
		return
	}

	// Initialize the array so no results returns an empty array instead of null
	elements := make([]releaseElement, 0, len(results))
	for _, r := range results {
		elements = append(elements, constructReleaseElement(r, false))
	}
	util.RespOK(c, elements)
}

func constructReleaseElement(r *release.Release, showStatus bool) releaseElement {
	element := releaseElement{
		Name:         r.Name,
		Namespace:    r.Namespace,
		Revision:     strconv.Itoa(r.Version),
		Status:       r.Info.Status.String(),
		Chart:        r.Chart.Metadata.Name,
		ChartVersion: r.Chart.Metadata.Version,
		AppVersion:   r.Chart.Metadata.AppVersion,
	}
	if showStatus {
		element.Notes = r.Info.Notes
	}
	t := "-"
	if tspb := r.Info.LastDeployed; !tspb.IsZero() {
		t = tspb.String()
	}
	element.Updated = t

	return element
}

func Retrieve(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")
	cluster := com.StrTo(c.Param("cluster")).MustInt()
	actionConfig, err := actionConfigInit(cluster, namespace)
	if err != nil {
		util.RespErr(c, err)
		return
	}
	client := action.NewGet(&actionConfig)
	results, err := client.Run(name)
	if err != nil {
		util.RespErr(c, err)
		return
	}
	util.RespOK(c, results)
}

func mergeValues(options releaseOptions) (map[string]interface{}, error) {
	vals := map[string]interface{}{}
	err := yaml.Unmarshal([]byte(options.Values), &vals)
	if err != nil {
			return vals, fmt.Errorf("failed parsing values")
	}

	for _, value := range options.SetValues {
			if err := strvals.ParseInto(value, vals); err != nil {
					return vals, fmt.Errorf("failed parsing set data")
			}
	}

	for _, value := range options.SetStringValues {
			if err := strvals.ParseIntoString(value, vals); err != nil {
					return vals, fmt.Errorf("failed parsing set_string data")
			}
	}

	return vals, nil
}


func Update(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")
	chart := c.Query("chart")
	cluster := com.StrTo(c.Param("cluster")).MustInt()
	var options releaseOptions
    err := c.BindJSON(&options)
    if err != nil && err != io.EOF {
        util.RespErr(c, err)
        return
	}
	actionConfig, err := actionConfigInit(cluster, namespace)
	if err != nil {
		util.RespErr(c, err)
		return
	}
	vals, err := mergeValues(options)
	if err != nil {
		util.RespErr(c, err)
		return
	}
	client := action.NewUpgrade(&actionConfig)
	histClient := action.NewHistory(&actionConfig)
	histClient.Max = 1
	if _, err := histClient.Run(name); err == driver.ErrReleaseNotFound {
		instClient := action.NewInstall(&actionConfig)
		instClient.Namespace = namespace
		rel, err := runInstall(name, chart, instClient, vals)
		if err != nil {
			util.RespErr(c, err)
			return
		}
		util.RespOK(c, rel)
	} else if err != nil {
		util.RespErr(c, err)
		return
	}
	chartPath, err := client.ChartPathOptions.LocateChart(chart, settings)
	if err != nil {
		util.RespErr(c, err)
		return
	}

	// Check chart dependencies to make sure all are present in /charts
	ch, err := loader.Load(chartPath)
	if err != nil {
		util.RespErr(c, err)
		return
	}
	if req := ch.Metadata.Dependencies; req != nil {
		if err := action.CheckDependencies(ch, req); err != nil {
			util.RespErr(c, err)
			return
		}
	}

	client.Namespace = namespace
	rel, err := client.Run(name, ch, vals)
	if err != nil {
		util.RespErr(c, err)
		return
	}
	util.RespOK(c, rel)
} 

func isChartInstallable(ch *chart.Chart) (bool, error) {
	switch ch.Metadata.Type {
	case "", "application":
		return true, nil
	}
	return false, errors.Errorf("%s charts are not installable", ch.Metadata.Type)
}

func runInstall(name string, chart string, client *action.Install, vals map[string]interface{}) (*release.Release, error) {
	client.ReleaseName = name

	cp, err := client.ChartPathOptions.LocateChart(chart, settings)
	if err != nil {
		return nil, err
	}

	p := getter.All(settings)
	// Check chart dependencies to make sure all are present in /charts
	chartRequested, err := loader.Load(cp)
	if err != nil {
		return nil, err
	}

	validInstallableChart, err := isChartInstallable(chartRequested)
	if !validInstallableChart {
		return nil, err
	}

	if req := chartRequested.Metadata.Dependencies; req != nil {
		if err := action.CheckDependencies(chartRequested, req); err != nil {
			if client.DependencyUpdate {
				man := &downloader.Manager{
					ChartPath:        cp,
					Keyring:          client.ChartPathOptions.Keyring,
					SkipUpdate:       false,
					Getters:          p,
					RepositoryConfig: settings.RepositoryConfig,
					RepositoryCache:  settings.RepositoryCache,
					Debug:            settings.Debug,
				}
				if err := man.Update(); err != nil {
					return nil, err
				}
				// Reload the chart with the updated Chart.lock file.
				if chartRequested, err = loader.Load(cp); err != nil {
					return nil, errors.Wrap(err, "failed reloading chart after repo update")
				}
			} else {
				return nil, err
			}
		}
	}
	return client.Run(chartRequested, vals)
}


func Delete(c *gin.Context) {

}