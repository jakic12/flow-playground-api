package playground_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/99designs/gqlgen/client"
	"github.com/99designs/gqlgen/handler"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	playground "github.com/dapperlabs/flow-playground-api"
	"github.com/dapperlabs/flow-playground-api/middleware"
	"github.com/dapperlabs/flow-playground-api/storage"
	"github.com/dapperlabs/flow-playground-api/storage/datastore"
	"github.com/dapperlabs/flow-playground-api/storage/memory"
	"github.com/dapperlabs/flow-playground-api/vm"
)

type Project struct {
	ID       string
	Title    string
	Seed     int
	Persist  bool
	Accounts []struct {
		ID        string
		Address   string
		DraftCode string
	}
	TransactionTemplates []TransactionTemplate
	Secret               string
}

const MutationCreateProject = `
mutation($title: String!, $seed: Int!, $accounts: [String!], $transactionTemplates: [NewProjectTransactionTemplate!]) {
  createProject(input: { title: $title, seed: $seed, accounts: $accounts, transactionTemplates: $transactionTemplates }) {
    id
    title
    seed
    persist
    accounts {
      id
      address
      draftCode
    }
    transactionTemplates {
      id
      title
      script
      index
    }
  }
}
`

type CreateProjectResponse struct {
	CreateProject Project
}

const QueryGetProject = `
query($projectId: UUID!) {
  project(id: $projectId) {
    id
    accounts {
      id
      address
    }
  }
}
`

type GetProjectResponse struct {
	Project Project
}

const MutationUpdateProjectPersist = `
mutation($projectId: UUID!, $persist: Boolean!) {
  updateProject(input: { id: $projectId, persist: $persist }) {
    id
    persist
  }
}
`

type UpdateProjectResponse struct {
	UpdateProject struct {
		ID      string
		Persist bool
	}
}

const QueryGetProjectTransactionTemplates = `
query($projectId: UUID!) {
  project(id: $projectId) {
    id
    transactionTemplates {
      id
      script 
      index
    }
  }
}
`

type GetProjectTransactionTemplatesResponse struct {
	Project struct {
		ID                   string
		TransactionTemplates []struct {
			ID     string
			Script string
			Index  int
		}
	}
}

const QueryGetProjectScriptTemplates = `
query($projectId: UUID!) {
  project(id: $projectId) {
    id
    scriptTemplates {
      id
      script 
      index
    }
  }
}
`

type GetProjectScriptTemplatesResponse struct {
	Project struct {
		ID              string
		ScriptTemplates []struct {
			ID     string
			Script string
			Index  int
		}
	}
}

const QueryGetAccount = `
query($accountId: UUID!, $projectId: UUID!) {
  account(id: $accountId, projectId: $projectId) {
    id
    address
    draftCode
    deployedCode
    state
  }
}
`

type GetAccountResponse struct {
	Account struct {
		ID           string
		Address      string
		DraftCode    string
		DeployedCode string
		State        string
	}
}

const MutationUpdateAccountDraftCode = `
mutation($accountId: UUID!, $projectId: UUID!, $code: String!) {
  updateAccount(input: { id: $accountId, projectId: $projectId, draftCode: $code }) {
	id
    address
    draftCode
    deployedCode
  }
}
`

const MutationUpdateAccountDeployedCode = `
mutation($accountId: UUID!, $projectId: UUID!, $code: String!) {
  updateAccount(input: { id: $accountId, projectId: $projectId, deployedCode: $code }) {
	id
    address
    draftCode
    deployedCode
  }
}
`

type UpdateAccountResponse struct {
	UpdateAccount struct {
		ID           string
		Address      string
		DraftCode    string
		DeployedCode string
	}
}

type TransactionTemplate struct {
	ID     string
	Title  string
	Script string
	Index  int
}

const MutationCreateTransactionTemplate = `
mutation($projectId: UUID!, $title: String!, $script: String!) {
  createTransactionTemplate(input: { projectId: $projectId, title: $title, script: $script }) {
    id
    title
    script
    index
  }
}
`

type CreateTransactionTemplateResponse struct {
	CreateTransactionTemplate TransactionTemplate
}

const QueryGetTransactionTemplate = `
query($templateId: UUID!, $projectId: UUID!) {
  transactionTemplate(id: $templateId, projectId: $projectId) {
    id
    script
    index
  }
}
`

type GetTransactionTemplateResponse struct {
	TransactionTemplate struct {
		ID     string
		Script string
		Index  int
	}
}

const MutationUpdateTransactionTemplateScript = `
mutation($templateId: UUID!, $projectId: UUID!, $script: String!) {
  updateTransactionTemplate(input: { id: $templateId, projectId: $projectId, script: $script }) {
    id
    script
    index
  }
}
`

const MutationUpdateTransactionTemplateIndex = `
mutation($templateId: UUID!, $projectId: UUID!, $index: Int!) {
  updateTransactionTemplate(input: { id: $templateId, projectId: $projectId, index: $index }) {
    id
    script
    index
  }
}
`

type UpdateTransactionTemplateResponse struct {
	UpdateTransactionTemplate struct {
		ID     string
		Script string
		Index  int
	}
}

const MutationDeleteTransactionTemplate = `
mutation($templateId: UUID!, $projectId: UUID!) {
  deleteTransactionTemplate(id: $templateId, projectId: $projectId)
}
`

type DeleteTransactionTemplateResponse struct {
	DeleteTransactionTemplate string
}

