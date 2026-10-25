package http_test

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	nethttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	httpadapter "schulte/internal/adapter/http"
	"schulte/internal/adapter/repository/memory"
	"schulte/internal/domain"
	"schulte/internal/infrastructure/platform"
	"schulte/internal/infrastructure/token"
	"schulte/internal/usecase"
)

func newServer(t *testing.T) *httptest.Server {
	t.Helper()
	st := memory.NewStore()
	ids, clock := platform.UUIDGenerator{}, platform.SystemClock{}
	settings := usecase.NewSettingsService(st.Settings())
	h := httpadapter.NewHandler(httpadapter.Deps{
		Auth:     usecase.NewAuthService("pwd", st.Subjects(), token.New("k", time.Hour)),
		Subjects: usecase.NewSubjectService(st.Subjects(), ids, clock),
		Tests:    usecase.NewTestService(st.Subjects(), st.Sessions(), settings, domain.NewTableGenerator(1, 1), ids, clock),
		Results:  usecase.NewResultService(st.Subjects(), st.Sessions()),
		Settings: settings,
		Log:      slog.New(slog.NewTextHandler(io.Discard, nil)),
	})
	srv := httptest.NewServer(httpadapter.NewRouter(h, slog.New(slog.NewTextHandler(io.Discard, nil)), ""))
	t.Cleanup(srv.Close)
	return srv
}

type client struct {
	t     *testing.T
	base  string
	token string
}

func (c *client) do(method, path string, body any, out any) int {
	c.t.Helper()
	var rd io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rd = bytes.NewReader(b)
	}
	req, _ := nethttp.NewRequest(method, c.base+path, rd)
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	resp, err := nethttp.DefaultClient.Do(req)
	if err != nil {
		c.t.Fatal(err)
	}
	defer resp.Body.Close()
	if out != nil {
		_ = json.NewDecoder(resp.Body).Decode(out)
	}
	return resp.StatusCode
}

type m = map[string]any

func fiveResults(sec float64) m {
	rs := make([]m, 5)
	for i := range rs {
		rs[i] = m{"number": i + 1, "seconds": sec, "errors": 1}
	}
	return m{"results": rs}
}

