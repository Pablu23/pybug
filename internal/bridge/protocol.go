package bridge

import (
	"encoding/json"
	"maps"
	"math/rand"
)

type CommandType string

const (
	ContinueCommand CommandType = "continue"
	BreakCommand    CommandType = "break"
	LocalsCommand   CommandType = "locals"
	StepCommand     CommandType = "step"
)

func makeCommand(cmd CommandType, values map[string]any) (requestId string, request string) {
	m := make(map[string]any, len(values)+2)
	maps.Copy(m, values)

	requestId = randString(8)

	m["cmd"] = cmd
	m["request_id"] = requestId

	b, err := json.Marshal(m)
	if err != nil {
		panic("failed to marshal command " + err.Error())
	}

	return requestId, string(b)
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func randString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
