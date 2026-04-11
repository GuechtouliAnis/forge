package project

// import (
// 	"fmt"
// 	"os"

// 	"github.com/GuechtouliAnis/forge/internal/lang"
// 	"github.com/GuechtouliAnis/forge/internal/repo"
// )

// // CreateProject scaffolds a new project directory with the given name and language.
// // It creates the folder, generates a .gitignore, and sets up the language environment.
// // If withGit is true, it also runs git init and creates an initial commit.
// func CreateProject(name string, language string, withGit bool) error {

// 	// create the project directory with standard permissions
// 	if err := os.Mkdir(name, 0755); err != nil {
// 		return err
// 	}

// 	// move into the project directory for all subsequent setup steps
// 	if err := os.Chdir(name); err != nil {
// 		return err
// 	}

// 	// generate .gitignore before anything else so venv and build artifacts are excluded
// 	if err := repo.CreateGitignore(language); err != nil {
// 		return err
// 	}

// 	// route to the correct environment setup based on language
// 	switch language {
// 	case "py":
// 		if err := lang.SetupPython(); err != nil {
// 			return err
// 		}
// 	case "go":
// 		// empty string — username falls back to git config inside SetupGo
// 		if err := lang.SetupGo(""); err != nil {
// 			return err
// 		}
// 	}

// 	// initialize git and create the first commit if requested (forge init)
// 	if withGit {
// 		if err := run("git", "init"); err != nil {
// 			return err
// 		}
// 		if err := run("git", "add", "."); err != nil {
// 			return err
// 		}
// 		if err := run("git", "commit", "-m", "init: "+name); err != nil {
// 			return err
// 		}
// 	}

// 	fmt.Println("Project", name, "ready.")
// 	return nil
// }
