package dao

import "time"

type Memory struct {
	EventName   string
	SessionTime time.Time
}

type Question struct {
	QuestionID string
	Question   string
	Options    []string
	Answer     string
}

type Exam struct {
	Questions            []*Question
	CurrentQuestionIndex int
	UserAnswers          map[int]string
	StartTime            time.Time
	EndTime              time.Time
}

func (e *Exam) GetNextQuestion() *Question {
	nextQuestionIndex := e.CurrentQuestionIndex + 1
	if nextQuestionIndex >= len(e.Questions) {
		return nil
	}
	question := e.Questions[nextQuestionIndex]
	e.CurrentQuestionIndex = nextQuestionIndex
	return question
}

func (e *Exam) SetUserAnswer(answer string) {
	e.UserAnswers[e.CurrentQuestionIndex] = answer
}
