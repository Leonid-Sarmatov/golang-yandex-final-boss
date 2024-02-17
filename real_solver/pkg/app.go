package pkg

import (
	"strconv"
)

/*
App описывает структуру с вычислителями
*/
type App struct {
	Solvers []*Solver
}

/*
NewApp создает приложение с вычислителями

Parameters:

	string: Шаблон имени для вычислителя
	int: Количество вычислителей

Returns:

	*App: Указатель на приложение
*/
func NewApp(name string, n int) *App {
	app := &App{
		Solvers: make([]*Solver, n),
	}

	for i := 0; i < n; i += 1 {
		app.Solvers[i] = NewSolver(name+" "+strconv.Itoa(i))
	}

	return app
}

/*
AppRun запускает все вычислители приложения
*/
func (app *App) AppRun() {
	for _, solver := range app.Solvers {
		solver.RunHandShakeStream()
		solver.RunSolverStream()
	}
}