const MutationCreateTransactionExecution = `
mutation($projectId: UUID!, $script: String!, $signers: [Address!]) {
  createTransactionExecution(input: {
    projectId: $projectId,
    script: $script,
    signers: $signers,
  }) {
    id
    script
    error
    logs
    events {
      type
      values
    }
  }
}
`

type CreateTransactionExecutionResponse struct {
	CreateTransactionExecution struct {
		ID     string
		Script string
		Error  string
		Logs   []string
		Events []struct {
			Type   string
			Values []string
		}
	}
}

const MutationCreateScriptTemplate = `
mutation($projectId: UUID!, $title: String!, $script: String!) {
  createScriptTemplate(input: { projectId: $projectId, title: $title, script: $script }) {
    id
    title
    script
    index
  }
}
`

type ScriptTemplate struct {
	ID     string
	Title  string
	Script string
	Index  int
}

type CreateScriptTemplateResponse struct {
	CreateScriptTemplate ScriptTemplate
}

const QueryGetScriptTemplate = `
query($templateId: UUID!, $projectId: UUID!) {
  scriptTemplate(id: $templateId, projectId: $projectId) {
    id
    script
  }
}
`

type GetScriptTemplateResponse struct {
	ScriptTemplate ScriptTemplate
}

const MutationUpdateScriptTemplateScript = `
mutation($templateId: UUID!, $projectId: UUID!, $script: String!) {
  updateScriptTemplate(input: { id: $templateId, projectId: $projectId, script: $script }) {
    id
    script
    index
  }
}
`

const MutationUpdateScriptTemplateIndex = `
mutation($templateId: UUID!, $projectId: UUID!, $index: Int!) {
  updateScriptTemplate(input: { id: $templateId, projectId: $projectId, index: $index }) {
    id
    script
    index
  }
}
`

type UpdateScriptTemplateResponse struct {
	UpdateScriptTemplate struct {
		ID     string
		Script string
		Index  int
	}
}

const MutationDeleteScriptTemplate = `
mutation($templateId: UUID!, $projectId: UUID!) {
  deleteScriptTemplate(id: $templateId, projectId: $projectId)
}
`

type DeleteScriptTemplateResponse struct {
	DeleteScriptTemplate string
}

func TestProjects(t *testing.T) {
	t.Run("Create empty project", func(t *testing.T) {
		c := newClient()

		var resp CreateProjectResponse

		c.MustPost(
			MutationCreateProject,
			&resp,
			client.Var("title", "foo"),
			client.Var("seed", 42),
		)

		assert.NotEmpty(t, resp.CreateProject.ID)
		assert.Equal(t, 42, resp.CreateProject.Seed)

		// project should be created with 4 default accounts
		assert.Len(t, resp.CreateProject.Accounts, playground.MaxAccounts)

		// project should not be persisted
		assert.False(t, resp.CreateProject.Persist)
	})

	t.Run("Create project with 2 accounts", func(t *testing.T) {
		c := newClient()

		var resp CreateProjectResponse

		accounts := []string{
			"pub contract Foo {}",
			"pub contract Bar {}",
		}

		c.MustPost(
			MutationCreateProject,
			&resp,
			client.Var("title", "foo"),
			client.Var("seed", 42),
			client.Var("accounts", accounts),
		)

		// project should still be created with 4 default accounts
		assert.Len(t, resp.CreateProject.Accounts, playground.MaxAccounts)

		assert.Equal(t, accounts[0], resp.CreateProject.Accounts[0].DraftCode)
		assert.Equal(t, accounts[1], resp.CreateProject.Accounts[1].DraftCode)
		assert.Equal(t, "", resp.CreateProject.Accounts[2].DraftCode)
	})

	t.Run("Create project with 4 accounts", func(t *testing.T) {
		c := newClient()

		var resp CreateProjectResponse

		accounts := []string{
			"pub contract Foo {}",
			"pub contract Bar {}",
			"pub contract Dog {}",
			"pub contract Cat {}",
		}

		c.MustPost(
			MutationCreateProject,
			&resp,
			client.Var("title", "foo"),
			client.Var("seed", 42),
			client.Var("accounts", accounts),
		)

		// project should still be created with 4 default accounts
		assert.Len(t, resp.CreateProject.Accounts, playground.MaxAccounts)

		assert.Equal(t, accounts[0], resp.CreateProject.Accounts[0].DraftCode)
		assert.Equal(t, accounts[1], resp.CreateProject.Accounts[1].DraftCode)
		assert.Equal(t, accounts[2], resp.CreateProject.Accounts[2].DraftCode)
	})

	t.Run("Create project with transaction templates", func(t *testing.T) {
		c := newClient()

		var resp CreateProjectResponse

		templates := []struct {
			Title  string `json:"title"`
			Script string `json:"script"`
		}{
			{
				"foo", "transaction { execute { log(\"foo\") } }",
			},
			{
				"bar", "transaction { execute { log(\"bar\") } }",
			},
		}

		c.MustPost(
			MutationCreateProject,
			&resp,
			client.Var("title", "foo"),
			client.Var("seed", 42),
			client.Var("transactionTemplates", templates),
		)

		assert.Len(t, resp.CreateProject.TransactionTemplates, 2)
		assert.Equal(t, templates[0].Title, resp.CreateProject.TransactionTemplates[0].Title)
		assert.Equal(t, templates[0].Script, resp.CreateProject.TransactionTemplates[0].Script)
		assert.Equal(t, templates[1].Title, resp.CreateProject.TransactionTemplates[1].Title)
		assert.Equal(t, templates[1].Script, resp.CreateProject.TransactionTemplates[1].Script)
	})

	t.Run("Get project", func(t *testing.T) {
		c := newClient()

		project := createProject(c)

		var resp GetProjectResponse

		c.MustPost(
			QueryGetProject,
			&resp,
			client.Var("projectId", project.ID),
		)

		assert.Equal(t, project.ID, resp.Project.ID)
	})

	t.Run("Get non-existent project", func(t *testing.T) {
		c := newClient()

		var resp CreateProjectResponse

		badID := uuid.New().String()

		err := c.Post(
			QueryGetProject,
			&resp,
			client.Var("projectId", badID),
		)

		assert.Error(t, err)
	})

	t.Run("Persist project without permission", func(t *testing.T) {
		c := newClient()

		project := createProject(c)

		var resp UpdateProjectResponse

		err := c.Post(
			MutationUpdateProjectPersist,
			&resp,
			client.Var("projectId", project.ID),
			client.Var("persist", true),
		)

		assert.Error(t, err)
	})

	t.Run("Persist project", func(t *testing.T) {
		c := newClient()

		project := createProject(c)

		var resp UpdateProjectResponse

		c.MustPost(
			MutationUpdateProjectPersist,
			&resp,
			client.Var("projectId", project.ID),
			client.Var("persist", true),
			client.AddCookie(middleware.MockProjectSessionCookie(project.ID, project.Secret)),
		)

		assert.Equal(t, project.ID, resp.UpdateProject.ID)
		assert.True(t, resp.UpdateProject.Persist)
	})
}

