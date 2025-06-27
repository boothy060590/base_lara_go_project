package core

type BaseDTO interface {
	GetID() uint
	FromModel(model interface{}) BaseDTO
}
