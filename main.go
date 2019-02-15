/*
	Copyright (C) 2019 Magnus Strömbrink

	This program is free software: you can redistribute it and/or modify
    it under the terms of the GNU General Public License as published by
    the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU General Public License for more details.

    You should have received a copy of the GNU General Public License
    along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"strings"
)

var debug bool
var giteaURL string
var giteaUsername string
var giteaPassword string
var sourcePath string

type Migrate struct {
	CloneAddr string `json:"clone_addr"`
	RepoName  string `json:"repo_name"`
	UID       int    `json:"UID"`
}

type RepoResponse struct {
	Archived        bool   `json:"archived"`
	CloneURL        string `json:"clone_url"`
	CreatedAt       string `json:"created_at"`
	DefaultBranch   string `json:"default_branch"`
	Description     string `json:"description"`
	Empty           bool   `json:"empty"`
	Fork            bool   `json:"fork"`
	ForksCount      int    `json:"forks_count"`
	FullName        string `json:"full_name"`
	HTMLURL         string `json:"html_url"`
	ID              int    `json:"id"`
	Mirror          bool   `json:"mirror"`
	Name            string `json:"name"`
	OpenIssuesCount int    `json:"open_issues_count"`
	Owner           struct {
		AvatarURL string `json:"avatar_url"`
		Email     string `json:"email"`
		FullName  string `json:"full_name"`
		ID        int    `json:"id"`
		Language  string `json:"language"`
		Login     string `json:"login"`
	} `json:"owner"`
	Permissions struct {
		Admin bool `json:"admin"`
		Pull  bool `json:"pull"`
		Push  bool `json:"push"`
	} `json:"permissions"`
	Private       bool   `json:"private"`
	Size          int    `json:"size"`
	SSHURL        string `json:"ssh_url"`
	StarsCount    int    `json:"stars_count"`
	UpdatedAt     string `json:"updated_at"`
	WatchersCount int    `json:"watchers_count"`
	Website       string `json:"website"`
}

type OrganizationResponse struct {
	AvatarURL   string `json:"avatar_url"`
	Description string `json:"description"`
	FullName    string `json:"full_name"`
	ID          int    `json:"id"`
	Location    string `json:"location"`
	Username    string `json:"username"`
	Website     string `json:"website"`
}

type OrganizationRequest struct {
	Description string `json:"description"`
	FullName    string `json:"full_name"`
	Location    string `json:"location"`
	Username    string `json:"username"`
	Website     string `json:"website"`
}

func httpPost(uri string, input interface{}, output interface{}) error {
	b, err := json.Marshal(input)
	if err != nil {
		log.Fatal(err)
	}
	buffer := bytes.NewBuffer(b)
	req, err := http.NewRequest("POST", giteaURL+uri, buffer)
	if err != nil {
		log.Fatal(err)
	}
	req.SetBasicAuth(giteaUsername, giteaPassword)
	req.Header.Set("Content-Type", "application/json")
	if debug {
		dump, err := httputil.DumpRequestOut(req, true)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("*** HTTP POST ***")
		log.Printf("Request:\n%s", string(dump))
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	if debug {
		dump, err := httputil.DumpResponse(resp, true)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Response:\n%s", string(dump))
	}
	if resp.StatusCode != 201 {
		return fmt.Errorf("%s", resp.Status)
	}
	b, err = ioutil.ReadAll(resp.Body)
	return json.Unmarshal(b, output)
}
func httpGet(uri string, output interface{}) error {
	req, err := http.NewRequest("GET", giteaURL+uri, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.SetBasicAuth(giteaUsername, giteaPassword)
	if debug {
		dump, err := httputil.DumpRequestOut(req, true)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("*** HTTP GET ***")
		log.Printf("Request:\n%s", string(dump))
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	if debug {
		dump, err := httputil.DumpResponse(resp, true)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Response:\n%s", string(dump))
	}
	b, err := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return fmt.Errorf("%s", resp.Status)
	}
	return json.Unmarshal(b, output)
}

func processGitRepo(path string, org string, name string) {
	log.Printf("GIT repo: %v", path)
	//
	// Check if repo exists
	if repo(org, name).ID != 0 {
		log.Printf("Repo already exists. Owner=%s Repo=%s", org, name)
		return
	}
	var m Migrate
	var repoResp RepoResponse
	m.CloneAddr = path
	m.RepoName = name
	m.UID = organization(org).ID
	//
	//
	err := httpPost("/api/v1/repos/migrate", m, &repoResp)
	if err != nil {
		log.Fatal(err)
	}
}

func processOrg(path string, org string) {
	log.Printf("ORG: %s", path)
	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		if file.IsDir() {
			if strings.HasSuffix(file.Name(), ".git") {
				processGitRepo(path+"/"+file.Name(), org, file.Name()[:len(file.Name())-4])
			} else {
				log.Fatal(fmt.Printf("Not a git repo: %s", path+"/"+file.Name()))
			}
		}
	}

}

func process(path string) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		if file.IsDir() {
			processOrg(path+"/"+file.Name(), file.Name())
		}
	}

}

func repo(owner string, name string) RepoResponse {
	var r RepoResponse
	err := httpGet(fmt.Sprintf("/api/v1/repos/%s/%s", owner, name), &r)
	if err != nil && err.Error() == "404 Not Found" {
		return RepoResponse{}
	} else if err != nil {
		log.Fatal(err)
	}
	return r
}

func organization(name string) OrganizationResponse {
	var orgResp OrganizationResponse
	err := httpGet("/api/v1/orgs/"+name, &orgResp)
	if err != nil && err.Error() == "404 Not Found" {
		var orgReq OrganizationRequest
		orgReq.Username = name
		err = httpPost("/api/v1/orgs", orgReq, &orgResp)
		if err != nil {
			log.Fatal(err)
		}
	} else if err != nil {
		log.Fatal(err)
	}
	return orgResp
}

func main() {
	flag.BoolVar(&debug, "debug", false, "Show debug output")
	flag.StringVar(&sourcePath, "path", "", "path to local repos")
	flag.StringVar(&giteaURL, "url", "", "URL to GITEA. Ex: https://git.server.com")
	flag.StringVar(&giteaUsername, "username", "", "GITEA username")
	flag.StringVar(&giteaPassword, "password", "", "GITEA password")
	flag.Parse()
	if sourcePath == "" {
		log.Fatal("Missing path")
	}
	if giteaURL == "" {
		log.Fatal("Missing url")
	}
	if giteaUsername == "" {
		log.Fatal("Missing username")
	}
	if giteaPassword == "" {
		log.Fatal("Missing password")
	}
	process(sourcePath)
}