func TestTransactionTemplates(t *testing.T) {
	t.Run("Create transaction template without permission", func(t *testing.T) {
		c := newClient()

		project := createProject(c)

		var resp CreateTransactionTemplateResponse

		err := c.Post(
			MutationCreateTransactionTemplate,
			&resp,
			client.Var("projectId", project.ID),
			client.Var("title", "foo"),
			client.Var("script", "bar"),
		)

		assert.Error(t, err)
		assert.Empty(t, resp.CreateTransactionTemplate.ID)
	})

	t.Run("Create transaction template", func(t *testing.T) {
		c := newClient()

		project := createProject(c)

		var resp CreateTransactionTemplateResponse

		c.MustPost(
			MutationCreateTransactionTemplate,
			&resp,
			client.Var("projectId", project.ID),
			client.Var("title", "foo"),
			client.Var("script", "bar"),
			client.AddCookie(middleware.MockProjectSessionCookie(project.ID, project.Secret)),
		)

		assert.NotEmpty(t, resp.CreateTransactionTemplate.ID)
		assert.Equal(t, "foo", resp.CreateTransactionTemplate.Title)
		assert.Equal(t, "bar", resp.CreateTransactionTemplate.Script)
	})

	t.Run("Get transaction template", func(t *testing.T) {
		c := newClient()

		project := createProject(c)

		var respA CreateTransactionTemplateResponse

		c.MustPost(
			MutationCreateTransactionTemplate,
			&respA,
			client.Var("projectId", project.ID),
			client.Var("title", "foo"),
			client.Var("script", "bar"),
			client.AddCookie(middleware.MockProjectSessionCookie(project.ID, project.Secret)),
		)

		var respB GetTransactionTemplateResponse

		c.MustPost(
			QueryGetTransactionTemplate,
			&respB,
			client.Var("projectId", project.ID),
			client.Var("templateId", respA.CreateTransactionTemplate.ID),
		)

		assert.Equal(t, respA.CreateTransactionTemplate.ID, respB.TransactionTemplate.ID)
		assert.Equal(t, respA.CreateTransactionTemplate.Script, respB.TransactionTemplate.Script)
	})

	t.Run("Get non-existent transaction template", func(t *testing.T) {
		c := newClient()

		project := createProject(c)

		var resp GetTransactionTemplateResponse

		badID := uuid.New().String()

		err := c.Post(
			QueryGetTransactionTemplate,
			&resp,
			client.Var("projectId", project.ID),
			client.Var("templateId", badID),
		)

		assert.Error(t, err)
	})

	t.Run("Update transaction template without permission", func(t *testing.T) {
		c := newClient()

		project := createProject(c)

		var respA CreateTransactionTemplateResponse

		c.MustPost(
			MutationCreateTransactionTemplate,
			&respA,
			client.Var("projectId", project.ID),
			client.Var("title", "foo"),
			client.Var("script", "apple"),
			client.AddCookie(middleware.MockProjectSessionCookie(project.ID, project.Secret)),
		)

		templateID := respA.CreateTransactionTemplate.ID

		var respB UpdateTransactionTemplateResponse

		err := c.Post(
			MutationUpdateTransactionTemplateScript,
			&respB,
			client.Var("projectId", project.ID),
			client.Var("templateId", templateID),
			client.Var("script", "orange"),
		)

		assert.Error(t, err)
	})

	t.Run("Update transaction template", func(t *testing.T) {
		c := newClient()

		project := createProject(c)

		var respA CreateTransactionTemplateResponse

		c.MustPost(
			MutationCreateTransactionTemplate,
			&respA,
			client.Var("projectId", project.ID),
			client.Var("title", "foo"),
			client.Var("script", "apple"),
			client.AddCookie(middleware.MockProjectSessionCookie(project.ID, project.Secret)),
		)

		templateID := respA.CreateTransactionTemplate.ID

		var respB UpdateTransactionTemplateResponse

		c.MustPost(
			MutationUpdateTransactionTemplateScript,
			&respB,
			client.Var("projectId", project.ID),
			client.Var("templateId", templateID),
			client.Var("script", "orange"),
			client.AddCookie(middleware.MockProjectSessionCookie(project.ID, project.Secret)),
		)

		assert.Equal(t, respA.CreateTransactionTemplate.ID, respB.UpdateTransactionTemplate.ID)
		assert.Equal(t, respA.CreateTransactionTemplate.Index, respB.UpdateTransactionTemplate.Index)
		assert.Equal(t, "orange", respB.UpdateTransactionTemplate.Script)

		var respC struct {
			UpdateTransactionTemplate struct {
				ID     string
				Script string
				Index  int
			}
		}

		c.MustPost(
			MutationUpdateTransactionTemplateIndex,
			&respC,
			client.Var("projectId", project.ID),
			client.Var("templateId", templateID),
			client.Var("index", 1),
			client.AddCookie(middleware.MockProjectSessionCookie(project.ID, project.Secret)),
		)

		assert.Equal(t, respA.CreateTransactionTemplate.ID, respC.UpdateTransactionTemplate.ID)
		assert.Equal(t, 1, respC.UpdateTransactionTemplate.Index)
		assert.Equal(t, respB.UpdateTransactionTemplate.Script, respC.UpdateTransactionTemplate.Script)
	})

	t.Run("Update non-existent transaction template", func(t *testing.T) {
		c := newClient()

		project := createProject(c)

		var resp UpdateTransactionTemplateResponse

		badID := uuid.New().String()

		err := c.Post(
			MutationUpdateTransactionTemplateScript,
			&resp,
			client.Var("projectId", project.ID),
			client.Var("templateId", badID),
			client.Var("script", "bar"),
		)

		assert.Error(t, err)
	})

	t.Run("Get transaction templates for project", func(t *testing.T) {
		c := newClient()

		project := createProject(c)

		templateA := createTransactionTemplate(c, project)
		templateB := createTransactionTemplate(c, project)
		templateC := createTransactionTemplate(c, project)

		var resp GetProjectTransactionTemplatesResponse

		c.MustPost(
			QueryGetProjectTransactionTemplates,
			&resp,
			client.Var("projectId", project.ID),
		)

		assert.Len(t, resp.Project.TransactionTemplates, 3)
		assert.Equal(t, templateA.ID, resp.Project.TransactionTemplates[0].ID)
		assert.Equal(t, templateB.ID, resp.Project.TransactionTemplates[1].ID)
		assert.Equal(t, templateC.ID, resp.Project.TransactionTemplates[2].ID)

		assert.Equal(t, 0, resp.Project.TransactionTemplates[0].Index)
		assert.Equal(t, 1, resp.Project.TransactionTemplates[1].Index)
		assert.Equal(t, 2, resp.Project.TransactionTemplates[2].Index)
	})

	t.Run("Get transaction templates for non-existent project", func(t *testing.T) {
		c := newClient()

		var resp GetProjectTransactionTemplatesResponse

		badID := uuid.New().String()

		err := c.Post(
			QueryGetProjectTransactionTemplates,
			&resp,
			client.Var("projectId", badID),
		)

		assert.Error(t, err)
	})

	t.Run("Delete transaction template without permission", func(t *testing.T) {
		c := newClient()

		project := createProject(c)

		template := createTransactionTemplate(c, project)

		var resp DeleteTransactionTemplateResponse

		err := c.Post(
			MutationDeleteTransactionTemplate,
			&resp,
			client.Var("projectId", project.ID),
			client.Var("templateId", template.ID),
		)

		assert.Error(t, err)
		assert.Empty(t, resp.DeleteTransactionTemplate)
	})

	t.Run("Delete transaction template", func(t *testing.T) {
		c := newClient()

		project := createProject(c)

		template := createTransactionTemplate(c, project)

		var resp DeleteTransactionTemplateResponse

		c.MustPost(
			MutationDeleteTransactionTemplate,
			&resp,
			client.Var("projectId", project.ID),
			client.Var("templateId", template.ID),
			client.AddCookie(middleware.MockProjectSessionCookie(project.ID, project.Secret)),
		)

		assert.Equal(t, template.ID, resp.DeleteTransactionTemplate)
	})
}

