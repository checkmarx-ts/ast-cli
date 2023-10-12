package wrappers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/checkmarx/ast-cli/internal/logger"
	"github.com/checkmarx/ast-cli/internal/params"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/tomnomnom/linkheader"
)

type GitHubHTTPWrapper struct {
	client               *http.Client
	repositoryTemplate   string
	organizationTemplate string
}

const (
	acceptHeader        = "Accept"
	AuthorizationHeader = "Authorization"
	apiVersion          = "application/vnd.github.v3+json"
	tokenFormat         = "token %s"
	ownerPlaceholder    = "{owner}"
	repoPlaceholder     = "{repo}"
	orgPlaceholder      = "{org}"
	linkHeaderName      = "Link"
	nextRel             = "next"
	perPageParam        = "per_page"
	perPageValue        = "100"
	retryLimit          = 3
)

func NewGitHubWrapper() GitHubWrapper {
	return &GitHubHTTPWrapper{
		client: GetClient(viper.GetUint(params.ClientTimeoutKey)),
	}
}

func (g *GitHubHTTPWrapper) GetOrganization(organizationName string) (Organization, error) {
	var err error
	var organization Organization

	organizationTemplate, err := g.getOrganizationTemplate()
	if err != nil {
		return organization, err
	}
	organizationURL := strings.ReplaceAll(organizationTemplate, orgPlaceholder, organizationName)

	_, err = g.get(organizationURL, &organization, map[string]string{})

	return organization, err
}

func (g *GitHubHTTPWrapper) GetRepository(organizationName, repositoryName string) (Repository, error) {
	var err error
	var repository Repository

	repositoryURL, err := g.getRepositoryTemplate()
	if err != nil {
		return repository, err
	}
	repositoryURL = strings.ReplaceAll(repositoryURL, ownerPlaceholder, organizationName)
	repositoryURL = strings.ReplaceAll(repositoryURL, repoPlaceholder, repositoryName)

  _, err = g.get(repositoryURL, &repository, map[string]string{})

	return repository, err
}

func (g *GitHubHTTPWrapper) GetRepositories(organization Organization) ([]Repository, error) {
	repositoriesURL := organization.RepositoriesURL

	pages, err := g.getWithPagination(repositoriesURL, map[string]string{})
	if err != nil {
		return nil, err
	}

	castedPages := make([]Repository, 0)
	for _, e := range pages {
		marshal, err := json.Marshal(e)
		if err != nil {
			return nil, err
		}
		holder := Repository{}
		err = json.Unmarshal(marshal, &holder)
		if err != nil {
			return nil, err
		}
		castedPages = append(castedPages, holder)
	}

	return castedPages, nil
}

func (g *GitHubHTTPWrapper) GetCommits(repository Repository, queryParams map[string]string) ([]CommitRoot, error) {
  commitsURL := repository.CommitsURL

  fmt.Println("Calling GetCommits(): " + commitsURL)
  index := strings.Index(commitsURL, "{")
	if index < 0 {
		return nil, errors.Errorf("Unable to collect commits URL for repository %s", repository.FullName)
	}
	commitsURL = commitsURL[:index]

	pages, err := g.getWithPagination(commitsURL, queryParams)
	if err != nil {
		return nil, err
	}

	castedPages := make([]CommitRoot, 0)
	for _, e := range pages {
		marshal, err := json.Marshal(e)
		if err != nil {
			return nil, err
		}
		holder := CommitRoot{}
		err = json.Unmarshal(marshal, &holder)
		if err != nil {
			return nil, err
		}
		castedPages = append(castedPages, holder)
	}

	return castedPages, nil
}

func (g *GitHubHTTPWrapper) getOrganizationTemplate() (string, error) {
	var err error

	if g.organizationTemplate == "" {
		err = g.getTemplates()
	}

	return g.organizationTemplate, err
}

