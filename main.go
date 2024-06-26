package main

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/go-git/go-git/v5"
	gitHttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/google/go-github/v41/github"
	"github.com/rs/cors"
	"github.com/xanzy/go-gitlab"
	"golang.org/x/oauth2"
)

//go:embed static/*
var staticFiles embed.FS

type Repo struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type RepoListResponse struct {
	Repos []Repo `json:"repos"`
}

func openBrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}

	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	mux := http.NewServeMux()

	// Serve static files
	fileServer := http.FileServer(http.FS(staticFiles))
	mux.Handle("/", http.StripPrefix("/", fileServer))

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
	})

	handler := c.Handler(mux)

	mux.HandleFunc("/list", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		service := r.FormValue("service")
		groupPath := r.FormValue("group_path")
		token := r.FormValue("token")

		if service == "" || groupPath == "" || token == "" {
			http.Error(w, "Parâmetros inválidos", http.StatusBadRequest)
			return
		}

		var repos []Repo
		var err error

		if strings.ToLower(service) == "github" {
			repos, err = listGitHubRepos(groupPath, token)
		} else if strings.ToLower(service) == "gitlab" {
			repos, err = listGitLabRepos(groupPath, token)
		} else {
			http.Error(w, "Serviço não suportado", http.StatusBadRequest)
			return
		}

		if err != nil {
			http.Error(w, "Erro ao listar repositórios: "+err.Error(), http.StatusInternalServerError)
			return
		}

		response := RepoListResponse{Repos: repos}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	mux.HandleFunc("/clone", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		repoURL := r.FormValue("repo_url")
		username := r.FormValue("username")
		password := r.FormValue("password")

		if repoURL == "" || username == "" || password == "" {
			http.Error(w, "Parâmetros inválidos", http.StatusBadRequest)
			return
		}

		_, err := git.PlainClone("./repos", false, &git.CloneOptions{
			URL:      repoURL,
			Progress: os.Stdout,
			Auth: &gitHttp.BasicAuth{
				Username: username,
				Password: password,
			},
		})

		if err != nil {
			http.Error(w, "Erro ao clonar o repositório: "+err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Repositório clonado com sucesso!")
	})

	go func() {
		log.Println("Server starting on port 8080...")
		log.Fatal(http.ListenAndServe(":8080", handler))
	}()

	openBrowser("http://localhost:8080/static")

	// Prevent the main function from exiting
	select {}
}

func listGitHubRepos(username, token string) ([]Repo, error) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	repos, _, err := client.Repositories.List(ctx, username, nil)
	if err != nil {
		return nil, err
	}

	var result []Repo
	for _, repo := range repos {
		result = append(result, Repo{
			Name: repo.GetName(),
			URL:  repo.GetHTMLURL(),
		})
	}
	return result, nil
}

func listGitLabRepos(groupPath, token string) ([]Repo, error) {
	client, err := gitlab.NewClient(token, gitlab.WithBaseURL("https://gitlab.com/api/v4"))
	if err != nil {
		return nil, err
	}

	opts := &gitlab.ListGroupProjectsOptions{}

	projects, _, err := client.Groups.ListGroupProjects(groupPath, opts)
	if err != nil {
		return nil, err
	}

	var result []Repo
	for _, project := range projects {
		result = append(result, Repo{
			Name: project.Name,
			URL:  project.WebURL,
		})
	}
	return result, nil
}

// func cloneRepoHandler(w http.ResponseWriter, r *http.Request) {
// 	r.ParseForm()
// 	repoURL := r.FormValue("repo_url")
// 	username := r.FormValue("username")
// 	password := r.FormValue("password")

// 	if repoURL == "" || username == "" || password == "" {
// 		http.Error(w, "Parâmetros inválidos", http.StatusBadRequest)
// 		return
// 	}

// 	_, err := git.PlainClone("./repos", false, &git.CloneOptions{
// 		URL:      repoURL,
// 		Progress: os.Stdout,
// 		Auth: &gitHttp.BasicAuth{
// 			Username: username,
// 			Password: password,
// 		},
// 	})
// 	if err != nil {
// 		http.Error(w, "Erro ao clonar o repositório: "+err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	fmt.Fprintf(w, "Repositório clonado com sucesso!")
// }
