package coredns_omada

import "regexp"

func filterSites(pattern string, sites []string) (results []string) {

	for _, site := range sites {
		match, _ := regexp.MatchString(pattern, site)
		if match {
			results = append(results, site)
		}
	}
	return
}