func TestTransactionExecutions(t *testing.T) {
	t.Run("Create execution for non-existent project", func(t *testing.T) {
		c := newClient()

		badID := uuid.New().String()

		var resp CreateTransactionExecutionResponse

		err := c.Post(
			MutationCreateTransactionExecution,
			&resp,
			client.Var("projectId", badID),
			client.Var("script", "transaction { execute { log(\"Hello, World!\") } }"),
		)

		assert.Error(t, err)
	})

	t.Run("Create execution without permission", func(t *testing.T) {
		c := newClient()

		project := createProject(c)

		var resp CreateTransactionExecutionResponse

		const script = "transaction { execute { log(\"Hello, World!\") } }"

		err := c.Post(
			MutationCreateTransactionExecution,
			&resp,
			client.Var("projectId", project.ID),
			client.Var("script", script),
		)

		assert.Error(t, err)
	})

	t.Run("Create execution", func(t *testing.T) {
		c := newClient()

		project := createProject(c)

		var resp CreateTransactionExecutionResponse

		const script = "transaction { execute { log(\"Hello, World!\") } }"

		c.MustPost(
			MutationCreateTransactionExecution,
			&resp,
			client.Var("projectId", project.ID),
			client.Var("script", script),
			client.AddCookie(middleware.MockProjectSessionCookie(project.ID, project.Secret)),
		)

		assert.Empty(t, resp.CreateTransactionExecution.Error)
		assert.Contains(t, resp.CreateTransactionExecution.Logs, "\"Hello, World!\"")
		assert.Equal(t, script, resp.CreateTransactionExecution.Script)
	})

	t.Run("Multiple executions", func(t *testing.T) {
		c := newClient()

		project := createProject(c)

		var respA CreateTransactionExecutionResponse

		const script = "transaction { prepare(signer: AuthAccount) { AuthAccount(payer: signer) } }"

		c.MustPost(
			MutationCreateTransactionExecution,
			&respA,
			client.Var("projectId", project.ID),
			client.Var("script", script),
			client.Var("signers", []string{project.Accounts[0].Address}),
			client.AddCookie(middleware.MockProjectSessionCookie(project.ID, project.Secret)),
		)

		assert.Empty(t, respA.CreateTransactionExecution.Error)
		require.Len(t, respA.CreateTransactionExecution.Events, 1)

		eventA := respA.CreateTransactionExecution.Events[0]

		// first account should have address 0x05
		assert.Equal(t, "flow.AccountCreated", eventA.Type)
		assert.JSONEq(t, `{"type":"Address","value":"0x0000000000000005"}`, eventA.Values[0])

		var respB CreateTransactionExecutionResponse

		c.MustPost(
			MutationCreateTransactionExecution,
			&respB,
			client.Var("projectId", project.ID),
			client.Var("script", script),
			client.Var("signers", []string{project.Accounts[0].Address}),
			client.AddCookie(middleware.MockProjectSessionCookie(project.ID, project.Secret)),
		)

		require.Len(t, respB.CreateTransactionExecution.Events, 1)

		eventB := respB.CreateTransactionExecution.Events[0]

		// second account should have address 0x06
		assert.Equal(t, "flow.AccountCreated", eventB.Type)
		assert.JSONEq(t, `{"type":"Address","value":"0x0000000000000006"}`, eventB.Values[0])
	})

	t.Run("Multiple executions with cache reset", func(t *testing.T) {
		// manually construct resolver
		store := memory.NewStore()
		computer, _ := vm.NewComputer(128)
		resolver := playground.NewResolver(store, computer)

		c := newClientWithResolver(resolver)

		project := createProject(c)

		var respA CreateTransactionExecutionResponse

		const script = "transaction { prepare(signer: AuthAccount) { AuthAccount(payer: signer) } }"

		c.MustPost(
			MutationCreateTransactionExecution,
			&respA,
			client.Var("projectId", project.ID),
			client.Var("script", script),
			client.Var("signers", []string{project.Accounts[0].Address}),
			client.AddCookie(middleware.MockProjectSessionCookie(project.ID, project.Secret)),
		)

		assert.Empty(t, respA.CreateTransactionExecution.Error)
		require.Len(t, respA.CreateTransactionExecution.Events, 1)

		eventA := respA.CreateTransactionExecution.Events[0]

		// first account should have address 0x05
		assert.Equal(t, "flow.AccountCreated", eventA.Type)
		assert.JSONEq(t, `{"type":"Address","value":"0x0000000000000005"}`, eventA.Values[0])

		// clear ledger cache
		computer.ClearCache()

		var respB CreateTransactionExecutionResponse

		c.MustPost(
			MutationCreateTransactionExecution,
			&respB,
			client.Var("projectId", project.ID),
			client.Var("script", script),
			client.Var("signers", []string{project.Accounts[0].Address}),
			client.AddCookie(middleware.MockProjectSessionCookie(project.ID, project.Secret)),
		)

		require.Len(t, respB.CreateTransactionExecution.Events, 1)

		eventB := respB.CreateTransactionExecution.Events[0]

		// second account should have address 0x06
		assert.Equal(t, "flow.AccountCreated", eventB.Type)
		assert.JSONEq(t, `{"type":"Address","value":"0x0000000000000006"}`, eventB.Values[0])
	})
}

