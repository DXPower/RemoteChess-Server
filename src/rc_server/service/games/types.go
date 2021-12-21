package games

import (
	"database/sql/driver"
	"reflect"

	sv "remotechess/src/rc_server/service"

	"github.com/notnil/chess"
)

type GameOutcome chess.Outcome

const (
	NO_OUTCOME GameOutcome = GameOutcome(chess.NoOutcome)
	WHITE_WON  GameOutcome = GameOutcome(chess.WhiteWon)
	BLACK_WON  GameOutcome = GameOutcome(chess.BlackWon)
	DRAW       GameOutcome = GameOutcome(chess.Draw)
)

type GameMethod chess.Method

const (
	NO_METHOD              GameMethod = GameMethod(chess.NoMethod)
	CHECKMATE              GameMethod = GameMethod(chess.Checkmate)
	RESIGNATION            GameMethod = GameMethod(chess.Resignation)
	DRAW_OFFER             GameMethod = GameMethod(chess.DrawOffer)
	STALEMATE              GameMethod = GameMethod(chess.Stalemate)
	THREEFOLD_REPETITION   GameMethod = GameMethod(chess.ThreefoldRepetition)
	FIVEFOLD_REPETITION    GameMethod = GameMethod(chess.FivefoldRepetition)
	FIFTY_MOVE_RULE        GameMethod = GameMethod(chess.FiftyMoveRule)
	SEVENTY_FIVE_MOVE_RULE GameMethod = GameMethod(chess.SeventyFiveMoveRule)
	INSUFFICIENT_MATERIAL  GameMethod = GameMethod(chess.InsufficientMaterial)
)

var (
	outcomeToStr = map[GameOutcome]string{
		NO_OUTCOME: "NONE",
		WHITE_WON:  "WHITE_WON",
		BLACK_WON:  "BLACK_WON",
		DRAW:       "DRAW",
	}
	strToOutcome = inverseMap(outcomeToStr).(map[string]GameOutcome)

	methodToStr = map[GameMethod]string{
		NO_METHOD:              "NONE",
		CHECKMATE:              "CHECKMATE",
		RESIGNATION:            "RESIGNATION",
		DRAW_OFFER:             "DRAW_AGREEMENT",
		STALEMATE:              "STALEMATE",
		THREEFOLD_REPETITION:   "THREEFOLD_REPETITION",
		FIVEFOLD_REPETITION:    "FIVEFOLD_REPETITION",
		FIFTY_MOVE_RULE:        "50_MOVES",
		SEVENTY_FIVE_MOVE_RULE: "75_MOVES",
		INSUFFICIENT_MATERIAL:  "INSUFFICIENT_MATERIAL",
	}

	strToMethod = inverseMap(methodToStr).(map[string]GameMethod)
)

func (o GameOutcome) ToStore() string {
	return outcomeToStr[o]
}

func GameOutcomeFromStore(s string) GameOutcome {
	return strToOutcome[s]
}

func (o GameMethod) String() string {
	return methodToStr[o]
}

func GameMethodFromString(s string) GameMethod {
	return strToMethod[s]
}

// Interface implementations for sql driver and GameOutcome
func (this *GameOutcome) Scan(value interface{}) error {
	b, ok := value.([]byte)

	if !ok {
		return sv.NewInternalError("Scan source is not []byte")
	}

	if val, ok := strToOutcome[string(b)]; ok {
		*this = val
	} else {
		return sv.NewInternalError("Invalid GameOutcome enum received: " + string(b))
	}

	return nil
}

func (this GameOutcome) Value() (driver.Value, error) {
	if val, ok := outcomeToStr[this]; ok {
		return val, nil
	} else {
		return nil, sv.NewInternalError("Unknown GameOutcome")
	}
}

// Interface implementations for sql driver and GameMethod
func (this *GameMethod) Scan(value interface{}) error {
	b, ok := value.([]byte)

	if !ok {
		return sv.NewInternalError("Scan source is not []byte")
	}

	if val, ok := strToMethod[string(b)]; ok {
		*this = val
	} else {
		return sv.NewInternalError("Invalid GameMethod enum received: " + string(b))
	}

	return nil
}

func (this GameMethod) Value() (driver.Value, error) {
	if val, ok := methodToStr[this]; ok {
		return val, nil
	} else {
		return nil, sv.NewInternalError("Unknown GameMethod")
	}
}

// Helper method for defining two inversed maps and keeping to DRY
func inverseMap(in interface{}) interface{} {
	v := reflect.ValueOf(in)
	if v.Kind() != reflect.Map {
		return nil
	}

	mapType := reflect.MapOf(v.Type().Elem(), v.Type().Key())
	out := reflect.MakeMap(mapType)

	for iter := v.MapRange(); iter.Next(); {
		out.SetMapIndex(iter.Value(), iter.Key())
	}

	return out.Interface()
}
