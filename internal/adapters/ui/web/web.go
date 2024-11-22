package web

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/mod-develop/backend/internal/adapters/types"
)

type Web struct {
}

func New() (*Web, error) {
	w := &Web{}

	return w, nil
}

func baseLayout(wr http.ResponseWriter, temp string) error {
	tmpl, err := template.ParseFiles("templates/base.html", temp)
	if err != nil {
		return fmt.Errorf("failed parse page: %w", err)
	}

	err = tmpl.ExecuteTemplate(wr, "base", nil)
	if err != nil {
		return fmt.Errorf("failed execute template: %w", err)
	}
	return nil
}

func baseUserLayout(wr http.ResponseWriter, temp string, data any) error {
	tmpl, err := template.ParseFiles("templates/user.html", temp)
	if err != nil {
		return fmt.Errorf("failed parse page: %w", err)
	}

	err = tmpl.ExecuteTemplate(wr, "base", data)
	if err != nil {
		return fmt.Errorf("failed execute template: %w", err)
	}
	return nil
}

func baseManagerLayout(wr http.ResponseWriter, temp string, data any) error {
	tmpl, err := template.ParseFiles("templates/user.html", "templates/manager/tabs.html", temp)
	if err != nil {
		return fmt.Errorf("failed parse page: %w", err)
	}

	err = tmpl.ExecuteTemplate(wr, "base", data)
	if err != nil {
		return fmt.Errorf("failed execute template: %w", err)
	}
	return nil
}

func basePlayerLayout(wr http.ResponseWriter, temp string, data any) error {
	tmpl, err := template.ParseFiles("templates/player/base.html", "templates/player/navigate.html", temp)
	if err != nil {
		return fmt.Errorf("failed parse page: %w", err)
	}

	err = tmpl.ExecuteTemplate(wr, "base", data)
	if err != nil {
		return fmt.Errorf("failed execute template: %w", err)
	}
	return nil
}

func (w *Web) RegistrationPage(wr http.ResponseWriter) error {
	err := baseLayout(wr, "templates/registration.html")
	if err != nil {
		return fmt.Errorf("failed execute template: %w", err)
	}
	return nil
}

func (w *Web) Error500Page(wr http.ResponseWriter) error {
	tmpl, err := template.ParseFiles("templates/500.html")
	if err != nil {
		return fmt.Errorf("failed parse page: %w", err)
	}

	wr.WriteHeader(http.StatusInternalServerError)
	err = tmpl.Execute(wr, nil)
	if err != nil {
		return fmt.Errorf("failed execute template: %w", err)
	}
	return nil
}

func (w *Web) Error403Page(wr http.ResponseWriter) error {
	err := baseLayout(wr, "templates/403.html")
	if err != nil {
		return fmt.Errorf("failed parse page: %w", err)
	}
	return nil
}

func (w *Web) Error404Page(wr http.ResponseWriter) error {
	err := baseLayout(wr, "templates/404.html")
	if err != nil {
		return fmt.Errorf("failed parse page: %w", err)
	}
	return nil
}

func (w *Web) AuthorizationPage(wr http.ResponseWriter) error {
	err := baseLayout(wr, "templates/authorization.html")
	if err != nil {
		return fmt.Errorf("failed execute template: %w", err)
	}
	return nil
}

func (w *Web) MainPage(wr http.ResponseWriter, page *types.TMainPage) error {
	err := baseUserLayout(wr, "templates/user/main.html", page)
	if err != nil {
		return fmt.Errorf("failed execute template: %w", err)
	}
	return nil
}

func (w *Web) ProfilePage(wr http.ResponseWriter, page *types.TProfilePage) error {
	err := baseUserLayout(wr, "templates/user/profile.html", page)
	if err != nil {
		return fmt.Errorf("failed execute template: %w", err)
	}
	return nil
}

func (w *Web) QuestGiverPage(wr http.ResponseWriter, page *types.TQuestGiverPage) error {
	err := baseManagerLayout(wr, "templates/manager/quests/index.html", page)
	if err != nil {
		return fmt.Errorf("failed execute template: %w", err)
	}
	return nil
}

func (w *Web) QuestEdit(wr http.ResponseWriter, page *types.TQuestEdit) error {
	err := baseManagerLayout(wr, "templates/manager/quests/edit.html", page)
	if err != nil {
		return fmt.Errorf("failed execute template: %w", err)
	}
	return nil
}

func (w *Web) QuestNew(wr http.ResponseWriter, page *types.TQuestNew) error {
	err := baseManagerLayout(wr, "templates/manager/quests/new.html", page)
	if err != nil {
		return fmt.Errorf("failed execute template: %w", err)
	}
	return nil
}

func (w *Web) QuestAwait(wr http.ResponseWriter, page *types.TQuestAwaitPage) error {
	err := baseManagerLayout(wr, "templates/manager/await/index.html", page)
	if err != nil {
		return fmt.Errorf("failed execute template: %w", err)
	}
	return nil
}

func (w *Web) PlayerQuests(wr http.ResponseWriter, page *types.TPlayerQuestsPage) error {
	err := basePlayerLayout(wr, "templates/player/quests/index.html", page)
	if err != nil {
		return fmt.Errorf("failed execute template: %w", err)
	}
	return nil
}

func (w *Web) PlayerProfile(wr http.ResponseWriter, page *types.TPlayerProfilePage) error {
	err := basePlayerLayout(wr, "templates/player/index.html", page)
	if err != nil {
		return fmt.Errorf("failed execute template: %w", err)
	}
	return nil
}

func (w *Web) PlayerQuest(wr http.ResponseWriter, page *types.TPlayerQuestPage) error {
	err := basePlayerLayout(wr, "templates/player/quests/get.html", page)
	if err != nil {
		return fmt.Errorf("failed execute template: %w", err)
	}
	return nil
}

func (w *Web) PlayerSettings(wr http.ResponseWriter, page *types.TPlayerSettingsPage) error {
	err := basePlayerLayout(wr, "templates/player/settings.html", page)
	if err != nil {
		return fmt.Errorf("failed execute template: %w", err)
	}
	return nil
}