func TestScriptTemplates(t *testing.T) {
	t.Run("Create script template without permission", func(t *testing.T) {
		c := newClient()

		project := createProject(c)

		var resp CreateScriptTemplateResponse

		err := c.Post(
			MutationCreateScriptTemplate,
			&resp,
			client.Var("projectId", project.ID),
			client.Var("title", "foo"),
			client.Var("script", "bar"),
		)

		assert.Error(t, err)
		assert.Empty(t, resp.CreateScriptTemplate.ID)
	})

	t.Run("Create script template", func(t *testing.T) {
		c := newClient()

		project := createProject(c)

		var resp CreateScriptTemplateResponse

		c.MustPost(
			MutationCreateScriptTemplate,
			&resp,
			client.Var("projectId", project.ID),
			client.Var("title", "foo"),
			client.Var("script", "bar"),
			client.AddCookie(middleware.MockProjectSessionCookie(project.ID, project.Secret)),
		)

		assert.NotEmpty(t, resp.CreateScriptTemplate.ID)
		assert.Equal(t, "foo", resp.CreateScriptTemplate.Title)
		assert.Equal(t, "bar", resp.CreateScriptTemplate.Script)
	})

	t.Run("Get script template", func(t *testing.T) {
		c := newClient()

		project := createProject(c)

		var respA CreateScriptTemplateResponse

		c.MustPost(
			MutationCreateScriptTemplate,
			&respA,
			client.Var("projectId", project.ID),
			client.Var("title", "foo"),
			client.Var("script", "bar"),
			client.AddCookie(middleware.MockProjectSessionCookie(project.ID, project.Secret)),
		)

		var respB GetScriptTemplateResponse

		c.MustPost(
			QueryGetScriptTemplate,
			&respB,
			client.Var("projectId", project.ID),
			client.Var("templateId", respA.CreateScriptTemplate.ID),
		)

		assert.Equal(t, respA.CreateScriptTemplate.ID, respB.ScriptTemplate.ID)
		assert.Equal(t, respA.CreateScriptTemplate.Script, respB.ScriptTemplate.Script)
	})

	t.Run("Get non-existent script template", func(t *testing.T) {
		c := newClient()

		project := createProject(c)

		var resp GetScriptTemplateResponse

		badID := uuid.New().String()

		err := c.Post(
			QueryGetScriptTemplate,
			&resp,
			client.Var("projectId", project.ID),
			client.Var("templateId", badID),
		)

		assert.Error(t, err)
	})

	t.Run("Update script template without permission", func(t *testing.T) {
		c := newClient()

		project := createProject(c)

		var respA CreateScriptTemplateResponse

		c.MustPost(
			MutationCreateScriptTemplate,
			&respA,
			client.Var("projectId", project.ID),
			client.Var("title", "foo"),
			client.Var("script", "apple"),
			client.AddCookie(middleware.MockProjectSessionCookie(project.ID, project.Secret)),
		)

		templateID := respA.CreateScriptTemplate.ID

		var respB UpdateScriptTemplateResponse

		err := c.Post(
			MutationUpdateScriptTemplateScript,
			&respB,
			client.Var("projectId", project.ID),
			client.Var("templateId", templateID),
			client.Var("script", "orange"),
		)

		assert.Error(t, err)
	})

	t.Run("Update script template", func(t *testing.T) {
		c := newClient()

		project := createProject(c)

		var respA CreateScriptTemplateResponse

		c.MustPost(
			MutationCreateScriptTemplate,
			&respA,
			client.Var("projectId", project.ID),
			client.Var("title", "foo"),
			client.Var("script", "apple"),
			client.AddCookie(middleware.MockProjectSessionCookie(project.ID, project.Secret)),
		)

		templateID := respA.CreateScriptTemplate.ID

		var respB UpdateScriptTemplateResponse

		c.MustPost(
			MutationUpdateScriptTemplateScript,
			&respB,
			client.Var("projectId", project.ID),
			client.Var("templateId", templateID),
			client.Var("script", "orange"),
			client.AddCookie(middleware.MockProjectSessionCookie(project.ID, project.Secret)),
		)

		assert.Equal(t, respA.CreateScriptTemplate.ID, respB.UpdateScriptTemplate.ID)
		assert.Equal(t, respA.CreateScriptTemplate.Index, respB.UpdateScriptTemplate.Index)
		assert.Equal(t, "orange", respB.UpdateScriptTemplate.Script)

		var respC UpdateScriptTemplateResponse

		c.MustPost(
			MutationUpdateScriptTemplateIndex,
			&respC,
			client.Var("projectId", project.ID),
			client.Var("templateId", templateID),
			client.Var("index", 1),
			client.AddCookie(middleware.MockProjectSessionCookie(project.ID, project.Secret)),
		)

		assert.Equal(t, respA.CreateScriptTemplate.ID, respC.UpdateScriptTemplate.ID)
		assert.Equal(t, 1, respC.UpdateScriptTemplate.Index)
		assert.Equal(t, respB.UpdateScriptTemplate.Script, respC.UpdateScriptTemplate.Script)
	})

	t.Run("Update non-existent script template", func(t *testing.T) {
		c := newClient()

		project := createProject(c)

		var resp UpdateScriptTemplateResponse

		badID := uuid.New().String()

		err := c.Post(
			MutationUpdateScriptTemplateScript,
			&resp,
			client.Var("projectId", project.ID),
			client.Var("templateId", badID),
			client.Var("script", "bar"),
		)

		assert.Error(t, err)
	})

	t.Run("Get script templates for project", func(t *testing.T) {
		c := newClient()

		project := createProject(c)

		templateIDA := createScriptTemplate(c, project)
		templateIDB := createScriptTemplate(c, project)
		templateIDC := createScriptTemplate(c, project)

		var resp GetProjectScriptTemplatesResponse

		c.MustPost(
			QueryGetProjectScriptTemplates,
			&resp,
			client.Var("projectId", project.ID),
		)

		assert.Len(t, resp.Project.ScriptTemplates, 3)
		assert.Equal(t, templateIDA, resp.Project.ScriptTemplates[0].ID)
		assert.Equal(t, templateIDB, resp.Project.ScriptTemplates[1].ID)
		assert.Equal(t, templateIDC, resp.Project.ScriptTemplates[2].ID)

		assert.Equal(t, 0, resp.Project.ScriptTemplates[0].Index)
		assert.Equal(t, 1, resp.Project.ScriptTemplates[1].Index)
		assert.Equal(t, 2, resp.Project.ScriptTemplates[2].Index)
	})

	t.Run("Get script templates for non-existent project", func(t *testing.T) {
		c := newClient()

		var resp GetProjectScriptTemplatesResponse

		badID := uuid.New().String()

		err := c.Post(

			QueryGetProjectScriptTemplates,
			&resp,
			client.Var("projectId", badID),
		)

		assert.Error(t, err)
	})

	t.Run("Delete script template without permission", func(t *testing.T) {
		c := newClient()

		project := createProject(c)

		templateID := createScriptTemplate(c, project)

		var resp DeleteScriptTemplateResponse

		err := c.Post(
			MutationDeleteScriptTemplate,
			&resp,
			client.Var("projectId", project.ID),
			client.Var("templateId", templateID),
		)

		assert.Error(t, err)
	})

	t.Run("Delete script template", func(t *testing.T) {
		c := newClient()

		project := createProject(c)

		templateID := createScriptTemplate(c, project)

		var resp DeleteScriptTemplateResponse

		c.MustPost(
			MutationDeleteScriptTemplate,
			&resp,
			client.Var("projectId", project.ID),
			client.Var("templateId", templateID),
			client.AddCookie(middleware.MockProjectSessionCookie(project.ID, project.Secret)),
		)

		assert.Equal(t, templateID, resp.DeleteScriptTemplate)
	})
}

