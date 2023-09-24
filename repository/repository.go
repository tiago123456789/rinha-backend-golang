package repository

import (
	"database/sql"
	"encoding/json"

	"github.com/google/uuid"
)

type Pessoa struct {
	Id         string   `json:'id'`
	Apelido    string   `json:"apelido"`
	Nome       string   `json:"nome"`
	Nascimento string   `json:"nascimento"`
	Stack      []string `json:"stack"`
}

func GetAllByTerm(db *sql.DB, term string) ([]Pessoa, error) {
	term = "%" + term + "%"

	rows, err := db.Query(
		"SELECT id, apelido, nome, nascimento, stack FROM peoples where nome ilike $1 or apelido ilike $2 LIMIT $3 ",
		term, term, 50,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	peoples := []Pessoa{}
	for rows.Next() {
		people := Pessoa{}
		var stack string
		err = rows.Scan(
			&people.Id, &people.Apelido,
			&people.Nome, &people.Nascimento, &stack,
		)
		if err != nil {
			return nil, err
		}

		if stack != "" {
			stackParse := []string{}
			json.Unmarshal([]byte(stack), &stackParse)
			people.Stack = stackParse
		}

		peoples = append(peoples, people)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return peoples, nil
}

func Insert(db *sql.DB, p *Pessoa) (string, error) {
	var id string
	stack := ""
	if len(p.Stack) > 0 {
		stackToString, _ := json.Marshal(p.Stack)
		stack = string(stackToString)
	}

	err := db.QueryRow(
		"INSERT INTO peoples (id, apelido, nome, nascimento, stack)"+
			" VALUES ($1, $2, $3, $4, $5) RETURNING id ",
		uuid.NewString(), p.Apelido, p.Nome, p.Nascimento, stack,
	).Scan(&id)

	return id, err
}

func GetPeopleById(db *sql.DB, id string) (Pessoa, error) {
	p := Pessoa{}
	var stack string

	err := db.QueryRow(
		"SELECT id, apelido, nome, nascimento, stack FROM peoples WHERE id = $1", id,
	).Scan(&p.Id, &p.Apelido, &p.Nome, &p.Nascimento, &stack)

	if stack != "" {
		stackParse := []string{}
		json.Unmarshal([]byte(stack), &stackParse)
		p.Stack = stackParse
	}

	return p, err
}

func GetTotalPeoples(db *sql.DB) (int, error) {
	var total int
	err := db.QueryRow(
		"SELECT count(*) as total FROM peoples",
	).Scan(&total)

	return total, err
}
