package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"
)

type BuildPlatform struct {
	Id           int    `json:"id"`
	OSName       string `json:"os_name"`
	Platform     string `json:"platform"`
	Architecture string `json:"architecture"`
	ActiveStatus string `json:"active_status"`
}

type ResultSetList struct {
	Meta    interface{} `json:"meta"`
	Results []ResultSet `json:"results"`
}

type ResultSet struct {
	RepositoryId  int        `json:"repository_id"`
	RevisionHash  string     `json:"revision_hash"`
	RevisionCount int        `json:"revision_count"`
	Id            int        `json:"id"`
	Revisions     []Revision `json:"revisions"`
}

type JobsList struct {
	Meta    interface{} `json:"meta"`
	Results []Job       `json:"results"`
}

type JobData struct {
	BuildSystemType string     `json:"build_system_type"`
	ResourceURI     string     `json:"resource_uri"`
	Artifacts       []Artifact `json:"artifacts"`
}

type Job struct {
	PlatformOption  string `json:"platform_option"`
	BuildPlatformId int    `json:"build_platform_id"`
	Id              int    `json:"id"`
}

type Artifact struct {
	ResourceURI string `json:"resource_uri"`
	Name        string `json:"name"`
}

type Revision struct {
	Revision string `json:"revision"`
}

type BBInfo struct {
	Blob struct {
		LogURL string `json:"logurl"`
	} `json:"blob"`
}

type TCInfo struct {
	Blob struct {
		JobDetails []struct {
			URL string `json:"url"`
		} `json:"job_details"`
	} `json:"blob"`
}

func readInto(url string, obj interface{}) {
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	decoder := json.NewDecoder(resp.Body)
	defer resp.Body.Close()
	err = decoder.Decode(obj)
	if err != nil {
		panic(err)
	}
}

func curl(url string, file string) {
	res, err := http.Get(url)
	if err != nil {
		log.Println("Warning: could not get URL " + url)
		return
	}
	if res.StatusCode != 200 {
		// log.Println("Warning: non 200 response code from URL " + url)
		return
	}
	scanner := bufio.NewScanner(res.Body)
	started := false
	text := ""
	for scanner.Scan() {
		if !started && strings.Contains(scanner.Text(), "MultiFileLogger online at") {
			started = true
		}
		if started {
			text += tcBuildDir.ReplaceAllLiteralString(bbBuildDir.ReplaceAllLiteralString(time.ReplaceAllLiteralString(scanner.Text(), "XX:XX:XX"), "<BUILD-DIR>"), "<BUILD-DIR>") + "\n"
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatalf("Warning: problem reading from URL %s: %s", url, err)
	}
	if len(text) == 0 {
		return
	}
	err = os.MkdirAll(path.Dir(file), 0755)
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile(file, []byte(text), 0644)
	if err != nil {
		panic(err)
	}
	log.Println("Fetched " + url + " into " + file)
}

var (
	time       *regexp.Regexp
	bbBuildDir *regexp.Regexp
	tcBuildDir *regexp.Regexp
)

func init() {
	time = regexp.MustCompile(`^[0-2][0-9]:[0-5][0-9]:[0-5][0-9]`)
	bbBuildDir = regexp.MustCompile(`[\\/]?[cC]:?[\\/]+builds[\\/]+moz2_slave[\\/]+try-w[0-9a-z\-]*`)
	tcBuildDir = regexp.MustCompile(`[\\/]?[cC]:?[\\/]+Users[\\/]+Task_[0-9]*`)
}

func main() {
	args := os.Args
	if len(args) != 2 {
		log.Fatalf("Please specify an hg commit author (email address)")
	}
	buildPlatforms := []BuildPlatform{}
	readInto("https://treeherder.mozilla.org/api/buildplatform/", &buildPlatforms)
	bpmap := make(map[int]BuildPlatform)
	for _, bp := range buildPlatforms {
		bpmap[bp.Id] = bp
	}
	resultSetList := new(ResultSetList)
	readInto("https://treeherder.mozilla.org/api/project/try/resultset/?count=10&author="+args[1], resultSetList)
	for _, rs := range resultSetList.Results {
		log.Printf("Revision: %s", rs.Revisions[0].Revision)
		jobsURL := fmt.Sprintf("https://treeherder.mozilla.org/api/project/try/jobs/?count=2000&result_set_id=%v", rs.Id)
		jobsList := new(JobsList)
		readInto(jobsURL, jobsList)
		for _, job := range jobsList.Results {
			if (job.PlatformOption == "opt" || job.PlatformOption == "debug") && (bpmap[job.BuildPlatformId].Platform == "windowsxp" || bpmap[job.BuildPlatformId].Platform == "windows8-64") {
				// log.Printf("  Job: %s, %v", job.PlatformOption, bpmap[job.BuildPlatformId].Platform)
				jobData := new(JobData)
				jobDataURL := fmt.Sprintf("https://treeherder.mozilla.org/api/project/try/jobs/%v/", job.Id)
				readInto(jobDataURL, jobData)
				for _, jd := range jobData.Artifacts {
					if jd.Name == "Job Info" {
						// log.Printf("    %s", jd.ResourceURI)
						resURL := "https://treeherder.mozilla.org" + jd.ResourceURI
						dir := path.Join(args[1], rs.Revisions[0].Revision, job.PlatformOption, bpmap[job.BuildPlatformId].Platform)
						switch jobData.BuildSystemType {
						case "buildbot":
							bbInfo := new(BBInfo)
							readInto(resURL, bbInfo)
							// log.Println("      buildbot:" + bbInfo.Blob.LogURL)
							curl(bbInfo.Blob.LogURL, path.Join(dir, "bb"))
						case "taskcluster":
							tcInfo := new(TCInfo)
							readInto(resURL, tcInfo)
							// log.Println("      taskcluster:")
							for _, d := range tcInfo.Blob.JobDetails {
								// log.Println("        " + d.URL)
								logURL := strings.Replace(d.URL, "tools.taskcluster.net/task-inspector/#", "public-artifacts.taskcluster.net/", 1) + "/public/logs/all_commands.log"
								curl(logURL, path.Join(dir, "tc"))
							}
						}
					}
				}
			}
		}
	}
}
