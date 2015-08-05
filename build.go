package main

func buildMap(c chan paint) {
	defer close(c)
}