func (g *GitHubHTTPWrapper) getRepositoryTemplate() (string, error) {
	var err error

	if g.repositoryTemplate == "" {
		err = g.getTemplates()
	}

	return g.repositoryTemplate, err
}

func (g *GitHubHTTPWrapper) getTemplates() error {
	var err error
	var rootAPIResponse rootAPI

	baseURL := viper.GetString(params.URLFlag)
	_, err = g.get(baseURL, &rootAPIResponse, map[string]string{})

	g.organizationTemplate = rootAPIResponse.OrganizationURL
	g.repositoryTemplate = rootAPIResponse.RepositoryURL

	return err
}

func (g *GitHubHTTPWrapper) getWithPagination(
	url string,
	queryParams map[string]string,
) ([]interface{}, error) {
	queryParams[perPageParam] = perPageValue

	var pageCollection = make([]interface{}, 0)

	next, err := g.collectPage(g.client, url, queryParams, &pageCollection)
	if err != nil {
		return nil, err
	}

	for next != "" {
		next, err = g.collectPage(g.client, next, map[string]string{}, &pageCollection)
		if err != nil {
			return nil, err
		}
	}

	return pageCollection, nil
}

func (g *GitHubHTTPWrapper) collectPage(
	client *http.Client,
	url string,
	queryParams map[string]string,
	pageCollection *[]interface{},
) (string, error) {
	var holder = make([]interface{}, 0)

	resp, err := g.get(url, &holder, queryParams)
	if err != nil {
    return "", err
	}

	*pageCollection = append(*pageCollection, holder...)
	next := g.getNextPageLink(resp)

	return next, nil
}

func (g *GitHubHTTPWrapper) cleanUpResponse(resp *http.Response) {
  if resp != nil {
    resp.Body.Close()
  }
}

func (g *GitHubHTTPWrapper) getNextPageLink(resp *http.Response) string {
	if resp != nil {
		linkHeader := resp.Header[linkHeaderName]
		if len(linkHeader) > 0 {
			links := linkheader.Parse(linkHeader[0])
			for _, link := range links {
				if link.Rel == nextRel {
					return link.URL
				}
			}
		}
	}
	return ""
}

//
/// NOTE: (Jeff Armstrong) There was two different get() method. One was a member of the class the other 
/// the member was a simple wrapper for this call. The Body.Close() would be called too many
/// times with the original implementation, or called when there was no body. This was triggering
/// the crash the client was seeing. Now the logic is the same, just with fewer calls. ALL cleanUp 
/// CODE is now housed within the "defer cleanUpReponse()" in here. This may not be pefect but it 
/// works better
//
/// Finally, I think this method could also be optimized a little bit. Not todays subject 
//
func (g *GitHubHTTPWrapper) get(url string, target interface{}, queryParams map[string]string) (*http.Response, error) {
  req, err := http.NewRequest(http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, err
	}
	req.Header.Add(acceptHeader, apiVersion)
	token := viper.GetString(params.SCMTokenFlag)
	logger.PrintRequest(req)
	resp, err := GetWithQueryParamsAndCustomRequest(g.client, req, url, token, tokenFormat, queryParams)
	if err != nil {
		return nil, err
	} else if resp == nil {
    return nil, nil
  }
	defer g.cleanUpResponse(resp)
  logger.PrintResponse(resp, true)

	switch resp.StatusCode {
	case http.StatusOK:
		logger.PrintIfVerbose(fmt.Sprintf("Request to URL %s OK", req.URL))
		err = json.NewDecoder(resp.Body).Decode(target)
		if err != nil {
			return nil, err
		}
	case http.StatusConflict:
		logger.PrintIfVerbose(fmt.Sprintf("Found empty repository in %s", req.URL))
		return nil, nil
	default:
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			logger.PrintIfVerbose(err.Error())
			return nil, err
		}
		message := fmt.Sprintf("Code %d %s", resp.StatusCode, string(body))
		return nil, errors.New(message)
	}
	return resp, nil
}

