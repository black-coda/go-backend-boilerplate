package utils

import (
	"encoding/json"
	"fmt"
)

func PrintJson(v any) {
	json, err := json.MarshalIndent(v, "", " ")
	if err != nil {
		fmt.Println("Error marshalling to Json:", err)
		return
	}
	fmt.Println("Json:", string(json))
}