func TestAccounts(t *testing.T) {
	t.Run("Get account", func(t *testing.T) {
		c := newClient()

		project := createProject(c)
		account := project.Accounts[0]

		var resp GetAccountResponse

		c.MustPost(
			QueryGetAccount,
			&resp,
			client.Var("projectId", project.ID),
			client.Var("accountId", account.ID),
		)

		assert.Equal(t, account.ID, resp.Account.ID)
	})

	t.Run("Get non-existent account", func(t *testing.T) {
		c := newClient()

		project := createProject(c)

		var resp GetAccountResponse

		badID := uuid.New().String()

		err := c.Post(
			QueryGetAccount,
			&resp,
			client.Var("projectId", project.ID),
			client.Var("accountId", badID),
		)

		assert.Error(t, err)
	})

	t.Run("Update account draft code without permission", func(t *testing.T) {
		c := newClient()

		project := createProject(c)
		account := project.Accounts[0]

		var respA GetAccountResponse

		c.MustPost(
			QueryGetAccount,
			&respA,
			client.Var("projectId", project.ID),
			client.Var("accountId", account.ID),
		)

		assert.Equal(t, "", respA.Account.DraftCode)

		var respB UpdateAccountResponse

		err := c.Post(
			MutationUpdateAccountDraftCode,
			&respB,
			client.Var("projectId", project.ID),
			client.Var("accountId", account.ID),
			client.Var("code", "bar"),
		)

		assert.Error(t, err)
	})

	t.Run("Update account draft code", func(t *testing.T) {
		c := newClient()

		project := createProject(c)
		account := project.Accounts[0]

		var respA GetAccountResponse

		c.MustPost(
			QueryGetAccount,
			&respA,
			client.Var("projectId", project.ID),
			client.Var("accountId", account.ID),
		)

		assert.Equal(t, "", respA.Account.DraftCode)

		var respB UpdateAccountResponse

		c.MustPost(
			MutationUpdateAccountDraftCode,
			&respB,
			client.Var("projectId", project.ID),
			client.Var("accountId", account.ID),
			client.Var("code", "bar"),
			client.AddCookie(middleware.MockProjectSessionCookie(project.ID, project.Secret)),
		)

		assert.Equal(t, "bar", respB.UpdateAccount.DraftCode)
	})

	t.Run("Update account invalid deployed code", func(t *testing.T) {
		c := newClient()

		project := createProject(c)
		account := project.Accounts[0]

		var respA GetAccountResponse

		c.MustPost(
			QueryGetAccount,
			&respA,
			client.Var("projectId", project.ID),
			client.Var("accountId", account.ID),
		)

		assert.Equal(t, "", respA.Account.DeployedCode)

		var respB UpdateAccountResponse

		err := c.Post(
			MutationUpdateAccountDeployedCode,
			&respB,
			client.Var("projectId", project.ID),
			client.Var("accountId", account.ID),
			client.Var("code", "INVALID CADENCE"),
		)

		assert.Error(t, err)
		assert.Equal(t, "", respB.UpdateAccount.DeployedCode)
	})

	t.Run("Update account deployed code without permission", func(t *testing.T) {
		c := newClient()

		project := createProject(c)

		account := project.Accounts[0]

		var resp UpdateAccountResponse

		const contract = "pub contract Foo {}"

		err := c.Post(
			MutationUpdateAccountDeployedCode,
			&resp,
			client.Var("projectId", project.ID),
			client.Var("accountId", account.ID),
			client.Var("code", contract),
		)

		assert.Error(t, err)
	})

	t.Run("Update account deployed code", func(t *testing.T) {
		c := newClient()

		project := createProject(c)

		account := project.Accounts[0]

		var respA GetAccountResponse

		c.MustPost(
			QueryGetAccount,
			&respA,
			client.Var("projectId", project.ID),
			client.Var("accountId", account.ID),
		)

		assert.Equal(t, "", respA.Account.DeployedCode)

		var respB UpdateAccountResponse

		const contract = "pub contract Foo {}"

		c.MustPost(
			MutationUpdateAccountDeployedCode,
			&respB,
			client.Var("projectId", project.ID),
			client.Var("accountId", account.ID),
			client.Var("code", contract),
			client.AddCookie(middleware.MockProjectSessionCookie(project.ID, project.Secret)),
		)

		assert.Equal(t, contract, respB.UpdateAccount.DeployedCode)
	})

	t.Run("Update non-existent account", func(t *testing.T) {
		c := newClient()

		project := createProject(c)

		var resp UpdateAccountResponse

		badID := uuid.New().String()

		err := c.Post(
			MutationUpdateAccountDraftCode,
			&resp,
			client.Var("projectId", project.ID),
			client.Var("accountId", badID),
			client.Var("script", "bar"),
		)

		assert.Error(t, err)
	})
}

