package registry

import (
	"net/url"

	"github.com/CenturyLinkLabs/docker-reg-client/registry"
)

type Registry struct {
	URL    string
        Client *registry.Client
}

// MirrorContainer 
func () {
    options := types.ContainerListOptions{
	Quiet   bool
	Size    bool
	All     bool
	Latest  bool
	Since   string
	Before  string
	Limit   int
	Filters filters.Args
}

func ListContainers() {
	c := registry.NewClient()
	c.BaseURL, _ = url.Parse(Registry.URL)

	results, err := c.Search.Query("mysql", 1, 25)
	if err != nil {
	panic(err)
	}

	for _, result := range results.Results {
		fmt.Println(result.Name)
	}
}
