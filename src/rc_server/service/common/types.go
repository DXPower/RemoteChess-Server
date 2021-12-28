package service

import (
	"bytes"
	"database/sql/driver"
	sv "remotechess/src/rc_server/service"
	"strings"
)

type PlayerColor int64

const (
	PLAYER_WHITE PlayerColor = iota
	PLAYER_BLACK
)

type NullablePlayerColor struct {
	PlayerColor
	Valid bool
}

func (this *PlayerColor) Other() PlayerColor {
	if *this == PLAYER_WHITE {
		return PLAYER_BLACK
	} else {
		return PLAYER_WHITE
	}
}

func (this *PlayerColor) Scan(value interface{}) error {
	b, ok := value.([]byte)

	if !ok {
		return sv.NewInternalError("Scan source is not []byte")
	}

	if bytes.Equal(b, []byte("WHITE")) {
		*this = PLAYER_WHITE
	} else if bytes.Equal(b, []byte("BLACK")) {
		*this = PLAYER_BLACK
	} else {
		return sv.NewInternalError("Invalid PlayerColor enum received: " + string(b))
	}

	return nil
}

func (this PlayerColor) Value() (driver.Value, error) {
	if this == PLAYER_WHITE {
		return "WHITE", nil
	} else if this == PLAYER_BLACK {
		return "BLACK", nil
	} else {
		return nil, sv.NewInternalError("Unknown PlayerColor")
	}
}

func (this PlayerColor) String() string {
	if this == PLAYER_WHITE {
		return "WHITE"
	} else {
		return "BLACK"
	}
}

func NewPlayerColor(color string) (PlayerColor, error) {
	if strings.ToUpper(color) == "WHITE" {
		return PLAYER_WHITE, nil
	} else if strings.ToUpper(color) == "BLACK" {
		return PLAYER_BLACK, nil
	} else {
		return PLAYER_WHITE, sv.NewInvalidInputError("PlayerColor " + color)
	}
}

func (this *NullablePlayerColor) Scan(value interface{}) error {
	if value == nil {
		this.Valid = false
		return nil
	}

	this.Valid = true
	return this.PlayerColor.Scan(value)
}

func (this NullablePlayerColor) Value() (driver.Value, error) {
	if !this.Valid {
		return nil, nil
	}

	return this.PlayerColor.Value()
}

func (this *NullablePlayerColor) ToPointer() *PlayerColor {
	if !this.Valid {
		return nil
	} else {
		return &this.PlayerColor
	}
}
