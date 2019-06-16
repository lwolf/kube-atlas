package config

const (
	defaultConfig = `
defaults:
	sourcePath: apps
	releasePath: releases
	clusterName: kube1
repositories: {}
releases: {}
`
	customRepoConfig = `
defaults:
	sourcePath: apps
	releasePath: releases
	clusterName: kube1
repositories:
  - name: lwolf-charts
    url: http://charts.lwolf.org
releases:
  - name: plex
    chart: lwolf-charts/plex
    version: 0.1.2
`
	validConfig = `
defaults:
	sourcePath: ./apps
	releasePath: ./releases
	clusterName: amz1

repositories: {}

releases:
  - name: prometheus
    namespace: monitoring
    chart: stable/prometheus
    version: v8.11.4
`
)
