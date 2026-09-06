package service

import (
	"math/rand"
)

func generateRandomCharacter() byte {
	const characters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	index := rand.Intn(len(characters))

	return characters[index]
}

func generateID() string {
	var id string

	for i := 0; i < 6; i++ {
		id += string(generateRandomCharacter())
	}

	return id
}