const counterContract = `
  pub contract Counting {

      pub event CountIncremented(count: Int)

      pub resource Counter {
          pub var count: Int

          init() {
              self.count = 0
          }

          pub fun add(_ count: Int) {
              self.count = self.count + count
              emit CountIncremented(count: self.count)
          }
      }

      pub fun createCounter(): @Counter {
          return <-create Counter()
      }
  }
`

// generateAddTwoToCounterScript generates a script that increments a counter.
// If no counter exists, it is created.
func generateAddTwoToCounterScript(counterAddress string) string {
	return fmt.Sprintf(
		`
            import 0x%s

            transaction {

                prepare(signer: AuthAccount) {
                    if signer.borrow<&Counting.Counter>(from: /storage/counter) == nil {
                        signer.save(<-Counting.createCounter(), to: /storage/counter)
                    }

                    signer.borrow<&Counting.Counter>(from: /storage/counter)!.add(2)
                }
            }
        `,
		counterAddress,
	)
}

func TestContractInteraction(t *testing.T) {
	c := newClient()

	project := createProject(c)

	accountA := project.Accounts[0]
	accountB := project.Accounts[1]

	var respA UpdateAccountResponse

	c.MustPost(
		MutationUpdateAccountDeployedCode,
		&respA,
		client.Var("projectId", project.ID),
		client.Var("accountId", accountA.ID),
		client.Var("code", counterContract),
		client.AddCookie(middleware.MockProjectSessionCookie(project.ID, project.Secret)),
	)

	assert.Equal(t, counterContract, respA.UpdateAccount.DeployedCode)

	addScript := generateAddTwoToCounterScript(accountA.Address)

	var respB CreateTransactionExecutionResponse

	c.MustPost(
		MutationCreateTransactionExecution,
		&respB,
		client.Var("projectId", project.ID),
		client.Var("script", addScript),
		client.Var("signers", []string{accountB.Address}),
		client.AddCookie(middleware.MockProjectSessionCookie(project.ID, project.Secret)),
	)

	assert.Empty(t, respB.CreateTransactionExecution.Error)
}

