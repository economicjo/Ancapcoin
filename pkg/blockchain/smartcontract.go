package blockchain

import ( 
	"fmt"
)

type SmartContract struct {
	Code      string            
	State     map[string]string 
	Owner     string            
	Functions map[string]func(args ...interface{}) (interface{}, error) 
}


func NewSmartContract(code string, owner string) *SmartContract {
	return &SmartContract{
		Code:      code,
		State:     make(map[string]string),
		Owner:     owner,
		Functions: make(map[string]func(args ...interface{}) (interface{}, error)),
	}
}


func (sc *SmartContract) RegisterFunction(name string, fn func(args ...interface{}) (interface{}, error)) {
	sc.Functions[name] = fn
}


func (sc *SmartContract) ExecuteFunction(name string, args ...interface{}) (interface{}, error) {
	if fn, exists := sc.Functions[name]; exists {
		return fn(args...)
	}
	return nil, fmt.Errorf("function '%s' not found in contract", name)
}
