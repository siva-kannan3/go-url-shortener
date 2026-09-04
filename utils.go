package main

import (
	"math/rand"
)

func GenerateRandomCharacter() byte {
	const characters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	index := rand.Intn(len(characters))

	return characters[index]
}

func GenerateId() string {
	var id string

	for i := 0; i < 6; i++ {
		id += string(GenerateRandomCharacter())
	}

	return id
}
