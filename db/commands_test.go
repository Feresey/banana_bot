package db

import (
	"database/sql"
	"os"
	"reflect"
	"testing"

	"github.com/Feresey/banana_bot/logging"
	"github.com/Feresey/banana_bot/model"
)

func TestMain(m *testing.M) {
	ddb, err := sql.Open("postgres", "user=lol dbname=test sslmode=disable")
	if err != nil {
		panic(err)
	}
	db = ddb
	log = logging.NewLogger("test")
	log.Quite = true
	defer db.Close()
	os.Exit(m.Run())
}

func TestWarn(t *testing.T) {
	_, _ = db.Exec("delete from " + subscriptions)

	type args struct {
		person *model.Person
		add    bool
	}
	tests := []struct {
		name      string
		args      args
		wantTotal int
		wantErr   bool
	}{
		{
			name:      "warn new",
			args:      args{person: &model.Person{ChatID: 1, UserID: 1}, add: true},
			wantTotal: 1,
		},
		{
			name:      "warn second",
			args:      args{person: &model.Person{ChatID: 1, UserID: 1}, add: true},
			wantTotal: 2,
		},
		{
			name:      "unwarn exist",
			args:      args{person: &model.Person{ChatID: 1, UserID: 1}, add: false},
			wantTotal: 1,
		},
		{
			name:      "unwarn exist twice",
			args:      args{person: &model.Person{ChatID: 1, UserID: 1}, add: false},
			wantTotal: 0,
		},
		{
			name:      "unwarn zero",
			args:      args{person: &model.Person{ChatID: 1, UserID: 1}, add: false},
			wantTotal: 0,
		},
		{
			name:      "unwarn new",
			args:      args{person: &model.Person{ChatID: 2, UserID: 2}, add: false},
			wantTotal: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTotal, err := Warn(tt.args.person, tt.args.add)
			if (err != nil) != tt.wantErr {
				t.Errorf("Warn() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotTotal != tt.wantTotal {
				t.Errorf("Warn() = %v, want %v", gotTotal, tt.wantTotal)
			}
		})
	}
}

func TestReport(t *testing.T) {
	_, _ = db.Exec("delete from " + subscriptions)
	_, _ = db.Exec(
		`INSERT INTO ` + subscriptions + ` (chatid, userid)
		VALUES (1,1),(1,2),(1,3),(1,5),
			   (2,1),(2,2)`)
	type args struct {
		chatID int64
	}
	tests := []struct {
		name    string
		args    args
		want    []int
		wantErr bool
	}{
		{
			name: "one",
			args: args{chatID: 1},
			want: []int{1, 2, 3, 5},
		},
		{
			name: "two",
			args: args{chatID: 2},
			want: []int{1, 2},
		},
		{
			name: "not exists",
			args: args{chatID: 3},
			want: []int{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Report(tt.args.chatID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Report() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Report() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSubs(t *testing.T) {
	_, _ = db.Exec(`delete from ` + subscriptions)

	tests := []struct {
		name    string
		p       *model.Person
		command func(*model.Person) error
		wantErr bool
	}{
		{
			name:    "create new",
			p:       &model.Person{ChatID: 1, UserID: 1},
			command: Subscribe,
		},
		{
			name:    "create exists",
			p:       &model.Person{ChatID: 1, UserID: 1},
			command: Subscribe,
			wantErr: true,
		},
		{
			name:    "unsub exists",
			p:       &model.Person{ChatID: 1, UserID: 1},
			command: UnSubscribe,
		},
		{
			name:    "unsub not exists",
			p:       &model.Person{ChatID: 1, UserID: 0},
			command: UnSubscribe,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.command(tt.p); (err != nil) != tt.wantErr {
				t.Errorf("Subscribe() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
