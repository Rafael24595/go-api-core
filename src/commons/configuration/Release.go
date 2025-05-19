package configuration

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Rafael24595/go-api-core/src/commons/log"
)

type Release struct {
	TagName string `json:"tag_name"`
}

func OriginLastVersion(owner, repo string) *Release {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)

	resp, err := http.Get(url)
	if err != nil {
		log.Errorf("Error fetching release: %s", err.Error())
		return nil
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Errorf("GitHub API returned status: %s", resp.Status)
		return nil
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		log.Errorf("Error decoding JSON: %s", err.Error())
		return nil
	}

	return &release
}