type Client struct {
	client   *client.Client
	resolver *playground.Resolver
}

func (c *Client) Post(query string, response interface{}, options ...client.Option) error {
	return c.client.Post(query, response, options...)
}

func (c *Client) MustPost(query string, response interface{}, options ...client.Option) {
	c.client.MustPost(query, response, options...)
}

func newClient() *Client {
	var store storage.Store

	// TODO: Should eventually start up the emulator and run all tests with datastore backend
	if strings.EqualFold(os.Getenv("FLOW_STORAGEBACKEND"), "datastore") {
		var err error
		store, err = datastore.NewDatastore(context.Background(), &datastore.Config{
			DatastoreProjectID: "dl-flow",
			DatastoreTimeout:   time.Second * 5,
		})
		if err != nil {
			// If datastore is expected, panic when we can't init
			panic(err)
		}
	} else {
		store = memory.NewStore()
	}

	computer, _ := vm.NewComputer(128)

	resolver := playground.NewResolver(store, computer)

	return newClientWithResolver(resolver)
}

func newClientWithResolver(resolver *playground.Resolver) *Client {
	router := chi.NewRouter()
	router.Use(middleware.MockProjectSessions())

	router.Handle(
		"/",
		handler.GraphQL(
			playground.NewExecutableSchema(playground.Config{Resolvers: resolver}),
		),
	)

	return &Client{
		client:   client.New(router),
		resolver: resolver,
	}
}

func createProject(c *Client) Project {
	var resp CreateProjectResponse

	c.MustPost(
		MutationCreateProject,
		&resp,
		client.Var("title", "foo"),
		client.Var("seed", 42),
		client.Var("accounts", []string{}),
		client.Var("transactionTemplates", []string{}),
	)

	proj := resp.CreateProject
	internalProj := c.resolver.LastCreatedProject()

	proj.Secret = internalProj.Secret.String()

	return proj
}

func createTransactionTemplate(c *Client, project Project) TransactionTemplate {
	var resp CreateTransactionTemplateResponse

	c.MustPost(
		MutationCreateTransactionTemplate,
		&resp,
		client.Var("projectId", project.ID),
		client.Var("title", "foo"),
		client.Var("script", "bar"),
		client.AddCookie(middleware.MockProjectSessionCookie(project.ID, project.Secret)),
	)

	return resp.CreateTransactionTemplate
}

func createScriptTemplate(c *Client, project Project) string {
	var resp CreateScriptTemplateResponse

	c.MustPost(
		MutationCreateScriptTemplate,
		&resp,
		client.Var("projectId", project.ID),
		client.Var("title", "foo"),
		client.Var("script", "bar"),
		client.AddCookie(middleware.MockProjectSessionCookie(project.ID, project.Secret)),
	)

	return resp.CreateScriptTemplate.ID
}
