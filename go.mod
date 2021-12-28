module remotechess

go 1.17

require (
	github.com/go-chi/chi/v5 v5.0.7
	github.com/go-chi/render v1.0.1
	github.com/lib/pq v1.10.4
	github.com/notnil/chess v1.7.0
)

replace github.com/notnil/chess => ../chess