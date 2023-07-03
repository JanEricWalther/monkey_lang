package object

import "testing"

func TestStringHashKey(t *testing.T) {
	hello1 := &String{Value: "Hello World"}
	hello2 := &String{Value: "Hello World"}
	diff1 := &String{Value: "My name is johnny"}
	diff2 := &String{Value: "My name is johnny"}

	if hello1.HashKey() != hello2.hashKey() {
		t.Errorf("strings with the same content have different hash keys")
	}

	if diff1.HashKey() != diff2.hashKey() {
		t.Errorf("strings with the same content have different hash keys")
	}

	if hello1.HashKey() == diff1.hashKey() {
		t.Errorf("strings with different content have different hash keys")
	}
}