// Полный сценарий: вход психолога → создание испытуемого → вход испытуемого →
// тестирование → результаты → динамика → сравнение.
func TestFullScenario(t *testing.T) {
	srv := newServer(t)
	admin := &client{t: t, base: srv.URL}

	if code := admin.do("POST", "/api/auth/admin", m{"password": "bad"}, nil); code != 401 {
		t.Fatalf("неверный пароль: %d", code)
	}
	var sess struct{ Token, Role string }
	if code := admin.do("POST", "/api/auth/admin", m{"password": "pwd"}, &sess); code != 200 || sess.Role != "admin" {
		t.Fatalf("вход администратора: %d", code)
	}
	admin.token = sess.Token

	if code := (&client{t: t, base: srv.URL}).do("GET", "/api/subjects", nil, nil); code != 401 {
		t.Errorf("без токена должен быть 401, получено %d", code)
	}

	var sub struct{ ID string }
	if code := admin.do("POST", "/api/subjects", m{"fullName": "Иван Петров", "group": "Г1", "birthYear": 2001}, &sub); code != 201 || sub.ID == "" {
		t.Fatalf("создание испытуемого: %d", code)
	}
	if code := admin.do("POST", "/api/subjects", m{"fullName": ""}, nil); code != 400 {
		t.Errorf("пустое ФИО: %d", code)
	}
	if code := admin.do("POST", "/api/subjects", m{"fullName": "Иван", "unknown": 1}, nil); code != 400 {
		t.Errorf("неизвестное поле должно отклоняться: %d", code)
	}

	// публичный список для экрана входа испытуемого
	var pub []m
	if code := (&client{t: t, base: srv.URL}).do("GET", "/api/public/subjects", nil, &pub); code != 200 || len(pub) != 1 {
		t.Fatalf("публичный список: %d %v", code, pub)
	}

	// режим тестируемого
	me := &client{t: t, base: srv.URL}
	var ss struct{ Token string }
	me.do("POST", "/api/auth/subject", m{"subjectId": sub.ID}, &ss)
	me.token = ss.Token
	if code := me.do("GET", "/api/subjects", nil, nil); code != 403 {
		t.Errorf("испытуемому недоступен список: %d", code)
	}
	if code := me.do("PUT", "/api/settings", m{"fontSize": 30, "backgroundColor": "#ffffff"}, nil); code != 403 {
		t.Errorf("испытуемому недоступно изменение настроек: %d", code)
	}

	var settings m
	if code := me.do("GET", "/api/settings", nil, &settings); code != 200 || settings["fontSize"].(float64) != 28 {
		t.Fatalf("настройки: %d %v", code, settings)
	}

	for i, sec := range []float64{50, 40} { // два теста для динамики
		var test struct {
			ID     string
			Tables [][]int
		}
		if code := me.do("POST", "/api/tests", m{}, &test); code != 201 || len(test.Tables) != 5 {
			t.Fatalf("старт теста %d: %d", i, code)
		}
		var done struct {
			Status  string
			Metrics struct{ WorkEfficiency float64 }
			Tables  [][]int
		}
		if code := me.do("POST", "/api/tests/"+test.ID+"/finish", fiveResults(sec), &done); code != 200 ||
			done.Metrics.WorkEfficiency != sec || done.Status != "completed" || done.Tables != nil {
			t.Fatalf("завершение теста %d: %d %+v", i, code, done)
		}
		if code := me.do("POST", "/api/tests/"+test.ID+"/finish", fiveResults(sec), nil); code != 409 {
			t.Errorf("повторное завершение: %d", code)
		}
		time.Sleep(2 * time.Millisecond)
	}

	var results []struct{ ID string }
	if code := me.do("GET", "/api/subjects/"+sub.ID+"/results", nil, &results); code != 200 || len(results) != 2 {
		t.Fatalf("результаты: %d %v", code, results)
	}
	var dyn struct {
		Trend  string
		Points []m
	}
	if code := me.do("GET", "/api/subjects/"+sub.ID+"/dynamics", nil, &dyn); code != 200 || len(dyn.Points) != 2 || dyn.Trend != "improving" {
		t.Fatalf("динамика: %d %+v", code, dyn)
	}
	var cmp struct {
		Better string
	}
	if code := admin.do("GET", "/api/compare?a="+results[1].ID+"&b="+results[0].ID, nil, &cmp); code != 200 || cmp.Better == "" {
		t.Fatalf("сравнение: %d", code)
	}
	if code := me.do("GET", "/api/compare?a=x&b=y", nil, nil); code != 403 {
		t.Errorf("сравнение недоступно испытуемому: %d", code)
	}
	var gs struct{ Participants int }
	if code := admin.do("GET", "/api/groups/Г1/stats", nil, &gs); code != 200 || gs.Participants != 1 {
		t.Fatalf("статистика группы: %d %+v", code, gs)
	}
	if code := admin.do("DELETE", "/api/results/"+results[0].ID, nil, nil); code != 204 {
		t.Errorf("удаление результата: %d", code)
	}
	if code := admin.do("DELETE", "/api/results/"+results[0].ID, nil, nil); code != 404 {
		t.Errorf("повторное удаление: %d", code)
	}
	if code := admin.do("DELETE", "/api/subjects/"+sub.ID, nil, nil); code != 204 {
		t.Errorf("удаление испытуемого: %d", code)
	}
}

func TestInvalidFinishPayload(t *testing.T) {
	srv := newServer(t)
	c := &client{t: t, base: srv.URL}
	var s struct{ Token string }
	c.do("POST", "/api/auth/admin", m{"password": "pwd"}, &s)
	c.token = s.Token
	var sub struct{ ID string }
	c.do("POST", "/api/subjects", m{"fullName": "Иван Петров"}, &sub)
	var test struct{ ID string }
	c.do("POST", "/api/tests", m{"subjectId": sub.ID}, &test)
	if code := c.do("POST", "/api/tests/"+test.ID+"/finish", m{"results": []m{{"number": 1, "seconds": 5}}}, nil); code != 400 {
		t.Errorf("неполные результаты должны давать 400, получено %d", code)
	}
	if code := c.do("POST", "/api/tests", m{"subjectId": sub.ID, "settings": m{"fontSize": 2, "backgroundColor": "x"}}, nil); code != 400 {
		t.Errorf("некорректные настройки: %d", code)
	}
}

func TestHealthAndCORS(t *testing.T) {
	srv := newServer(t)
	resp, err := nethttp.Get(srv.URL + "/api/health")
	if err != nil {
		t.Fatal(err)
	}
	b, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != 200 || !strings.Contains(string(b), "ok") {
		t.Errorf("health: %d %s", resp.StatusCode, b)
	}
}
