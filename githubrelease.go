package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

var (
	// GitHub API URL
	apiURLFlag = flag.String("api-url", "https://api.github.com", "Base URL for the GitHub API")

	// Access token used for all interactions with the github api. The user will need to have access to the repo.
	patFlag = flag.String("pat", "", "Github Personal Access Token that should be used for the releases")

	// Repository user name name
	userFlag = flag.String("user", "imitablerabbit", "User namespace that the repository is located under")
	repoFlag = flag.String("repo", "", "Repository name exactly as it appears on GitHub")

	// Command line flags that can be used to create the release data
	tagFlag             = flag.String("release-tag", "", "The tag_name that should be used for the release. This does not have to be related to an actual git tag, although it probably should be.")
	targetCommitishFlag = flag.String("target", "master", "The commit/branch/tag that the release should be based on")
	nameFlag            = flag.String("name", "", "The name of the release")
	bodyFlag            = flag.String("body", "", "The body of the release")
	draftFlag           = flag.Bool("draft", false, "Is this release a draft? i.e. should it be shown publically")
	prereleaseFlag      = flag.Bool("prerelease", false, "Is this release a pre-release?")

	// The folder that contains all of the files that should be uploaded as part of the release.
	// If there are no files found in the folder, then no files will be uploaded as part of the release. The upload
	// URL can be retrieved later on for manual upload by using the github api to list details of the release.
	uploadsFlag = flag.String("uploads", "uploads/", "Directory that contains all of the tar.gx files that should be uploaded with the release")

	httpClient = http.Client{}
)

func init() {
	flag.Parse()
}

// CreateReleaseRequest represents the post data in the request to create a new GitHub release.
type CreateReleaseRequest struct {
	TagName         string `json:"tag_name"`
	TargetCommitish string `json:"target_commitish"`
	Name            string `json:"name"`
	Body            string `json:"body"`
	Draft           bool   `json:"draft"`
	PreRelease      bool   `json:"prerelease"`
}

// Send will send the http POST request that will create the GitHub release. A CreateReleaseResponse
// will be returned.
func (crr *CreateReleaseRequest) Send(apiURL, user, repo, pat string) (*Release, error) {
	releaseURL := fmt.Sprintf("%s/repos/%s/%s/releases", apiURL, user, repo)
	log.Printf("info: sending create request to %s", releaseURL)
	data, err := json.Marshal(crr)
	if err != nil {
		return nil, fmt.Errorf("json marshal CreateReleaseRequest: %v", err)
	}
	request, err := http.NewRequest(http.MethodPost, releaseURL, bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("creating release request: %v", err)
	}
	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("Authorization", "token "+pat)
	resp, err := httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("sending create release request: %v", err)
	}
	defer resp.Body.Close()
	respData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading create release response body: %v", err)
	}
	if resp.StatusCode != 201 {
		return nil, fmt.Errorf("non 201 response: %s: %s", resp.Status, respData)
	}
	log.Printf("info: received 201 response: %s", respData)
	crResponse := &Release{}
	if err := json.Unmarshal(respData, crResponse); err != nil {
		return nil, fmt.Errorf("unmarshaling response body: %v", err)
	}
	return crResponse, nil
}

// Release is the data that the GitHub api sends back from the
// create release endpoint.
type Release struct {
	URL        string `json:"url"`
	HTMLURL    string `json:"html_url"`
	AssetsURL  string `json:"assets_url"`
	UploadURL  string `json:"upload_url"`
	TarballURL string `json:"tarball_url"`
	ZipballURL string `json:"zipball_url"`

	ID     int    `json:"id"`
	NodeID string `json:"node_id"`

	TagName         string `json:"tag_name"`
	TargetCommitish string `json:"target_commitish"`
	Name            string `json:"name"`
	Body            string `json:"body"`
	Draft           bool   `json:"draft"`
	PreRelease      bool   `json:"prerelease"`

	CreatedAt   string `json:"created_at"`
	PublishedAt string `json:"published_at"`

	// Author information about who created the asset
	Author map[string]interface{} `json:"author"`

	// Assets contains all of the assets for that release
	Assets []map[string]interface{} `json:"assets"`
}

// UploadAsset will upload an asset to the newly created release.
func (crr *Release) UploadAsset(dir, filename, pat string) error {
	filepath := dir + "/" + filename
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("reading file for upload: %v", err)
	}
	uploadURL := strings.TrimSuffix(crr.UploadURL, "{?name,label}")
	url := fmt.Sprintf("%s?name=%s", uploadURL, filename)
	log.Printf("info: sending upload request to %s", url)
	request, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("creating upload request: %v", err)
	}
	request.Header.Add("Content-Type", "application/tar+gzip")
	request.Header.Add("Authorization", "token "+pat)
	resp, err := httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("sending upload request: %v", err)
	}
	defer resp.Body.Close()
	respData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading upload response body: %v", err)
	}
	if resp.StatusCode != 201 {
		return fmt.Errorf("non 201 response: %s: %s", resp.Status, respData)
	}
	return nil
}

func main() {
	req := &CreateReleaseRequest{
		TagName:         *tagFlag,
		TargetCommitish: *targetCommitishFlag,
		Name:            *nameFlag,
		Body:            *bodyFlag,
		Draft:           *draftFlag,
		PreRelease:      *prereleaseFlag,
	}
	release, err := req.Send(*apiURLFlag, *userFlag, *repoFlag, *patFlag)
	if err != nil {
		log.Fatalf("error: creating release: %v\n", err)
	}

	// Loop through all the files in the directory and upload them
	files, err := ioutil.ReadDir(*uploadsFlag)
	if err != nil {
		log.Fatalf("error: reading assets dir: %v\n", err)
	}
	for _, f := range files {
		// Just ignore sub directories, this should just be a directory full of .tar.gz files
		if f.IsDir() {
			continue
		}

		err := release.UploadAsset(*uploadsFlag, f.Name(), *patFlag)
		if err != nil {
			log.Printf("warn: uploading an asset: %v\n", err)
		}
	}
}
