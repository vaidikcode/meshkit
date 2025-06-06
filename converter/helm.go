package converter

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/meshery/meshkit/models/patterns"
	"github.com/meshery/meshkit/utils/helm"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"sigs.k8s.io/yaml"
)

type HelmConverter struct{}

func (h *HelmConverter) Convert(patternFile string) (string, error) {
	if patternFile == "" {
		return "", ErrLoadPattern(fmt.Errorf("empty input"), "input")
	}

	pattern, err := patterns.GetPatternFormat(patternFile)
	if err != nil {
		return "", ErrLoadPattern(err, patternFile)
	}
	if pattern.Name == "" {
		return "", ErrCreateHelmChart(fmt.Errorf("missing name"), "chart metadata")
	}
	if pattern.Version == "" {
		return "", ErrCreateHelmChart(fmt.Errorf("missing version"), "chart metadata")
	}		
	//fmt.Println("Pattern loaded successfully" + pattern.Name + " " + pattern.Version)
	k8sConverter := K8sConverter{}
	k8sManifest, err := k8sConverter.Convert(patternFile)
	if err != nil {
		return "", ErrConvertK8s(err)
	}

	chartName := helm.SanitizeHelmName(pattern.Name)
	if chartName == "" {
		chartName = pattern.Name
	}

	chartVersion := pattern.Version

	chartContent, err := createHelmChartContent(k8sManifest, chartName, chartVersion)
	if err != nil {
		return "", err 
	}

	return chartContent, nil
}

func createHelmChartContent(manifestContent, chartName, chartVersion string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", ErrCreateHelmChart(err, "getting user home directory")
	}

	mesheryDir := filepath.Join(homeDir, ".meshery")
	packageDir := filepath.Join(mesheryDir, "helm-packages")
	tempDir := filepath.Join(mesheryDir, "tmp", "helm")

	if err := os.MkdirAll(packageDir, 0755); err != nil {
		return "", ErrCreateHelmChart(err, "creating package directory")
	}

	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", ErrCreateHelmChart(err, "creating temp directory")
	}

	buildID := uuid.New().String()
	buildDir := filepath.Join(tempDir, buildID)
	chartSourcePath := filepath.Join(buildDir, chartName)

	defer func() {
		err := os.RemoveAll(buildDir)
		if err != nil {
			fmt.Printf("Warning: Failed to clean up build directory: %v\n", err)
		}
	}()
	
	if err := os.MkdirAll(chartSourcePath, 0755); err != nil {
		return "", ErrCreateHelmChart(err, "creating chart source directory")
	}

	templatesDir := filepath.Join(chartSourcePath, "templates")
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		return "", ErrCreateHelmChart(err, "creating templates directory")
	}

	chartMeta := &chart.Metadata{
		APIVersion:  "v3",
		Name:        chartName,
		Version:     chartVersion,
		Description: fmt.Sprintf("Helm chart for '%s' generated by Meshery", chartName),
		Type:        "application",
	}
	//fmt.Println("ChartMeta version:", chartMeta.Version)

	chartYamlContent, err := yaml.Marshal(chartMeta)
	if err != nil {
		return "", ErrCreateHelmChart(err, "marshaling Chart.yaml metadata")
	}

	if err := os.WriteFile(filepath.Join(chartSourcePath, "Chart.yaml"), chartYamlContent, 0644); err != nil {
		return "", ErrCreateHelmChart(err, "writing Chart.yaml")
	}

	if err := os.WriteFile(filepath.Join(templatesDir, "manifest.yaml"), []byte(manifestContent), 0644); err != nil {
		return "", ErrCreateHelmChart(err, "writing manifest.yaml")
	}

	packager := action.NewPackage()
	packager.Destination = packageDir

	packagedChartPath, err := packager.Run(chartSourcePath, nil)
	if err != nil {
		return "", ErrHelmPackage(err)
	}

	chartData, err := os.ReadFile(packagedChartPath)
	if err != nil {
		return "", ErrCreateHelmChart(err, "reading packaged chart")
	}

	if err := os.Remove(packagedChartPath); err != nil {
		fmt.Printf("Warning: Failed to clean up packaged chart: %v\n", err)	
	}

	return string(chartData), nil
}

