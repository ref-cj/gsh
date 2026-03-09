//go:build !debug
// +build !debug

package main

const prompt = "\033[38;5;213m$\033[0m "

func GetPS1() string { return prompt }
